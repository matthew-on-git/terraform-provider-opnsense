#!/usr/bin/env python3
"""Bounded, private IPsec acceptance; no provisioning or target fallback.

OPNSENSE_IPSEC_GUARDED_RUNTIME is a mode0600 JSON file accepted by the Trust
load_runtime contract, plus source_sha256, isolation_receipt and journal absolute
paths inside its mode0700 private root. Receipt must exist privately; journal may
be absent, but its existing parent must be private. timeout_seconds defaults to
2400 (allowed 60..2400). Credentials and trusted SSL_CERT_FILE are explicit.

Helpers use argv ACTION ATTEMPT_DIRECTORY OWNER and the Trust helper's exact
owner/guest_uuid/endpoint/tls_sha256/isolation/reservation/core26.7.8_2 assertions.
Guard action: guard. Oracle actions: snapshot, observe, verify-zero, cleanup,
restore, verify-restored. Snapshot and verify-restored return config_sha256,
settings_sha256, routes_sha256, runtime_sha256; final exact_restored must be true.
verify-zero returns owned_count=0 and unresolved_count=0. Oracle reads the runtime
and durable journal, never discovers arbitrary objects as cleanup authority.
Cleanup may reconcile a failed Go run, but cannot make it acceptance proof.
"""

import json
import os
import re
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path

from trust_lifecycle_acceptance import (
    AcceptanceError,
    helper,
    load_runtime,
    owned_command as generic_owned_command,
    require,
    sha256,
    source_digest,
)

TEST = "TestAccIPsecAdvancedDisposable"
HASHES = ("config_sha256", "settings_sha256", "routes_sha256", "runtime_sha256")


class QuiescenceError(AcceptanceError):
    """Restoration cannot run concurrently with an unproven owned process."""


def group_live(group):
    """Pinned Linux container: zombies cannot execute appliance mutations."""
    require(Path("/proc/self/stat").exists(), "Linux process proof required")
    for entry in Path("/proc").iterdir():
        if not entry.name.isdigit():
            continue
        try:
            # comm can contain spaces and parentheses; parse after its last ')'.
            fields = (entry / "stat").read_text().rsplit(")", 1)[1].split()
            if int(fields[2]) == group and fields[0] != "Z":
                return True
        except (FileNotFoundError, ProcessLookupError):
            continue
    return False


