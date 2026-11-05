"""Host-side whole-wrapper deadline and quiescence gate, not acceptance proof.

Private callers supply already-attested bindings and argv/env (never shell text).
Call run_wrapper before oracle cleanup/restore, including on wrapper failure.
Only a returned result permits those actions; QuiescenceError forbids concurrent
restoration. No provisioning, credentials, guest commands or oracle calls occur
here. Timeout is for the whole wrapper, not the provider discovery deadline.
2700 seconds allows the observed 1479-second Go matrix plus build/final gates;
bounded termination/proof is additional, not part of that execution allowance.
Children must retain their new host session/process group. A detached daemon is
outside this contract; Docker daemon children have the separate exact gate below.
"""

import os
import re
import signal
import subprocess
import time
import uuid
from dataclasses import dataclass
from pathlib import Path


class QuiescenceError(RuntimeError):
    """Owned execution may remain: caller must not restore concurrently."""


@dataclass(frozen=True)
class Result:
    returncode: int
    timed_out: bool
    canceled: bool
    host_group_empty: bool = True
    containers_empty: bool = True


def group_live(group):
    """Linux procfs/macOS ps; zombies cannot execute appliance mutations."""
    if Path("/proc/self/stat").exists():
        for entry in Path("/proc").iterdir():
            if not entry.name.isdigit():
                continue
            try:
                fields = (entry / "stat").read_text().rsplit(")", 1)[1].split()
                if int(fields[2]) == group and fields[0] != "Z":
                    return True
            except (FileNotFoundError, ProcessLookupError):
                continue
        return False
    result = subprocess.run(
        ["ps", "-axo", "pid=,pgid=,stat="],
        capture_output=True,
        text=True,
        check=True,
        timeout=10,
    )
    for line in result.stdout.splitlines():
        fields = line.split()
        if len(fields) != 3 or not fields[0].isdigit() or not fields[1].isdigit():
            raise QuiescenceError("host process inventory malformed")
        if int(fields[1]) == group and not fields[2].startswith("Z"):
            return True
    return False


class Containers:
    """Do not print inspect output: it can contain inherited credentials."""

    def __init__(self, repo, guest_uuid, command):
        self.repo = str(Path(repo).resolve(strict=True))
        self.guest_uuid = str(uuid.UUID(guest_uuid))
        if self.guest_uuid != guest_uuid:
            raise ValueError("canonical guest UUID required")
        if not command or not all(isinstance(part, str) for part in command):
            raise ValueError("exact container argv required")
        self.command = list(command)

    def call(self, argv):
        return subprocess.run(
            ["docker", *argv], capture_output=True, text=True, check=True, timeout=10
        ).stdout

    def owns(self, info):
        config = info["Config"]
        return (
            config["Cmd"] == self.command
            and config["Image"] == "ghcr.io/devrail-dev/dev-toolchain:1.12.24"
            and config.get("WorkingDir") == "/workspace"
            and not config.get("Entrypoint")
            and ("OPNSENSE_DISPOSABLE_VM_UUID=" + self.guest_uuid) in config["Env"]
            and any(
                mount.get("Type") == "bind"
                and mount.get("Source") == self.repo
                and mount.get("Destination") == "/workspace"
                for mount in info["Mounts"]
            )
        )

    def matching(self):
        import json

        ids = self.call(["ps", "--no-trunc", "-q"]).split()
        if not all(re.fullmatch(r"[0-9a-f]{64}", item) for item in ids):
            raise QuiescenceError("container inventory malformed")
        if not ids:
            return []
        # One bounded inventory operation, not unbounded per-container calls.
        records = json.loads(self.call(["inspect", *ids]))
        if len(records) != len(ids) or {r["Id"] for r in records} != set(ids):
            raise QuiescenceError("container inventory incomplete")
        return [r["Id"] for r in records if self.owns(r)]

    def terminate(self):
        ids = self.matching()
        if ids:
            # kill, unlike stop, cannot exceed its grace period in the daemon.
            self.call(["kill", *ids])
        deadline = time.monotonic() + 10
        while self.matching():
            if time.monotonic() >= deadline:
                raise QuiescenceError("owned container remains")
            time.sleep(0.05)


def quiesce(process, containers, grace_seconds=5):
    """Leader exit is not proof: kill remaining group before container cleanup."""
    group = process.pid  # start_new_session makes leader PID its owned PGID.
    try:
        if group_live(group):
            try:
                os.killpg(group, signal.SIGTERM)
            except ProcessLookupError:
                pass
        deadline = time.monotonic() + grace_seconds
        while group_live(group) and time.monotonic() < deadline:
            process.poll()  # Reap leader while waiting for remaining children.
            time.sleep(0.02)
        if group_live(group):
            try:
                os.killpg(group, signal.SIGKILL)
            except ProcessLookupError:
                pass
        process.wait(timeout=10)
        deadline = time.monotonic() + 10
        while group_live(group) and time.monotonic() < deadline:
            time.sleep(0.02)
        if group_live(group):
            raise QuiescenceError("owned host process group remains")
        # Host docker client is now gone and cannot launch a new container.
        containers.terminate()
        if group_live(group) or containers.matching():
            raise QuiescenceError("owned execution remains")
    except Exception as exc:
        # Never leak argv, environment, Docker inspect or helper stderr.
        raise QuiescenceError("quiescence unproven; restoration unsafe") from exc


def run_wrapper(
    argv, *, repo, guest_uuid, container_command, log, env=None, timeout_seconds=2700
):
    """Run private argv, returning exit/deadline status only after quiescence.

    log is an already-open PRIVATE binary file; stdout/stderr never go to chat.
    env is a complete environment, or None to inherit. No shell interpolation.
    Exact Docker Config.Cmd must match the public runner's full bash argv.
    Call from the main thread (temporarily owns SIGINT/SIGTERM handling).
    QuiescenceError means NO oracle cleanup/restore until separately reconciled.
    A clean returned nonzero/timeout/canceled result permits restoration but is
    never acceptance PASS. The caller must still verify all runner proof gates.
    """
    if not 0 < timeout_seconds <= 7200:
        raise ValueError("whole-wrapper timeout must be positive and bounded")
    containers = Containers(repo, guest_uuid, container_command)
    try:
        if containers.matching():
            raise QuiescenceError(
                "pre-existing matching container; ownership ambiguous"
            )
    except Exception as exc:
        raise QuiescenceError("preflight container quiescence unproven") from exc
    canceled = False

    def cancel(_signum, _frame):
        nonlocal canceled
        canceled = True  # Repeated signals cannot interrupt quiescence.

    previous = {sig: signal.getsignal(sig) for sig in (signal.SIGINT, signal.SIGTERM)}
    process = None
    quiescence_attempted = False
    try:
        for sig in previous:
            signal.signal(sig, cancel)
        process = subprocess.Popen(
            argv, cwd=repo, env=env, stdout=log, stderr=log, start_new_session=True
        )
        deadline = time.monotonic() + timeout_seconds
        timed_out = False
        while process.poll() is None and not canceled:
            if time.monotonic() >= deadline:
                timed_out = True
                break
            time.sleep(0.05)
        quiescence_attempted = True
        quiesce(process, containers)
        return Result(process.returncode, timed_out, canceled)
    finally:
        # Unexpected errors must also pass the same quiescence gate.
        try:
            if process is not None and not quiescence_attempted:
                quiesce(process, containers)
        finally:
            for sig, handler in previous.items():
                signal.signal(sig, handler)
