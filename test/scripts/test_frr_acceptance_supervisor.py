"""Offline process/ownership controls, not appliance acceptance evidence."""

import copy
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import frr_acceptance_supervisor as supervisor

GUEST = "00000000-0000-4000-8000-000000000001"
COMMAND = ["bash", "-c", "fixture-test-only"]


class SupervisorControls(unittest.TestCase):
    def binding(self, root):
        return supervisor.Containers(root, GUEST, COMMAND)

    def record(self, root):
        return {
            "Id": "a" * 64,
            "Config": {
                "Cmd": COMMAND,
                "Image": "ghcr.io/devrail-dev/dev-toolchain:1.12.24",
                "WorkingDir": "/workspace",
                "Entrypoint": None,
                "Env": ["OPNSENSE_DISPOSABLE_VM_UUID=" + GUEST],
            },
            "Mounts": [
                {
                    "Type": "bind",
                    "Source": str(Path(root).resolve()),
                    "Destination": "/workspace",
                }
            ],
        }

    def test_exact_container_binding_rejects_each_near_match(self):
        with tempfile.TemporaryDirectory() as root:
            binding = self.binding(root)
            record = self.record(root)
            self.assertTrue(binding.owns(record))
            mutations = (
                ("Config", "Cmd", [*COMMAND, "extra"]),
                ("Config", "Image", "unrelated"),
                ("Config", "WorkingDir", "/elsewhere"),
                ("Config", "Entrypoint", ["other"]),
                ("Config", "Env", ["OPNSENSE_DISPOSABLE_VM_UUID=" + GUEST + "extra"]),
            )
            for section, key, bad in mutations:
                candidate = copy.deepcopy(record)
                candidate[section][key] = bad
                self.assertFalse(binding.owns(candidate))
            for key, bad in (
                ("Source", root + "-other"),
                ("Destination", "/other"),
                ("Type", "volume"),
            ):
                candidate = copy.deepcopy(record)
                candidate["Mounts"][0][key] = bad
                self.assertFalse(binding.owns(candidate))

    def test_container_inventory_and_exact_kill(self):
        with tempfile.TemporaryDirectory() as root:
            binding = self.binding(root)
            own = self.record(root)
            other = copy.deepcopy(own)
            other["Id"] = "b" * 64
            other["Config"]["Env"] = []
            calls = []
            running = [own, other]

            def docker(argv):
                calls.append(argv)
                if argv[0] == "ps":
                    return "\n".join(r["Id"] for r in running)
                if argv[0] == "inspect":
                    return json.dumps(running)
                self.assertEqual(argv, ["kill", own["Id"]])
                running.remove(own)
                return ""

            with mock.patch.object(binding, "call", side_effect=docker):
                binding.terminate()
            self.assertEqual(running, [other])
            self.assertIn(["kill", own["Id"]], calls)

    def test_container_failure_and_persistent_container_fail_closed(self):
        with tempfile.TemporaryDirectory() as root:
            binding = self.binding(root)
            with (
                mock.patch.object(binding, "call", return_value="not-an-id"),
                self.assertRaises(supervisor.QuiescenceError),
            ):
                binding.matching()
            with (
                mock.patch.object(binding, "matching", return_value=["a" * 64]),
                mock.patch.object(binding, "call"),
                mock.patch.object(supervisor.time, "monotonic", side_effect=[0, 11]),
                self.assertRaises(supervisor.QuiescenceError),
            ):
                binding.terminate()

    def run_fixture(self, code, timeout):
        with tempfile.TemporaryDirectory() as root, tempfile.TemporaryFile() as log:
            binding = mock.Mock()
            binding.matching.return_value = []
            pidfile = Path(root) / "child-pid"
            command = [sys.executable, "-c", code, str(pidfile)]
            with mock.patch.object(supervisor, "Containers", return_value=binding):
                result = supervisor.run_wrapper(
                    command,
                    repo=root,
                    guest_uuid=GUEST,
                    container_command=COMMAND,
                    log=log,
                    timeout_seconds=timeout,
                )
            self.assertTrue(result.host_group_empty)
            self.assertTrue(result.containers_empty)
            binding.terminate.assert_called_once()
            if pidfile.exists():
                pid = int(pidfile.read_text())
                # Child was deliberately in the wrapper's group, not its own.
                if Path("/proc/self/stat").exists():
                    stat = Path("/proc") / str(pid) / "stat"
                    self.assertTrue(
                        not stat.exists()
                        or stat.read_text().rsplit(")", 1)[1].split()[0] == "Z"
                    )
                else:
                    ps = subprocess.run(
                        ["ps", "-p", str(pid), "-o", "stat="],
                        text=True,
                        capture_output=True,
                        check=False,
                    )
                    self.assertTrue(
                        not ps.stdout.strip() or ps.stdout.strip().startswith("Z")
                    )
            return result

    def test_successful_leader_with_lingering_term_resistant_child(self):
        code = (
            "import os,signal,sys,time; pid=os.fork(); "
            "\nif pid == 0:\n signal.signal(signal.SIGTERM,signal.SIG_IGN); "
            "open(sys.argv[1],'w').write(str(os.getpid())); time.sleep(60)"
            "\nelse:\n time.sleep(.2)"
        )
        result = self.run_fixture(code, 2)
        self.assertEqual(result.returncode, 0)
        self.assertFalse(result.timed_out)

    def test_real_whole_wrapper_timeout_kills_children(self):
        code = (
            "import os,sys,time; pid=os.fork(); "
            "\nif pid == 0: open(sys.argv[1],'w').write(str(os.getpid()))"
            "\ntime.sleep(60)"
        )
        result = self.run_fixture(code, 0.4)
        self.assertTrue(result.timed_out)
        self.assertNotEqual(result.returncode, 0)

    def test_sigterm_cancellation_is_not_acceptance(self):
        code = "import os,signal,time; os.kill(os.getppid(),signal.SIGTERM); time.sleep(60)"
        result = self.run_fixture(code, 2)
        self.assertTrue(result.canceled)
        self.assertNotEqual(result.returncode, 0)

    def test_inventory_error_prevents_any_wrapper_start(self):
        with tempfile.TemporaryDirectory() as root, tempfile.TemporaryFile() as log:
            binding = mock.Mock()
            binding.matching.side_effect = subprocess.TimeoutExpired("docker", 10)
            with (
                mock.patch.object(supervisor, "Containers", return_value=binding),
                mock.patch.object(supervisor.subprocess, "Popen") as start,
            ):
                with self.assertRaises(supervisor.QuiescenceError):
                    supervisor.run_wrapper(
                        ["unused"],
                        repo=root,
                        guest_uuid=GUEST,
                        container_command=COMMAND,
                        log=log,
                    )
                start.assert_not_called()

    def test_mac_ps_live_zombie_and_malformed_inventory(self):
        for raw, expected in (
            ("12 99 S+\n13 88 S\n", True),
            ("12 99 Z+\n13 88 S\n", False),
        ):
            with (
                mock.patch.object(Path, "exists", return_value=False),
                mock.patch.object(
                    supervisor.subprocess,
                    "run",
                    return_value=subprocess.CompletedProcess([], 0, raw, ""),
                ) as ps,
            ):
                self.assertEqual(supervisor.group_live(99), expected)
                self.assertEqual(
                    ps.call_args.args[0], ["ps", "-axo", "pid=,pgid=,stat="]
                )
        with (
            mock.patch.object(Path, "exists", return_value=False),
            mock.patch.object(
                supervisor.subprocess,
                "run",
                return_value=subprocess.CompletedProcess([], 0, "malformed\n", ""),
            ),
            self.assertRaises(supervisor.QuiescenceError),
        ):
            supervisor.group_live(99)

    def test_failed_wrapper_returns_failure_after_proof(self):
        result = self.run_fixture("raise SystemExit(7)", 2)
        self.assertEqual(result.returncode, 7)
        self.assertFalse(result.timed_out)

    def test_ps_proof_failure_blocks_return_even_after_leader_exit(self):
        process = mock.Mock(pid=123456)
        with (
            mock.patch.object(
                supervisor, "group_live", side_effect=RuntimeError("private detail")
            ),
            self.assertRaisesRegex(
                supervisor.QuiescenceError, "restoration unsafe"
            ) as error,
        ):
            supervisor.quiesce(process, mock.Mock())
        self.assertNotIn("private detail", str(error.exception))


if __name__ == "__main__":
    unittest.main()
