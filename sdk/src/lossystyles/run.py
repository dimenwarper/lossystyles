"""Run — the core session object that manages the lifecycle of a lossystyles run."""

from __future__ import annotations

import subprocess
import shutil
import uuid
from typing import Any

from lossystyles.client import Client
from lossystyles.config import RunConfig


class Run:
    def __init__(self, config: RunConfig, auto_launch: bool = True):
        self.config = config
        self.run_id = uuid.uuid4().hex[:12]
        self.auto_launch = auto_launch
        self._client: Client | None = None
        self._process: subprocess.Popen | None = None  # type: ignore[type-arg]
        self._step = 0
        self._started = False

    def start(self) -> None:
        sock_path = Client.socket_path(self.run_id)

        if self.auto_launch:
            self._launch_cli(sock_path)

        self._client = Client(sock_path)
        self._client.connect()

        self._client.send({
            "type": "init",
            "run_id": self.run_id,
            "project": self.config.project,
            "config": self.config.config,
            "theme": self.config.theme,
        })
        self._started = True

    def log(self, metrics: dict[str, float], step: int | None = None) -> None:
        if not self._started or self._client is None:
            raise RuntimeError("Run not started. Call start() first.")

        if step is not None:
            self._step = step
        else:
            self._step += 1

        self._client.send({
            "type": "log",
            "run_id": self.run_id,
            "step": self._step,
            "metrics": metrics,
        })

    def finish(self) -> None:
        if not self._started or self._client is None:
            return

        try:
            self._client.send({
                "type": "finish",
                "run_id": self.run_id,
            })
        except OSError:
            pass

        self._client.close()
        self._started = False

        if self._process is not None:
            self._process.wait(timeout=5)
            self._process = None

    def _launch_cli(self, sock_path: str) -> None:
        binary = shutil.which("lossystyles")
        if binary is None:
            raise RuntimeError(
                "lossystyles CLI binary not found in PATH. "
                "Build it with: cd cli && go build -o lossystyles ./cmd/lossystyles"
            )

        self._process = subprocess.Popen(
            [binary, "--run-id", self.run_id, "--theme", self.config.theme],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

    def __enter__(self) -> Run:
        return self

    def __exit__(self, *_: Any) -> None:
        self.finish()
