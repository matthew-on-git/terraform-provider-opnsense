"""Offline negative controls for the disposable Trust acceptance boundary."""

import contextlib
import importlib.util
import io
import itertools
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import tempfile
import time
import unittest
from pathlib import Path
from unittest import mock

SPEC = importlib.util.spec_from_file_location(
    "trust_acceptance", Path(__file__).with_name("trust_lifecycle_acceptance.py")
)
acceptance = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(acceptance)


class TrustAcceptanceGuardTest(unittest.TestCase):
    def test_reissue_requires_new_certificate_and_preserved_identity_key(self):
        before = {
            "id": "certificate-id",
            "returned_private_key": "private-key",
            "returned_certificate": "certificate-before",
        }
        after = {**before, "returned_certificate": "certificate-after"}
        acceptance.verify_reissue(before, after)
        for bad in (
            before,
            {**after, "id": "other-id"},
            {**after, "returned_private_key": "other-key"},
            {**after, "returned_private_key": None},
            {**after, "returned_certificate": None},
        ):
            with self.subTest(bad=bad), self.assertRaises(acceptance.AcceptanceError):
                acceptance.verify_reissue(before, bad)

    def test_one_time_private_key_is_parseable_and_matches_certificate(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for label in ("first", "other"):
                result = subprocess.run(
                    [
                        "openssl",
                        "req",
                        "-x509",
                        "-newkey",
                        "rsa:2048",
                        "-nodes",
                        "-keyout",
                        str(root / (label + ".key")),
                        "-out",
                        str(root / (label + ".crt")),
                        "-days",
                        "1",
                        "-subj",
                        "/CN=example.invalid",
                    ],
                    capture_output=True,
                )
                self.assertEqual(result.returncode, 0, "fixture generation failed")
            key, certificate = (
                (root / "first.key").read_text(),
                (root / "first.crt").read_text(),
            )
            acceptance.verify_private_key(root, key, certificate)
            for bad in ("PRIVATE KEY but not PEM", (root / "other.key").read_text()):
                with self.assertRaises(acceptance.AcceptanceError):
                    acceptance.verify_private_key(root, bad, certificate)

    def test_missing_runtime_performs_no_command(self):
        with (
            mock.patch.dict(os.environ, {}, clear=True),
            mock.patch.object(acceptance, "owned_command") as command,
        ):
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.main()
            command.assert_not_called()

    def test_helper_identity_mismatch_fails(self):
        with tempfile.TemporaryDirectory() as directory:
            helper = Path(directory) / "guard"
            helper.write_text("#!/bin/sh\nexit 1\n")
            helper.chmod(0o700)
            runtime = {
                "guard": str(helper),
                "guard_sha256": acceptance.sha256(helper),
                "owner": "00000000-0000-4000-8000-000000000001",
                "guest_uuid": "00000000-0000-4000-8000-000000000002",
                "endpoint": "https://disposable.example.invalid",
                "tls_sha256": "a" * 64,
            }
            response = mock.Mock(
                returncode=0, stdout=json.dumps({"ok": True, "owner": "wrong"}).encode()
            )
            with (
                mock.patch.object(acceptance, "owned_command", return_value=response),
                self.assertRaisesRegex(acceptance.AcceptanceError, "owner mismatch"),
            ):
                acceptance.helper(runtime, "guard", "guard", Path(directory))

    def test_rejects_every_unsafe_attestation(self):
        with tempfile.TemporaryDirectory() as directory:
            executable = Path(directory) / "guard"
            executable.write_text("#!/bin/sh\nexit 0\n")
            executable.chmod(0o700)
            runtime = {
                "guard": str(executable),
                "guard_sha256": acceptance.sha256(executable),
                "owner": "00000000-0000-4000-8000-000000000001",
                "guest_uuid": "00000000-0000-4000-8000-000000000002",
                "endpoint": "https://disposable.example.invalid",
                "tls_sha256": "a" * 64,
            }
            safe = {
                key: runtime[key]
                for key in ("owner", "guest_uuid", "endpoint", "tls_sha256")
            }
            safe.update(
                ok=True,
                isolated=True,
                non_production=True,
                reservation_current=True,
                core_version="26.7.8_2",
            )
            for key, unsafe in (
                ("ok", False),
                ("isolated", False),
                ("non_production", False),
                ("reservation_current", False),
                ("core_version", "26.7.7"),
                ("guest_uuid", "wrong"),
                ("endpoint", "https://other.example.invalid"),
                ("tls_sha256", "b" * 64),
            ):
                with self.subTest(gate=key):
                    reply = {**safe, key: unsafe}
                    response = mock.Mock(
                        returncode=0, stdout=json.dumps(reply).encode()
                    )
                    with (
                        mock.patch.object(
                            acceptance, "owned_command", return_value=response
                        ),
                        self.assertRaises(acceptance.AcceptanceError),
                    ):
                        acceptance.helper(runtime, "guard", "guard", Path(directory))

    def test_changed_helper_never_executes(self):
        with tempfile.TemporaryDirectory() as directory:
            helper = Path(directory) / "guard"
            helper.write_text("#!/bin/sh\nexit 1\n")
            runtime = {"guard": str(helper), "guard_sha256": "0" * 64}
            with mock.patch.object(acceptance, "owned_command") as command:
                with self.assertRaisesRegex(
                    acceptance.AcceptanceError, "helper changed"
                ):
                    acceptance.helper(runtime, "guard", "guard", Path(directory))
                command.assert_not_called()

    def test_unsafe_permissions_never_execute(self):
        with tempfile.TemporaryDirectory() as directory:
            runtime = Path(directory) / "runtime.json"
            runtime.write_text("{}")
            runtime.chmod(0o644)
            with mock.patch.object(acceptance, "owned_command") as command:
                with self.assertRaisesRegex(acceptance.AcceptanceError, "private"):
                    acceptance.load_runtime(str(runtime))
                command.assert_not_called()

    def test_nonobject_attestation_is_sanitized(self):
        with tempfile.TemporaryDirectory() as directory:
            executable = Path(directory) / "guard"
            executable.write_text("#!/bin/sh\nexit 0\n")
            executable.chmod(0o700)
            runtime = {
                "guard": str(executable),
                "guard_sha256": acceptance.sha256(executable),
                "owner": "00000000-0000-4000-8000-000000000001",
            }
            for raw in (b"null", b"[]", b"1", b'"scalar"'):
                response = mock.Mock(returncode=0, stdout=raw)
                with (
                    self.subTest(raw=raw),
                    mock.patch.object(
                        acceptance, "owned_command", return_value=response
                    ),
                    self.assertRaisesRegex(acceptance.AcceptanceError, "JSON object"),
                ):
                    acceptance.helper(runtime, "guard", "guard", Path(directory))

    def test_owned_command_timeout_reaps_process_group(self):
        for redirected in (False, True):
            with (
                self.subTest(redirected=redirected),
                tempfile.TemporaryDirectory() as directory,
            ):
                child = Path(directory) / "child.pid"
                script = "import subprocess,time,pathlib,sys; p=subprocess.Popen([sys.executable,'-c','import signal,time;signal.signal(signal.SIGTERM,signal.SIG_IGN);time.sleep(30)']);pathlib.Path(sys.argv[1]).write_text(str(p.pid));time.sleep(30)"
                started = time.monotonic()
                with (
                    open(Path(directory) / "output", "wb") as log,
                    self.assertRaises(subprocess.TimeoutExpired),
                ):
                    acceptance.owned_command(
                        [sys.executable, "-c", script, str(child)],
                        timeout=0.5,
                        stdout=log if redirected else subprocess.PIPE,
                        stderr=log if redirected else subprocess.PIPE,
                    )
                self.assertLess(time.monotonic() - started, 5)
                self.assertTrue(child.exists(), "control did not launch its child")
                # Killed grandchildren may briefly be init-owned zombies.
                for _ in range(100):
                    try:
                        status = (
                            Path("/proc/" + child.read_text() + "/stat")
                            .read_text()
                            .rsplit(") ", 1)[1]
                            .split()[0]
                        )
                    except FileNotFoundError:
                        status = ""
                    if not status or status == "Z":
                        break
                    time.sleep(0.01)
                self.assertTrue(
                    not status or status == "Z", "owned descendant remains live"
                )

    def synthetic_runtime(self, root):
        helper = root / "helper"
        helper.write_text("""#!/usr/bin/env python3
import json,os,pathlib,sys
root=pathlib.Path(os.environ['TRUST_TEST_ROOT']); action=sys.argv[1]
with (root/'trace').open('a') as out: out.write(action+'\\n')
if action==os.environ.get('TRUST_TEST_FAIL_PHASE') and (root/'trace').read_text().splitlines().count(action)==int(os.environ.get('TRUST_TEST_FAIL_OCCURRENCE','1')): raise SystemExit(1)
if action==os.environ.get('TRUST_TEST_NONOBJECT_PHASE'):
 print('null'); raise SystemExit(0)
value={'ok':True,'owner':'00000000-0000-4000-8000-000000000001','guest_uuid':'00000000-0000-4000-8000-000000000002','endpoint':'https://disposable.example.invalid','tls_sha256':'a'*64,'isolated':True,'non_production':True,'reservation_current':True,'core_version':os.environ.get('TRUST_TEST_CORE_VERSION','26.7.8_2'),'owned_count':0,'exact_restored':True,'cipher_choice':'DEFAULT'}
value.update(json.loads(os.environ.get('TRUST_TEST_ATTESTATION_OVERRIDE','{}')))
if action=='verify-operation': value['cryptographic_writes']=1 if (root/'trace').read_text().splitlines().count(action)<=2 else 0
if action=='before-destroy': value['owned_count']=1
value['settings_destroy_equal']=True
for key in ('config_sha256','settings_sha256','routes_sha256','runtime_sha256'): value[key]='a'*64
if action=='verify-restored' and os.environ.get('TRUST_TEST_MISMATCH'): value['config_sha256']='b'*64
print(json.dumps(value))
""")
        helper.chmod(0o700)
        binary = root / "bin"
        binary.mkdir()
        fake = """#!/usr/bin/env python3
import os,pathlib,signal,sys,time
name=pathlib.Path(sys.argv[0]).name
if name=='openssl':
 pathlib.Path('import.crt').write_text('fixture-certificate')
 pathlib.Path('import.key').write_text('fixture-private-key')
if name=='terraform' and sys.argv[1]=='apply':
 if os.environ.get('TRUST_TEST_APPLY_SIGNAL'): os.kill(os.getppid(), signal.SIGTERM); time.sleep(30)
 if os.environ.get('TRUST_TEST_APPLY_HANG'): time.sleep(30)
 if not os.environ.get('TRUST_TEST_APPLY_SUCCESS'): raise SystemExit(1)
"""
        for name in ("go", "openssl", "terraform"):
            command = binary / name
            command.write_text(fake)
            command.chmod(0o700)
        runtime = {
            "owner": "00000000-0000-4000-8000-000000000001",
            "guest_uuid": "00000000-0000-4000-8000-000000000002",
            "endpoint": "https://disposable.example.invalid",
            "tls_sha256": "a" * 64,
            "source_sha256": "d" * 64,
            "guard": str(helper),
            "oracle": str(helper),
            "guard_sha256": acceptance.sha256(helper),
            "oracle_sha256": acceptance.sha256(helper),
        }
        path = root / "runtime.json"
        path.write_text(json.dumps(runtime))
        path.chmod(0o600)
        return {
            "OPNSENSE_TRUST_RUNTIME": str(path),
            "OPNSENSE_URI": runtime["endpoint"],
            "OPNSENSE_API_KEY": "fixture-key",
            "OPNSENSE_API_SECRET": "fixture-secret",
            "OPNSENSE_ALLOW_INSECURE": "false",
            "TRUST_TEST_ROOT": str(root),
            "PATH": str(binary) + os.pathsep + os.environ["PATH"],
        }

    def test_wrapper_checkout_worktree_mounts_and_private_boundaries(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            checkout = root / "ordinary"
            scripts = checkout / "test/scripts"
            (scripts / "lib").mkdir(parents=True)
            original = Path(__file__).parent
            for name in ("run-story-31-0f-acceptance.sh", "lib/log.sh"):
                shutil.copyfile(original / name, scripts / name)
            for args in (
                ["init", "-q"],
                ["add", "."],
                [
                    "-c",
                    "user.name=Fixture",
                    "-c",
                    "user.email=fixture@example.invalid",
                    "commit",
                    "-qm",
                    "fixture",
                ],
            ):
                subprocess.run(
                    ["git", "-C", str(checkout), *args], check=True, capture_output=True
                )
            worktree = root / "worktree"
            subprocess.run(
                [
                    "git",
                    "-C",
                    str(checkout),
                    "worktree",
                    "add",
                    "--detach",
                    "-q",
                    str(worktree),
                ],
                check=True,
                capture_output=True,
            )
            common = (checkout / ".git").resolve()
            private = root / "private"
            private.mkdir()
            runtime = private / "runtime.json"
            runtime.write_text("{}")
            ca = private / "ca.pem"
            ca.write_text("fixture-ca")
            outside_ca = root / "outside.pem"
            outside_ca.write_text("not-mounted")
            ca_link = private / "ca-link.pem"
            ca_link.symlink_to(outside_ca)
            ancestor_runtime = root / "runtime.json"
            ancestor_runtime.write_text("{}")
            ancestor_link = private / "ancestor.json"
            ancestor_link.symlink_to(ancestor_runtime)
            shim = root / "bin"
            shim.mkdir()
            record = root / "docker.json"
            docker = shim / "docker"
            docker.write_text(
                "#!" + sys.executable + "\n"
                "import json,os,pathlib,sys\n"
                "pathlib.Path(os.environ['TRUST_WRAPPER_RECORD']).write_text(json.dumps({'args':sys.argv[1:],'ca':os.environ.get('SSL_CERT_FILE')}))\n"
            )
            docker.chmod(0o700)
            environment = {
                **os.environ,
                "PATH": str(shim) + os.pathsep + os.environ["PATH"],
                "TRUST_WRAPPER_RECORD": str(record),
                "OPNSENSE_TRUST_RUNTIME": str(runtime),
                "SSL_CERT_FILE": str(ca),
            }
            for repo in (checkout, worktree):
                with self.subTest(repo=repo.name):
                    wrapper = repo / "test/scripts/run-story-31-0f-acceptance.sh"
                    result = subprocess.run(
                        ["bash", str(wrapper)],
                        env=environment,
                        check=False,
                        capture_output=True,
                        text=True,
                        timeout=10,
                    )
                    self.assertEqual(result.returncode, 0, result.stderr)
                    observed = json.loads(record.read_text())
                    args = observed["args"]
                    volumes = [
                        args[index + 1] for index, arg in enumerate(args) if arg == "-v"
                    ]
                    self.assertCountEqual(
                        volumes,
                        [
                            f"{common}:{common}:ro",
                            f"{repo}:/workspace:ro",
                            f"{private}:{private}",
                            "opnsense-dhcp-go-mod:/root/go/pkg/mod",
                            "opnsense-dhcp-go-build:/root/.cache/go-build",
                        ],
                    )
                    forwarded = [
                        args[index + 1] for index, arg in enumerate(args) if arg == "-e"
                    ]
                    self.assertIn("SSL_CERT_FILE", forwarded)
                    self.assertEqual(observed["ca"], str(ca.resolve()))
                    child = repo / "private"
                    child.mkdir()
                    child_runtime = child / "runtime.json"
                    child_runtime.write_text("{}")
                    exact_runtime = repo / "runtime.json"
                    exact_runtime.write_text("{}")
                    child_link = private / (repo.name + "-child.json")
                    child_link.symlink_to(child_runtime)
                    variants = (
                        {"SSL_CERT_FILE": str(outside_ca)},
                        {"SSL_CERT_FILE": str(ca_link)},
                        {
                            "OPNSENSE_TRUST_RUNTIME": str(ancestor_runtime),
                            "SSL_CERT_FILE": str(ancestor_runtime),
                        },
                        {
                            "OPNSENSE_TRUST_RUNTIME": str(ancestor_link),
                            "SSL_CERT_FILE": str(ancestor_link),
                        },
                        {
                            "OPNSENSE_TRUST_RUNTIME": str(exact_runtime),
                            "SSL_CERT_FILE": str(exact_runtime),
                        },
                        {
                            "OPNSENSE_TRUST_RUNTIME": str(child_runtime),
                            "SSL_CERT_FILE": str(child_runtime),
                        },
                        {
                            "OPNSENSE_TRUST_RUNTIME": str(child_link),
                            "SSL_CERT_FILE": str(child_link),
                        },
                    )
                    for variant in variants:
                        with self.subTest(variant=variant):
                            record.unlink(missing_ok=True)
                            result = subprocess.run(
                                ["bash", str(wrapper)],
                                env={**environment, **variant},
                                check=False,
                                capture_output=True,
                                text=True,
                                timeout=10,
                            )
                            self.assertNotEqual(result.returncode, 0)
                            self.assertFalse(
                                record.exists(), "unsafe wrapper invoked Docker"
                            )

    def test_settings_constraints_prerequisite_and_invalid_oracle_choice(self):
        initial = {
            "store_intermediate_certs": False,
            "install_crls": False,
            "fetch_crls": False,
            "enable_legacy_sect": False,
            "enable_config_constraints": False,
            "cipher_string": [],
        }
        for enabled, independent_enabled, choice in itertools.product(
            (False, True),
            (False, True),
            ("AES256-SHA", None, "", "AES256-SHA,AES128-SHA", [], 1),
        ):
            with (
                self.subTest(
                    enabled=enabled,
                    independent_enabled=independent_enabled,
                    choice=choice,
                ),
                tempfile.TemporaryDirectory() as directory,
            ):
                root = Path(directory)
                config = root / "main.tf"
                original = acceptance.resources("fixture", "certificate", "key")
                config.write_text(original)
                initial = {
                    **initial,
                    "store_intermediate_certs": independent_enabled,
                    "install_crls": independent_enabled,
                    "fetch_crls": independent_enabled,
                    "enable_legacy_sect": independent_enabled,
                    "enable_config_constraints": enabled,
                    "cipher_string": ["AES128-SHA"] if enabled else [],
                }
                state = dict(initial)
                applied = []

                def apply(
                    label, config=config, state=state, applied=applied, initial=initial
                ):
                    block = re.search(
                        r'resource "opnsense_trust_settings" "observed" \{([^}]*)\}',
                        config.read_text(),
                    )
                    self.assertIsNotNone(block)
                    intent = {}
                    for line in block.group(1).splitlines():
                        if line.strip():
                            field, raw = line.split("=", 1)
                            intent[field.strip()] = json.loads(raw)
                    candidate = {**state, **intent}
                    # All upstream constraint fields except the tested cipher
                    # are empty. Model the nonempty requirement, so
                    # the former constraints=true fixture fails this oracle.
                    if candidate["enable_config_constraints"]:
                        self.assertTrue(
                            candidate["cipher_string"],
                            "constraints enabled without prerequisite",
                        )
                    self.assertTrue(
                        set(candidate["cipher_string"]) <= {"AES256-SHA", "AES128-SHA"}
                    )
                    if label == "settings-omitted-no-change":
                        self.assertEqual(intent, {})
                        self.assertEqual(state, initial)
                    state.update(intent)
                    applied.append((label, dict(state)))

                def observed_state(_directory, state=state):
                    return {
                        "opnsense_trust_settings.observed": dict(state),
                        "data.opnsense_trust_settings.observed": dict(state),
                    }

                with mock.patch.object(
                    acceptance, "state_values", side_effect=observed_state
                ):
                    if choice != "AES256-SHA":
                        with self.assertRaisesRegex(
                            acceptance.AcceptanceError,
                            "authoritative available CipherString choice",
                        ):
                            acceptance.exercise_settings(
                                {},
                                root,
                                config,
                                {"opnsense_trust_settings.observed": initial},
                                apply,
                                {"cipher_choice": choice},
                            )
                        self.assertEqual(applied, [])
                        self.assertEqual(config.read_text(), original)
                        continue
                    acceptance.exercise_settings(
                        {},
                        root,
                        config,
                        {"opnsense_trust_settings.observed": initial},
                        apply,
                        {"cipher_choice": choice},
                    )
                self.assertEqual(len(applied), 19)
                for field, initial_value in initial.items():
                    if field == "cipher_string":
                        continue
                    for suffix, expected in (
                        ("true", True),
                        ("false", False),
                        ("restore", initial_value),
                    ):
                        value = next(
                            state
                            for label, state in applied
                            if label == field + "-" + suffix
                        )
                        self.assertIs(value[field], expected)
                by_label = dict(applied)
                self.assertEqual(
                    by_label["enable_config_constraints-true"]["cipher_string"],
                    [choice],
                )
                self.assertFalse(
                    by_label["enable_config_constraints-false"][
                        "enable_config_constraints"
                    ]
                )
                self.assertEqual(
                    by_label["enable_config_constraints-restore"]["cipher_string"],
                    initial["cipher_string"],
                )
                self.assertEqual(
                    by_label["cipher-selection"]["cipher_string"], [choice]
                )
                self.assertEqual(by_label["cipher-empty-clear"]["cipher_string"], [])
                self.assertFalse(
                    by_label["cipher-empty-clear"]["enable_config_constraints"]
                )
                self.assertEqual(by_label["cipher-restore"], initial)
                self.assertEqual(state, initial)
                self.assertEqual(config.read_text(), original)

    def test_identity_controls_with_executable_helper(self):
        for key, unsafe in (
            ("owner", "00000000-0000-4000-8000-000000000003"),
            ("guest_uuid", "00000000-0000-4000-8000-000000000003"),
            ("endpoint", "https://other.example.invalid"),
            ("tls_sha256", "b" * 64),
            ("isolated", False),
            ("non_production", False),
            ("reservation_current", False),
            ("ok", False),
        ):
            with self.subTest(gate=key), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                environment = {
                    **self.synthetic_runtime(root),
                    "TRUST_TEST_ATTESTATION_OVERRIDE": json.dumps({key: unsafe}),
                }
                with (
                    mock.patch.dict(os.environ, environment),
                    mock.patch.object(
                        acceptance, "source_digest", return_value="d" * 64
                    ),
                    self.assertRaises(acceptance.AcceptanceError),
                ):
                    acceptance.main()
                self.assertEqual((root / "trace").read_text().splitlines(), ["guard"])
                self.assertFalse(list(root.glob("trust-attempt-*/snapshot.json")))

    def test_exact_firmware_gate_with_executable_helper(self):
        for version in (
            "26.7.8_2",
            "26.7.8",
            "26.7.8_1",
            "26.7.8_3",
            "26.7.9",
            "26.7.8_2-extra",
            "",
        ):
            with (
                self.subTest(version=version),
                tempfile.TemporaryDirectory() as directory,
            ):
                root = Path(directory)
                environment = {
                    **self.synthetic_runtime(root),
                    "TRUST_TEST_CORE_VERSION": version,
                }
                with (
                    mock.patch.dict(os.environ, environment),
                    mock.patch.object(
                        acceptance, "source_digest", return_value="d" * 64
                    ),
                ):
                    if version == "26.7.8_2":
                        # Reach the synthetic Terraform failure only after both
                        # real executable guard and snapshot helper phases pass.
                        with self.assertRaisesRegex(
                            acceptance.AcceptanceError, "command failed"
                        ):
                            acceptance.main()
                        trace = (root / "trace").read_text().splitlines()
                        self.assertEqual(trace[:2], ["guard", "snapshot"])
                        self.assertEqual(
                            trace[-4:],
                            ["cleanup", "verify-zero", "restore", "verify-restored"],
                        )
                    else:
                        with self.assertRaisesRegex(
                            acceptance.AcceptanceError,
                            "exact acceptance firmware required",
                        ):
                            acceptance.main()
                        self.assertEqual(
                            (root / "trace").read_text().splitlines(), ["guard"]
                        )
                        self.assertFalse(
                            list(root.glob("trust-attempt-*/snapshot.json"))
                        )

    def test_failed_terraform_attempts_every_cleanup_phase(self):
        variants = (
            {},
            *(
                {"TRUST_TEST_FAIL_PHASE": phase}
                for phase in ("cleanup", "verify-zero", "restore", "verify-restored")
            ),
            {"TRUST_TEST_NONOBJECT_PHASE": "cleanup"},
            {"TRUST_TEST_MISMATCH": "1"},
        )
        for variant in variants:
            with (
                self.subTest(variant=variant),
                tempfile.TemporaryDirectory() as directory,
            ):
                root = Path(directory)
                environment = {**self.synthetic_runtime(root), **variant}
                output = io.StringIO()
                with (
                    mock.patch.dict(os.environ, environment),
                    mock.patch.object(
                        acceptance, "source_digest", return_value="d" * 64
                    ),
                    contextlib.redirect_stdout(output),
                    self.assertRaises(acceptance.AcceptanceError),
                ):
                    acceptance.main()
                trace = (root / "trace").read_text().splitlines()
                self.assertEqual(
                    trace[-4:], ["cleanup", "verify-zero", "restore", "verify-restored"]
                )
                self.assertNotIn("TRUST_LIVE_MATRIX", output.getvalue())
                self.assertFalse(list(root.glob("trust-attempt-*/core-passed")))

    def test_cleanup_failure_and_hash_mismatch_veto_successful_body(self):
        observed = {
            "opnsense_trust_ca.root": {
                "certificate": "CERTIFICATE-root",
                "id": "root",
                "description": "root",
            },
            "opnsense_trust_cert.local": {
                "generated_private_key": "PRIVATE KEY-local",
                "returned_certificate": "CERTIFICATE-local",
            },
            "data.opnsense_trust_cert.csr_holder": {
                "csr": "CERTIFICATE REQUEST-holder"
            },
            "opnsense_trust_cert.external": {
                "id": "holder",
                "returned_private_key": "PRIVATE KEY-holder",
                "returned_csr": "CERTIFICATE REQUEST-holder",
                "returned_certificate": "CERTIFICATE-signed",
            },
            "opnsense_trust_cert.internal": {
                "id": "leaf",
                "description": "leaf",
                "returned_certificate": "CERTIFICATE-leaf",
            },
            "opnsense_trust_cert.signed": {
                "returned_certificate": "CERTIFICATE-signed"
            },
        }
        for address in (
            "opnsense_trust_ca.imported", "opnsense_trust_cert.imported",
            "data.opnsense_trust_ca.legacy", "data.opnsense_trust_cert.legacy",
        ):
            observed[address] = {"valid_from": "-157766400", "valid_to": "-126230400"}
        variants = (
            {},
            *(
                {"TRUST_TEST_FAIL_PHASE": phase}
                for phase in ("cleanup", "verify-zero", "restore", "verify-restored")
            ),
            {"TRUST_TEST_NONOBJECT_PHASE": "cleanup"},
            {"TRUST_TEST_MISMATCH": "1"},
        )
        for variant in variants:
            with (
                self.subTest(variant=variant),
                tempfile.TemporaryDirectory() as directory,
            ):
                root = Path(directory)
                environment = {
                    **self.synthetic_runtime(root),
                    "TRUST_TEST_APPLY_SUCCESS": "1",
                    **variant,
                    **(
                        {"TRUST_TEST_FAIL_OCCURRENCE": "2"}
                        if variant.get("TRUST_TEST_FAIL_PHASE") == "verify-zero"
                        else {}
                    ),
                }
                output = io.StringIO()
                with (
                    mock.patch.dict(os.environ, environment),
                    mock.patch.object(
                        acceptance, "source_digest", return_value="d" * 64
                    ),
                    mock.patch.object(
                        acceptance, "state_values", return_value=observed
                    ),
                    mock.patch.object(acceptance, "verify_signed_key"),
                    mock.patch.object(acceptance, "verify_private_key"),
                    mock.patch.object(acceptance, "verify_reissue"),
                    mock.patch.object(acceptance, "verify_adoption"),
                    mock.patch.object(acceptance, "exercise_settings"),
                    contextlib.redirect_stdout(output),
                ):
                    if variant:
                        with self.assertRaisesRegex(
                            acceptance.AcceptanceError, "cleanup/restoration failed"
                        ):
                            acceptance.main()
                    else:
                        acceptance.main()
                self.assertEqual(
                    (root / "trace").read_text().splitlines()[-4:],
                    ["cleanup", "verify-zero", "restore", "verify-restored"],
                )
                self.assertEqual(
                    "TRUST_LIVE_MATRIX" in output.getvalue(), not bool(variant)
                )
                self.assertTrue(
                    list(root.glob("trust-attempt-*/core-passed")),
                    "positive synthetic body did not reach cleanup",
                )

    def test_timeout_and_cancellation_restore_after_owned_command_reaped(self):
        original_run = acceptance.run_private
        for mode, expected in (
            ("TRUST_TEST_APPLY_HANG", subprocess.TimeoutExpired),
            ("TRUST_TEST_APPLY_SIGNAL", acceptance.AcceptanceError),
        ):
            with self.subTest(mode=mode), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                environment = {**self.synthetic_runtime(root), mode: "1"}
                prior_signal = signal.getsignal(signal.SIGTERM)

                def bounded(*args, **kwargs):
                    return original_run(*args, **kwargs, timeout=0.3)

                with (
                    mock.patch.dict(os.environ, environment),
                    mock.patch.object(
                        acceptance, "source_digest", return_value="d" * 64
                    ),
                    mock.patch.object(acceptance, "run_private", side_effect=bounded),
                    self.assertRaises(expected),
                ):
                    acceptance.main()
                self.assertEqual(
                    (root / "trace").read_text().splitlines()[-4:],
                    ["cleanup", "verify-zero", "restore", "verify-restored"],
                )
                self.assertEqual(signal.getsignal(signal.SIGTERM), prior_signal)


if __name__ == "__main__":
    unittest.main()