def owned_command(command, **kwargs):
    """Reuse group termination, then prove no live group members before restore."""
    with tempfile.TemporaryDirectory(prefix="ipsec-process-") as temporary:
        pid_file = Path(temporary) / "group"
        launcher = (
            "import os,sys; "
            "open(sys.argv[1],'w').write(str(os.getpgrp())); "
            "os.execvp(sys.argv[2],sys.argv[2:])"
        )
        try:
            return generic_owned_command(
                [sys.executable, "-c", launcher, str(pid_file), *command], **kwargs
            )
        finally:
            try:
                if pid_file.exists():
                    group = int(pid_file.read_text())
                    try:
                        os.killpg(group, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
                    deadline = time.monotonic() + 3
                    while group_live(group) and time.monotonic() < deadline:
                        time.sleep(0.01)
                    if group_live(group):
                        raise QuiescenceError(
                            "owned processes remain; restoration unsafe"
                        )
            except (OSError, ValueError, AcceptanceError) as exc:
                raise QuiescenceError("owned process quiescence unproven") from exc


def private_path(runtime, key, missing=False):
    raw = runtime.get(key)
    require(isinstance(raw, str) and Path(raw).is_absolute(), "private path required")
    path = Path(raw)
    require(not path.is_symlink(), "private input must not be a symlink")
    path = path.resolve()
    require(path.is_relative_to(runtime["root"]), "private input outside runtime")
    require(
        path.parent.is_dir() and path.parent.stat().st_mode & 0o077 == 0,
        "private input parent required",
    )
    if path.exists():
        require(
            path.is_file() and path.stat().st_mode & 0o077 == 0,
            "private regular input required",
        )
    else:
        require(missing, "private input missing")
    return path


def verify_go_log(path):
    ran = passed = False
    for line in path.read_text().splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        if not isinstance(event, dict) or event.get("Test") != TEST:
            continue
        require(
            event.get("Action") not in ("skip", "fail"), "live test skipped or failed"
        )
        ran |= event.get("Action") == "run"
        passed |= event.get("Action") == "pass"
    require(ran and passed, "non-skipped live Go PASS missing")


def verify_zero(value):
    require(
        value.get("owned_count") == 0 and value.get("unresolved_count") == 0,
        "owned or unresolved mutations remain",
    )


def execute(runtime, directory, runtime_path, runtime_digest, candidate):
    """Every post-snapshot path runs all independent reconciliation stages."""
    baseline = None
    failure = None
    cleanup_errors = []

    def canceled(_number, _frame):
        raise AcceptanceError("acceptance interrupted")

    old = {s: signal.signal(s, canceled) for s in (signal.SIGINT, signal.SIGTERM)}
    try:
        helper(runtime, "guard", "guard", directory)
        # A snapshot helper may fail after writing its snapshot: finally must
        # still attempt restoration, without converting partial output to proof.
        baseline = helper(runtime, "oracle", "snapshot", directory)
        for key in HASHES:
            require(
                re.fullmatch(r"[0-9a-f]{64}", baseline.get(key, "")),
                "baseline hash missing",
            )
        helper(runtime, "guard", "guard", directory)
        require(
            source_digest() == candidate and sha256(runtime_path) == runtime_digest,
            "inputs changed before mutation",
        )
        timeout = runtime.get("timeout_seconds", 2400)
        require(
            type(timeout) is int and 60 <= timeout <= 2400, "invalid execution bound"
        )
        with (directory / "go-test.jsonl").open("wb") as log:
            result = owned_command(
                [
                    "go",
                    "test",
                    "-json",
                    "-count=1",
                    "./internal/service/ipsec",
                    "-run",
                    "^" + TEST + "$",
                    "-timeout",
                    str(timeout - 15) + "s",
                ],
                cwd=Path(__file__).resolve().parents[2],
                stdout=log,
                stderr=log,
                timeout=timeout,
            )
        require(result.returncode == 0, "live Go command failed")
        verify_go_log(directory / "go-test.jsonl")
        helper(runtime, "oracle", "observe", directory)
        verify_zero(helper(runtime, "oracle", "verify-zero", directory))
    except BaseException as exc:
        failure = exc
    finally:
        # owned_command has terminated/reaped the Go/Terraform group before this
        # point. Further signals cannot interrupt the bounded restoration stages.
        for s in old:
            signal.signal(s, signal.SIG_IGN)
        for action in ("cleanup", "verify-zero", "restore", "verify-restored"):
            if isinstance(failure, QuiescenceError):
                cleanup_errors.append("unsafe-process-group")
                break
            try:
                value = helper(runtime, "oracle", action, directory)
                if action == "verify-zero":
                    verify_zero(value)
                if action == "verify-restored":
                    require(
                        baseline is not None and value.get("exact_restored") is True,
                        "exact restoration missing",
                    )
                    require(
                        all(value.get(key) == baseline.get(key) for key in HASHES),
                        "baseline not restored",
                    )
            except (AcceptanceError, OSError, ValueError, subprocess.SubprocessError):
                cleanup_errors.append(action)
        for s, previous in old.items():
            signal.signal(s, previous)
    require(not cleanup_errors, "independent cleanup/restoration failed")
    if failure is not None:
        raise failure
    require(source_digest() == candidate, "source changed during acceptance")
    require(sha256(runtime_path) == runtime_digest, "runtime changed during acceptance")


def main():
    os.umask(0o077)
    runtime_path = os.environ.get("OPNSENSE_IPSEC_GUARDED_RUNTIME")
    runtime = load_runtime(runtime_path)
    require(
        Path(runtime["root"]).stat().st_mode & 0o077 == 0,
        "private runtime root required",
    )
    require(os.environ.get("OPNSENSE_URI") == runtime["endpoint"], "endpoint mismatch")
    require(
        os.environ.get("OPNSENSE_API_KEY") and os.environ.get("OPNSENSE_API_SECRET"),
        "explicit ephemeral credentials required",
    )
    require(
        os.environ.get("OPNSENSE_ALLOW_INSECURE", "false").lower() != "true",
        "verified TLS required",
    )
    receipt = private_path(runtime, "isolation_receipt")
    journal = private_path(runtime, "journal", missing=True)
    require(receipt != journal, "receipt and journal must differ")
    candidate = source_digest()
    require(
        candidate == runtime.get("source_sha256"), "reviewed source digest mismatch"
    )
    runtime_digest, receipt_digest = sha256(runtime_path), sha256(receipt)
    directory = Path(tempfile.mkdtemp(prefix="ipsec-attempt-", dir=runtime["root"]))
    (directory / "source-sha256").write_text(candidate)
    os.environ.update(
        {
            "TF_ACC": "1",
            "OPNSENSE_IPSEC_DISPOSABLE_ACK": "I_RESERVED_A_NEW_ISOLATED_GUEST",
            "OPNSENSE_IPSEC_ISOLATION_RECEIPT": str(receipt),
            "OPNSENSE_IPSEC_OWNERSHIP_JOURNAL": str(journal),
        }
    )
    execute(runtime, directory, runtime_path, runtime_digest, candidate)
    require(sha256(receipt) == receipt_digest, "isolation receipt changed")
    (directory / "accepted.json").write_text(
        json.dumps({"accepted": True, "source_sha256": candidate})
    )
    print(
        "IPSEC_LIVE_MATRIX go=true zero=true exact_restore=true inputs_unchanged=true"
    )


if __name__ == "__main__":
    try:
        main()
    except (AcceptanceError, OSError, ValueError, subprocess.SubprocessError):
        raise SystemExit(
            "IPsec acceptance failed; private evidence retained. No live proof claimed."
        ) from None
