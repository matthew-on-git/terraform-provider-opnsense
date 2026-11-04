"""Offline negative controls for the disposable Trust acceptance boundary."""

import contextlib
import importlib.util
import io
import json
import os
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
                core_version="26.7.8",
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
value={'ok':True,'owner':'00000000-0000-4000-8000-000000000001','guest_uuid':'00000000-0000-4000-8000-000000000002','endpoint':'https://disposable.example.invalid','tls_sha256':'a'*64,'isolated':True,'non_production':True,'reservation_current':True,'core_version':'26.7.8','owned_count':0,'exact_restored':True,'cipher_choice':'DEFAULT'}
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
            "opnsense_trust_cert.local": {"generated_private_key": "PRIVATE KEY-local"},
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
