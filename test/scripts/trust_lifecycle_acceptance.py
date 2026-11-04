#!/usr/bin/env python3
"""Execute real Terraform Trust actions only on an externally attested disposable.

Run through run-story-31-0f-acceptance.sh. Runtime JSON and all outputs are private.
No embedded endpoint, credential, VM provisioner, or production fallback exists.
Guard/oracle helper executables must be mounted inside the private runtime root.
The oracle owns snapshot, independent observations, owned-only cleanup and exact
restore. Its contract is documented in the matching implementation evidence.
"""

import hashlib
import json
import os
import re
import signal
import subprocess
import tempfile
import uuid
from pathlib import Path


class AcceptanceError(Exception):
    """A sanitized acceptance gate failure."""


def require(condition, message):
    if not condition:
        raise AcceptanceError(message)


def sha256(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def load_runtime(path):
    require(path is not None, "new disposable runtime required; no API access")
    p = Path(path).resolve()
    require(p.is_file() and p.stat().st_mode & 0o077 == 0, "runtime must be private")
    value = json.loads(p.read_text())
    require(isinstance(value, dict), "runtime must be a JSON object")
    for key in ("owner", "guest_uuid", "endpoint", "tls_sha256", "guard", "oracle"):
        require(
            isinstance(value.get(key), str) and value[key], "runtime binding missing"
        )
    for key in ("owner", "guest_uuid"):
        try:
            uuid.UUID(value[key])
        except ValueError as exc:
            raise AcceptanceError("invalid identity binding") from exc
    require(value["endpoint"].startswith("https://"), "HTTPS endpoint required")
    require(re.fullmatch(r"[0-9a-f]{64}", value["tls_sha256"]), "TLS pin required")
    for key in ("guard", "oracle"):
        helper = Path(value[key]).resolve()
        require(
            helper.is_relative_to(p.parent), "helper must be in private runtime root"
        )
        require(
            helper.is_file() and os.access(helper, os.X_OK), "executable helper missing"
        )
        require(
            helper.stat().st_mode & 0o022 == 0, "helper must not be writable by others"
        )
        require(sha256(helper) == value.get(key + "_sha256"), "helper digest mismatch")
    value["root"] = str(p.parent)
    return value


COMMAND_TIMEOUT = 180


def owned_command(
    command,
    *,
    cwd=None,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    timeout=COMMAND_TIMEOUT,
):
    """Bound and reap the entire owned process group before restoration starts."""
    process = subprocess.Popen(
        command, cwd=cwd, stdout=stdout, stderr=stderr, start_new_session=True
    )
    try:
        out, err = process.communicate(timeout=timeout)
    except BaseException:
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
        try:
            process.communicate(timeout=2)
        except subprocess.TimeoutExpired:
            pass
        finally:
            # A parent can exit on TERM while a log-file-backed descendant
            # ignores it. Parent completion is not process-group completion.
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.communicate(timeout=2)
        raise
    return subprocess.CompletedProcess(command, process.returncode, out, err)


def helper(runtime, kind, action, directory):
    executable = runtime[kind]
    require(sha256(executable) == runtime[kind + "_sha256"], "helper changed")
    result = owned_command(
        [executable, action, str(directory), runtime["owner"]], timeout=120
    )
    require(result.returncode == 0, "private helper failed; inspect private runtime")
    try:
        value = json.loads(result.stdout)
    except (ValueError, UnicodeDecodeError) as exc:
        raise AcceptanceError("malformed attestation") from exc
    require(isinstance(value, dict), "attestation must be a JSON object")
    require(value.get("ok") is True, "attestation failed")
    require(value.get("owner") == runtime["owner"], "reservation owner mismatch")
    require(value.get("guest_uuid") == runtime["guest_uuid"], "guest identity mismatch")
    require(value.get("endpoint") == runtime["endpoint"], "endpoint binding mismatch")
    require(value.get("tls_sha256") == runtime["tls_sha256"], "TLS binding mismatch")
    require(value.get("isolated") is True, "isolation attestation missing")
    require(value.get("non_production") is True, "non-production attestation missing")
    require(value.get("reservation_current") is True, "reservation expired")
    require(
        value.get("core_version") == "26.7.8_2", "exact acceptance firmware required"
    )
    (directory / (action + ".json")).write_text(json.dumps(value))
    return value


def run_private(
    command, directory, label, expected=(0,), timeout=COMMAND_TIMEOUT, cwd=None
):
    with (directory / (label + ".log")).open("wb") as log:
        result = owned_command(
            command, cwd=cwd or directory, stdout=log, stderr=log, timeout=timeout
        )
    require(
        result.returncode in expected, "command failed; inspect private acceptance logs"
    )
    return result.returncode


def hcl_string(value):
    return json.dumps(value)


def resources(
    marker, cert_pem, key_pem, csr=None, signed=None, reissue=False, token=None
):
    # Persisted fixtures are confined to the private attempt directory.
    text = """terraform {
  required_providers {
    opnsense = { source = "matthew-on-git/opnsense" }
  }
}
provider "opnsense" {}
"""
    for name, action, parent in (
        ("root", "internal", ""),
        ("intermediate", "internal", "opnsense_trust_ca.root.refid"),
        ("ocsp", "ocsp", "opnsense_trust_ca.root.refid"),
    ):
        text += '\nresource "opnsense_trust_ca" "' + name + '" {\n'
        text += "  description = " + hcl_string(marker + "-" + name) + "\n"
        text += "  action = " + hcl_string(action) + '\n  country = "US"\n'
        text += "  common_name = " + hcl_string(name + ".example.invalid") + "\n"
        text += "  lifetime = 30\n"
        if parent:
            text += "  ca_ref = " + parent + "\n"
        text += "}\n"
    text += '\nresource "opnsense_trust_ca" "imported" {\n'
    text += (
        " description = "
        + hcl_string(marker + "-ca-import")
        + '\n action = "existing"\n'
    )
    text += (
        " certificate = "
        + hcl_string(cert_pem)
        + "\n private_key = "
        + hcl_string(key_pem)
        + "\n}\n"
    )
    for name, action, location in (
        ("internal", "internal", "firewall"),
        ("local", "internal", "local"),
        ("external", "external", "firewall"),
        ("imported", "import", "firewall"),
    ):
        active_action = "reissue" if name == "internal" and reissue else action
        if name == "external" and signed:
            active_action = "import_csr"
        text += '\nresource "opnsense_trust_cert" "' + name + '" {\n'
        text += " description = " + hcl_string(marker + "-" + name) + "\n"
        text += " action = " + hcl_string(active_action) + "\n"
        if name == "internal" and token:
            text += " operation_token = " + hcl_string(token) + "\n"
        if name == "external" and signed:
            text += " certificate = " + hcl_string(signed) + "\n"
        if name != "imported":
            text += (
                " common_name = "
                + hcl_string(name + ".example.invalid")
                + '\n country = "US"\n lifetime = 30\n'
            )
            text += " private_key_location = " + hcl_string(location) + "\n"
            text += ' altnames_dns = "one.example.invalid\\ntwo.example.invalid"\n'
            if name != "external":
                text += " ca_ref = opnsense_trust_ca.root.refid\n"
        else:
            text += (
                " certificate = "
                + hcl_string(cert_pem)
                + "\n private_key = "
                + hcl_string(key_pem)
                + "\n"
            )
        text += "}\n"
    if csr:
        text += '\nresource "opnsense_trust_cert" "signed" {\n'
        text += (
            " description = "
            + hcl_string(marker + "-signed")
            + '\n action = "sign_csr"\n'
        )
        text += (
            " csr = "
            + hcl_string(csr)
            + "\n ca_ref = opnsense_trust_ca.root.refid\n lifetime = 30\n}\n"
        )
    text += """
data "opnsense_trust_cert" "observed" { id = opnsense_trust_cert.internal.id }
data "opnsense_trust_cert" "csr_holder" { id = opnsense_trust_cert.external.id }
data "opnsense_trust_ca" "observed" { id = opnsense_trust_ca.root.id }
data "opnsense_trust_ca" "legacy" { id = opnsense_trust_ca.imported.id }
data "opnsense_trust_cert" "legacy" { id = opnsense_trust_cert.imported.id }
data "opnsense_trust_settings" "observed" {
  depends_on = [opnsense_trust_settings.observed]
}
resource "opnsense_trust_settings" "observed" {}
"""
    return text


def state_values(directory):
    response = owned_command(["terraform", "show", "-json"], cwd=directory)
    require(response.returncode == 0, "private state observation failed")
    output = response.stdout
    (directory / "state-observation.json").write_bytes(output)
    result = json.loads(output)
    return {
        r["address"]: r["values"]
        for r in result["values"]["root_module"].get("resources", [])
    }


def verify_signed_key(directory, csr, certificate):
    require(csr and certificate, "signed result material missing")
    (directory / "holder.csr").write_text(csr)
    (directory / "signed.crt").write_text(certificate)
    csr_key = owned_command(
        ["openssl", "req", "-in", "holder.csr", "-pubkey", "-noout"], cwd=directory
    )
    cert_key = owned_command(
        ["openssl", "x509", "-in", "signed.crt", "-pubkey", "-noout"], cwd=directory
    )
    require(
        csr_key.returncode == 0 and cert_key.returncode == 0,
        "cannot inspect signed public keys",
    )
    require(
        csr_key.stdout == cert_key.stdout and csr_key.stdout,
        "signed result does not match holder CSR",
    )


def verify_private_key(directory, key, certificate):
    require(key and certificate, "one-time key/certificate missing")
    (directory / "local.key").write_text(key)
    (directory / "local.crt").write_text(certificate)
    private = owned_command(
        ["openssl", "pkey", "-in", "local.key", "-pubout"], cwd=directory
    )
    public = owned_command(
        ["openssl", "x509", "-in", "local.crt", "-pubkey", "-noout"], cwd=directory
    )
    require(
        private.returncode == 0 and public.returncode == 0,
        "one-time key/certificate cannot be parsed",
    )
    require(
        private.stdout and private.stdout == public.stdout,
        "one-time key does not match certificate",
    )


def verify_reissue(before, after):
    require(before["id"] == after["id"], "reissue replaced certificate UUID")
    require(
        before.get("returned_private_key")
        and before["returned_private_key"] == after.get("returned_private_key"),
        "reissue changed or lost stored key",
    )
    require(
        before.get("returned_certificate")
        and after.get("returned_certificate")
        and before["returned_certificate"] != after["returned_certificate"],
        "reissue did not change certificate",
    )


def verify_adoption(runtime, directory, observed):
    helper(runtime, "oracle", "before-import", directory)
    adoption = directory / "adoption"
    adoption.mkdir()
    text = 'terraform {\n required_providers {\n opnsense = { source = "matthew-on-git/opnsense" }\n }\n}\nprovider "opnsense" {}\n'
    for kind, name in (("ca", "root"), ("cert", "internal")):
        record = observed["opnsense_trust_" + kind + "." + name]
        text += (
            'resource "opnsense_trust_'
            + kind
            + '" "adopted" {\n description = '
            + hcl_string(record["description"])
            + "\n}\n"
        )
    (adoption / "main.tf").write_text(text)
    try:
        for kind, name in (("ca", "root"), ("cert", "internal")):
            record = observed["opnsense_trust_" + kind + "." + name]
            run_private(
                [
                    "terraform",
                    "import",
                    "-input=false",
                    "opnsense_trust_" + kind + ".adopted",
                    record["id"],
                ],
                adoption,
                kind + "-import",
            )
        run_private(
            ["terraform", "plan", "-input=false", "-detailed-exitcode"],
            adoption,
            "adoption-no-change",
        )
        adopted = state_values(adoption)
        for kind, name in (("ca", "root"), ("cert", "internal")):
            require(
                adopted["opnsense_trust_" + kind + ".adopted"]["id"]
                == observed["opnsense_trust_" + kind + "." + name]["id"],
                "adoption changed UUID",
            )
        receipt = helper(runtime, "oracle", "verify-import", directory)
        require(
            receipt.get("cryptographic_writes") == 0
            and receipt.get("config_equal") is True,
            "UUID adoption caused cryptographic writes or config drift",
        )
    finally:
        # This second workspace is observational adoption proof. The main
        # workspace remains the owner: never destroy its objects from here.
        run_private(
            [
                "terraform",
                "state",
                "rm",
                "opnsense_trust_ca.adopted",
                "opnsense_trust_cert.adopted",
            ],
            adoption,
            "adoption-state-only-removal",
        )


def exercise_settings(runtime, directory, config, observed, apply, baseline):
    initial = observed["opnsense_trust_settings.observed"]
    choice = baseline.get("cipher_choice")
    require(
        isinstance(choice, str) and choice and "," not in choice,
        "oracle must supply an authoritative available CipherString choice",
    )
    original = config.read_text()
    marker = 'resource "opnsense_trust_settings" "observed" {}'
    require(marker in original, "settings fixture marker missing")

    def set_intent(values, label):
        args = "\n".join(
            key + " = " + json.dumps(value) for key, value in values.items()
        )
        config.write_text(
            original.replace(
                marker,
                'resource "opnsense_trust_settings" "observed" {\n' + args + "\n}",
            )
        )
        apply(label)
        truth = state_values(directory)
        managed, lookup = (
            truth["opnsense_trust_settings.observed"],
            truth["data.opnsense_trust_settings.observed"],
        )
        for field in (
            "store_intermediate_certs",
            "install_crls",
            "fetch_crls",
            "enable_legacy_sect",
            "enable_config_constraints",
            "cipher_string",
        ):
            require(
                managed[field] == lookup[field], "settings lookup differs from owner"
            )
        for field, expected in values.items():
            require(
                managed[field] == expected
                or (field == "cipher_string" and set(managed[field]) == set(expected)),
                "settings requested value did not round-trip",
            )

    for field in (
        "store_intermediate_certs",
        "install_crls",
        "fetch_crls",
        "enable_legacy_sect",
    ):
        set_intent({field: True}, field + "-true")
        set_intent({field: False}, field + "-false")
        set_intent({field: initial[field]}, field + "-restore")
    # Released General.php requires at least one nonempty constraint when
    # enabling constraints. Preserve the actual supported oracle selection;
    # this is a fixture prerequisite, not a relaxation of firmware validation.
    set_intent(
        {"enable_config_constraints": True, "cipher_string": [choice]},
        "enable_config_constraints-true",
    )
    set_intent(
        {"enable_config_constraints": False, "cipher_string": [choice]},
        "enable_config_constraints-false",
    )
    set_intent(
        {
            "enable_config_constraints": initial["enable_config_constraints"],
            "cipher_string": initial["cipher_string"],
        },
        "enable_config_constraints-restore",
    )
    set_intent({"cipher_string": [choice]}, "cipher-selection")
    # Clearing the last configured constraint must disable enforcement in the
    # same update, including when the original baseline had it enabled.
    set_intent(
        {"enable_config_constraints": False, "cipher_string": []},
        "cipher-empty-clear",
    )
    set_intent(
        {
            "enable_config_constraints": initial["enable_config_constraints"],
            "cipher_string": initial["cipher_string"],
        },
        "cipher-restore",
    )
    config.write_text(original)
    apply("settings-omitted-no-change")


def source_digest():
    response = owned_command(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        timeout=30,
    )
    require(response.returncode == 0, "source inventory failed")
    names = response.stdout.split(b"\0")
    digest = hashlib.sha256()
    for name in sorted(set(names) - {b""}):
        path = Path(os.fsdecode(name))
        if path.is_file():
            digest.update(name + b"\0" + path.read_bytes() + b"\0")
    return digest.hexdigest()


def main():
    os.umask(0o077)
    runtime = load_runtime(os.environ.get("OPNSENSE_TRUST_RUNTIME"))
    require(
        os.environ.get("OPNSENSE_URI") == runtime["endpoint"],
        "configured endpoint mismatch",
    )
    require(
        os.environ.get("OPNSENSE_API_KEY") and os.environ.get("OPNSENSE_API_SECRET"),
        "ephemeral credentials required",
    )
    require(
        os.environ.get("OPNSENSE_ALLOW_INSECURE", "false").lower() != "true",
        "verified TLS required",
    )
    candidate_digest = source_digest()
    require(
        candidate_digest == runtime.get("source_sha256"),
        "reviewed source digest mismatch",
    )
    directory = Path(tempfile.mkdtemp(prefix="trust-attempt-", dir=runtime["root"]))
    (directory / "source-sha256").write_text(candidate_digest)
    private_runtime_digest = sha256(os.environ["OPNSENSE_TRUST_RUNTIME"])
    helper(runtime, "guard", "guard", directory)
    baseline = helper(runtime, "oracle", "snapshot", directory)
    snapshots = ("config_sha256", "settings_sha256", "routes_sha256", "runtime_sha256")
    for key in snapshots:
        require(
            re.fullmatch(r"[0-9a-f]{64}", baseline.get(key, "")),
            "baseline hash missing",
        )
    marker = "tf-trust-" + runtime["owner"]
    (directory / "marker").write_text(marker)
    cleanup_errors = []

    def canceled(_number, _frame):
        raise AcceptanceError("acceptance interrupted; cleanup required")

    old_signals = {
        s: signal.signal(s, canceled) for s in (signal.SIGTERM, signal.SIGINT)
    }
    try:
        binary_dir = directory / "provider"
        binary_dir.mkdir()
        run_private(
            ["go", "build", "-o", str(binary_dir / "terraform-provider-opnsense"), "."],
            directory,
            "provider-build",
            cwd=Path(__file__).resolve().parents[2],
        )
        cli = directory / "terraform.rc"
        cli.write_text(
            'provider_installation {\n  dev_overrides {\n    "matthew-on-git/opnsense" = '
            + hcl_string(str(binary_dir))
            + "\n  }\n  direct {}\n}\n"
        )
        os.environ["TF_CLI_CONFIG_FILE"] = str(cli)
        os.environ["TF_IN_AUTOMATION"] = "1"
        os.environ["CHECKPOINT_DISABLE"] = "1"
        # A real parseable pre-epoch certificate is a supported legacy import;
        # its signed Unix validity observations must survive resource/lookup reads.
        # 1969 avoids PHP's two-digit-year pivot (65/66 become2065/2066).
        (directory / "fixture-index").write_text("")
        (directory / "fixture-serial").write_text("01\n")
        (directory / "fixture.cnf").write_text(
            "[ca]\ndefault_ca=fixture\n[fixture]\n"
            "database=fixture-index\nserial=fixture-serial\nnew_certs_dir=.\n"
            "default_md=sha256\npolicy=subject_policy\nx509_extensions=ca_ext\n"
            "[subject_policy]\ncommonName=supplied\n"
            "[ca_ext]\nbasicConstraints=critical,CA:TRUE\n"
            "keyUsage=critical,keyCertSign,cRLSign,digitalSignature\n"
        )
        run_private(
            ["openssl", "req", "-new", "-newkey", "rsa:2048", "-nodes",
             "-keyout", "import.key", "-out", "import.csr", "-subj",
             "/CN=import.example.invalid"], directory, "fixture-request",
        )
        run_private(
            ["openssl", "ca", "-selfsign", "-batch", "-config", "fixture.cnf",
             "-keyfile", "import.key", "-in", "import.csr", "-out", "import.crt",
             "-startdate", "19690101000000Z", "-enddate", "19691231000000Z",
             "-notext"], directory, "fixture-sign",
        )
        cert_pem = (directory / "import.crt").read_text()
        key_pem = (directory / "import.key").read_text()
        config = directory / "main.tf"
        config.write_text(resources(marker, cert_pem, key_pem))
        # Import-first singleton: creation is forbidden and import is read-only.
        run_private(
            [
                "terraform",
                "import",
                "-input=false",
                "opnsense_trust_settings.observed",
                "trust_settings",
            ],
            directory,
            "settings-import",
        )

        def apply(label):
            helper(runtime, "guard", "guard", directory)
            run_private(
                ["terraform", "apply", "-input=false", "-auto-approve"],
                directory,
                label,
            )
            helper(runtime, "oracle", "observe", directory)
            run_private(
                ["terraform", "plan", "-input=false", "-detailed-exitcode"],
                directory,
                label + "-no-change",
            )

        apply("initial-actions")
        observed = state_values(directory)
        for address in (
            "opnsense_trust_ca.imported", "opnsense_trust_cert.imported",
            "data.opnsense_trust_ca.legacy", "data.opnsense_trust_cert.legacy",
        ):
            require(
                int(observed[address]["valid_from"]) == -31536000
                and int(observed[address]["valid_to"]) == -86400,
                "legacy certificate signed timestamps not preserved",
            )
        local_key = observed["opnsense_trust_cert.local"].get("generated_private_key")
        verify_private_key(
            directory,
            local_key,
            observed["opnsense_trust_cert.local"].get("returned_certificate"),
        )
        require(
            not observed["opnsense_trust_cert.local"].get("returned_private_key"),
            "one-time key persisted on appliance",
        )
        external_before = observed["opnsense_trust_cert.external"]
        external_id = external_before["id"]
        external_key = external_before.get("returned_private_key")
        require(external_key, "CSR holder stored key missing")
        csr = external_before.get("returned_csr")
        require(csr and "CERTIFICATE REQUEST" in csr, "CSR missing")
        require(
            observed["data.opnsense_trust_cert.csr_holder"].get("csr") == csr,
            "certificate lookup lost CSR",
        )
        root_certificate = observed["opnsense_trust_ca.root"].get("certificate", "")
        require("CERTIFICATE" in root_certificate, "CA public certificate missing")
        fingerprint = sha256_bytes(root_certificate)
        run_private(
            ["terraform", "plan", "-input=false", "-detailed-exitcode"],
            directory,
            "unchanged-plan",
        )
        config.write_text(resources(marker, cert_pem, key_pem, csr=csr))
        apply("sign-csr")
        signed = state_values(directory)["opnsense_trust_cert.signed"].get(
            "returned_certificate"
        )
        require(signed and "CERTIFICATE" in signed, "signed certificate missing")
        config.write_text(resources(marker, cert_pem, key_pem, csr=csr, signed=signed))
        apply("import-signed")
        holder = state_values(directory)["opnsense_trust_cert.external"]
        require(holder["id"] == external_id, "signed import replaced CSR holder UUID")
        require(
            holder.get("returned_private_key") == external_key,
            "signed import lost holder key",
        )
        verify_signed_key(directory, csr, holder.get("returned_certificate", ""))
        config.write_text(
            resources(marker, cert_pem, key_pem, csr=csr, signed=signed, reissue=True)
        )
        before_reissue = state_values(directory)["opnsense_trust_cert.internal"]
        helper(runtime, "oracle", "before-operation", directory)
        apply("reissue")
        after_reissue = state_values(directory)["opnsense_trust_cert.internal"]
        verify_reissue(before_reissue, after_reissue)
        require(
            helper(runtime, "oracle", "verify-operation", directory).get(
                "cryptographic_writes"
            )
            == 1,
            "reissue was not exactly one cryptographic write",
        )
        config.write_text(
            resources(
                marker,
                cert_pem,
                key_pem,
                csr=csr,
                signed=signed,
                reissue=True,
                token="repeat-1",
            )
        )
        helper(runtime, "oracle", "before-operation", directory)
        apply("same-action-new-token")
        repeated = state_values(directory)["opnsense_trust_cert.internal"]
        verify_reissue(after_reissue, repeated)
        require(
            helper(runtime, "oracle", "verify-operation", directory).get(
                "cryptographic_writes"
            )
            == 1,
            "new token was not exactly one cryptographic write",
        )
        helper(runtime, "oracle", "before-operation", directory)
        apply("same-token-no-replay")
        require(
            helper(runtime, "oracle", "verify-operation", directory).get(
                "cryptographic_writes"
            )
            == 0,
            "unchanged token replayed operation",
        )
        cert_fingerprint = sha256_bytes(
            state_values(directory)["opnsense_trust_cert.internal"].get(
                "returned_certificate", ""
            )
        )
        config.write_text(
            config.read_text()
            .replace(marker + "-root", marker + "-root-renamed")
            .replace(marker + "-internal", marker + "-internal-renamed")
        )
        apply("metadata-only")
        refreshed = state_values(directory)
        require(
            fingerprint
            == sha256_bytes(refreshed["opnsense_trust_ca.root"].get("certificate", "")),
            "metadata update regenerated CA",
        )
        require(
            refreshed["opnsense_trust_cert.local"].get("generated_private_key")
            == local_key,
            "one-time key lost on refresh",
        )
        require(
            cert_fingerprint
            == sha256_bytes(
                refreshed["opnsense_trust_cert.internal"].get(
                    "returned_certificate", ""
                )
            ),
            "metadata update reissued certificate",
        )
        verify_adoption(runtime, directory, refreshed)
        exercise_settings(runtime, directory, config, refreshed, apply, baseline)
        before_destroy = helper(runtime, "oracle", "before-destroy", directory)
        require(
            before_destroy.get("owned_count", 0) > 0,
            "destroy proof requires owned objects",
        )
        run_private(
            ["terraform", "destroy", "-input=false", "-auto-approve"],
            directory,
            "terraform-destroy",
        )
        receipt = helper(runtime, "oracle", "verify-zero", directory)
        require(
            receipt.get("owned_count") == 0, "owned resources remain before restoration"
        )
        require(
            receipt.get("settings_destroy_equal") is True,
            "state-only settings destroy changed appliance policy",
        )
        (directory / "core-passed").write_text("true\n")
    finally:
        # Attempt every cleanup stage; restoration cannot conceal failed Core
        # deletion or turn an interrupted/failed run into successful proof.
        for action in ("cleanup", "verify-zero", "restore", "verify-restored"):
            try:
                value = helper(runtime, "oracle", action, directory)
                if action == "verify-zero":
                    require(value.get("owned_count") == 0, "owned resources remain")
                if action == "verify-restored":
                    require(
                        value.get("exact_restored") is True, "restore proof missing"
                    )
                    for key in snapshots:
                        require(
                            value.get(key) == baseline[key], "baseline not restored"
                        )
            except (AcceptanceError, OSError, ValueError, subprocess.SubprocessError):
                cleanup_errors.append(action)
        for s, old in old_signals.items():
            signal.signal(s, old)
        require(
            not cleanup_errors, "cleanup/restoration failed; private evidence retained"
        )
    require(source_digest() == candidate_digest, "source changed during acceptance")
    require(
        sha256(os.environ["OPNSENSE_TRUST_RUNTIME"]) == private_runtime_digest,
        "runtime inputs changed",
    )
    require((directory / "core-passed").exists(), "Core proof incomplete")
    print(
        "TRUST_LIVE_MATRIX core=true actions=true one_time_key=true metadata_stable=true settings_matrix=true adoption=true csr_transition=true terraform_destroy=true exact_restore=true"
    )


def sha256_bytes(value):
    return hashlib.sha256(value.encode()).hexdigest()


if __name__ == "__main__":
    try:
        main()
    except (AcceptanceError, OSError, ValueError, subprocess.SubprocessError):
        raise SystemExit(
            "Trust acceptance failed; inspect private attempt evidence. No live proof claimed."
        ) from None
