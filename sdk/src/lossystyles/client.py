"""IPC client — sends NDJSON messages to the Go CLI over a Unix socket."""

import json
import socket
import os


class Client:
    def __init__(self, sock_path: str):
        self.sock_path = sock_path
        self._sock: socket.socket | None = None

    def connect(self, timeout: float = 10.0) -> None:
        """Connect to the CLI's Unix socket, retrying until timeout."""
        import time

        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            try:
                self._sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
                self._sock.connect(self.sock_path)
                return
            except (ConnectionRefusedError, FileNotFoundError):
                self._sock = None
                time.sleep(0.1)
        raise ConnectionError(f"Could not connect to {self.sock_path} within {timeout}s")

    def send(self, msg: dict) -> None:
        """Send a single NDJSON message."""
        if self._sock is None:
            raise RuntimeError("Not connected")
        data = json.dumps(msg, separators=(",", ":")) + "\n"
        self._sock.sendall(data.encode())

    def close(self) -> None:
        if self._sock is not None:
            try:
                self._sock.close()
            except OSError:
                pass
            self._sock = None

    @staticmethod
    def socket_path(run_id: str) -> str:
        return os.path.join("/tmp", f"lossystyles-{run_id}.sock")
