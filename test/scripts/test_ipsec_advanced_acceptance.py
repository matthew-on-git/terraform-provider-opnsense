"""Offline controls: no appliance, credentials, or provisioning required."""

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import ipsec_advanced_acceptance as acceptance


class IPsecAcceptanceControls(unittest.TestCase):
    def test_missing_runtime_never_runs_command(self):
        with mock.patch.dict(os.environ, {}, clear=True), mock.patch.object(
            acceptance, "owned_command"
        ) as command:
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.main()
            command.assert_not_called()

    def test_non_skipped_actual_test_pass_required(self):
        with tempfile.TemporaryDirectory() as root:
            log = Path(root) / "go.jsonl"
            run = {"Test": acceptance.TEST, "Action": "run"}
            passed = {"Test": acceptance.TEST, "Action": "pass"}
            for events in (
                [],
                [passed],
                [run],
                [run, {**passed, "Action": "skip"}],
                [run, {**passed, "Action": "fail"}],
                [{**run, "Test": "Other"}, {**passed, "Test": "Other"}],
            ):
                log.write_text("\n".join(json.dumps(e) for e in events))
                with self.subTest(events=events), self.assertRaises(
                    acceptance.AcceptanceError
                ):
                    acceptance.verify_go_log(log)
            log.write_text("\n".join(json.dumps(e) for e in (run, passed)))
            acceptance.verify_go_log(log)

    def test_private_receipt_and_missing_journal_parent(self):
        with tempfile.TemporaryDirectory() as root:
            runtime = {"root": root, "journal": str(Path(root) / "journal")}
            self.assertEqual(
                acceptance.private_path(runtime, "journal", missing=True),
                Path(runtime["journal"]),
            )
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.private_path(runtime, "journal")
            Path(runtime["journal"]).write_text("{}")
            Path(runtime["journal"]).chmod(0o644)
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.private_path(runtime, "journal", missing=True)
            runtime["journal"] = str(Path(root) / "absent" / "journal")
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.private_path(runtime, "journal", missing=True)

    def test_unresolved_zero_is_not_cleanup(self):
        for value in (
            {"owned_count": 0},
            {"owned_count": 0, "unresolved_count": 1},
            {"owned_count": 1, "unresolved_count": 0},
        ):
            with self.assertRaises(acceptance.AcceptanceError):
                acceptance.verify_zero(value)

    def test_real_helper_rejects_spoofed_identity_and_firmware(self):
        with tempfile.TemporaryDirectory() as root:
            directory = Path(root)
            executable = directory / "guard"
            runtime = {
                "guard": str(executable),
                "owner": "owner-fixture",
                "guest_uuid": "guest-fixture",
                "endpoint": "https://fixture.invalid",
                "tls_sha256": "a" * 64,
            }
            truth = {
                **{
                    key: runtime[key]
                    for key in ("owner", "guest_uuid", "endpoint", "tls_sha256")
                },
                "ok": True,
                "isolated": True,
                "non_production": True,
                "reservation_current": True,
                "core_version": "26.7.8_2",
            }
            for field, bad in (
                ("owner", "other"),
                ("guest_uuid", "other"),
                ("endpoint", "https://other.invalid"),
                ("tls_sha256", "b" * 64),
                ("core_version", "26.7.8"),
                ("reservation_current", False),
            ):
                executable.write_text(
                    "#!"
                    + sys.executable
                    + "\nprint("
                    + repr(json.dumps({**truth, field: bad}))
                    + ")\n"
                )
                executable.chmod(0o700)
                runtime["guard_sha256"] = acceptance.sha256(executable)
                with self.subTest(field=field), self.assertRaises(
                    acceptance.AcceptanceError
                ):
                    acceptance.helper(runtime, "guard", "guard", directory)

    def exercise(self, failure=None):
        with tempfile.TemporaryDirectory() as root:
            directory = Path(root)
            actions = []
            baseline = dict.fromkeys(acceptance.HASHES, "a" * 64)

            def helper(_runtime, _kind, action, _directory):
                actions.append(action)
                if failure == action:
                    raise acceptance.AcceptanceError("synthetic failure")
                value = {
                    **baseline,
                    "owned_count": 0,
                    "unresolved_count": 0,
                    "exact_restored": True,
                }
                if failure == "restore-hash" and action == "verify-restored":
                    value["config_sha256"] = "b" * 64
                return value

            def command(_argv, **kwargs):
                if failure == "unsafe-process":
                    raise acceptance.QuiescenceError("synthetic live descendant")
                if failure == "timeout":
                    raise subprocess.TimeoutExpired("go", 1)
                events = [
                    {"Test": acceptance.TEST, "Action": action}
                    for action in ("run", "pass")
                ]
                kwargs["stdout"].write(
                    ("\n".join(json.dumps(e) for e in events)).encode()
                )
                return subprocess.CompletedProcess([], 1 if failure == "go" else 0)

            with mock.patch.object(
                acceptance, "helper", side_effect=helper
            ), mock.patch.object(
                acceptance, "owned_command", side_effect=command
            ), mock.patch.object(
                acceptance,
                "source_digest",
                side_effect=(
                    ["source", "changed"] if failure == "source-final" else None
                ),
                return_value="changed" if failure == "source" else "source",
            ), mock.patch.object(
                acceptance,
                "sha256",
                return_value="runtime",
                side_effect=(
                    ["runtime", "changed"] if failure == "runtime-final" else None
                ),
            ):
                if failure:
                    with self.assertRaises(
                        (acceptance.AcceptanceError, subprocess.TimeoutExpired)
                    ):
                        acceptance.execute(
                            {}, directory, "runtime.json", "runtime", "source"
                        )
                else:
                    acceptance.execute(
                        {}, directory, "runtime.json", "runtime", "source"
                    )
            if failure == "unsafe-process":
                self.assertEqual(actions, ["guard", "snapshot", "guard"])
            else:
                self.assertEqual(
                    actions[-4:],
                    ["cleanup", "verify-zero", "restore", "verify-restored"],
                )
            self.assertFalse((directory / "accepted.json").exists())
            return actions

    def test_guard_snapshot_and_go_failure_always_restore(self):
        for failure in (
            "guard",
            "snapshot",
            "go",
            "timeout",
            "source",
            "observe",
            "cleanup",
            "verify-zero",
            "restore",
            "verify-restored",
            "restore-hash",
            "source-final",
            "runtime-final",
            "unsafe-process",
        ):
            with self.subTest(failure=failure):
                self.exercise(failure)

    def test_success_requires_preflight_before_go_and_final_oracle(self):
        self.assertEqual(self.exercise()[:3], ["guard", "snapshot", "guard"])

    def test_timeout_reaps_term_ignoring_descendant_before_cleanup(self):
        with tempfile.TemporaryDirectory() as root:
            pid_file = Path(root) / "pid"
            descendant = (
                "import os,signal,time; "
                "signal.signal(signal.SIGTERM,signal.SIG_IGN); "
                f"open({str(pid_file)!r},'w').write(str(os.getpid())); time.sleep(60)"
            )
            parent = (
                "import subprocess,sys,time; "
                f"subprocess.Popen([sys.executable,'-c',{descendant!r}]); time.sleep(60)"
            )
            with (Path(root) / "log").open("wb") as log:
                with self.assertRaises(subprocess.TimeoutExpired):
                    acceptance.owned_command(
                        [sys.executable, "-c", parent],
                        stdout=log,
                        stderr=log,
                        timeout=1,
                    )
            self.assertTrue(pid_file.exists(), "descendant fixture did not start")
            pid = int(pid_file.read_text())
            # A reparented zombie is no longer executable and cannot mutate;
            # Linux init owns its eventual wait, not this process.
            status = Path(f"/proc/{pid}/stat")
            self.assertTrue(not status.exists() or status.read_text().split()[2] == "Z")


if __name__ == "__main__":
    unittest.main()
