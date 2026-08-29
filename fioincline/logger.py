"""Асинхронный/синхронный логгер запросов (порт logging.go)."""

import os
import threading
import time


def _now_stamp():
    return time.strftime("%Y-%m-%d %H:%M:%S", time.localtime()) + ".%03d" % (int(time.time() * 1000) % 1000)


class Logger:
    def __init__(self, path, mode="async", flush_ms=50, buffer_kb=64, enabled=True):
        self.path = path
        self.mode = mode if mode == "sync" else "async"
        self.flush_ms = flush_ms if flush_ms > 0 else 50
        self.max_buf = buffer_kb * 1024 if buffer_kb > 0 else 64 * 1024
        self.enabled = enabled

        self._lock = threading.Lock()
        self.last_ip = ""
        self.buf = []
        self.buf_size = 0
        self.stopped = False
        self._started = False

        if self.mode == "async" and enabled:
            self._started = True
            self._stop = threading.Event()
            self._thread = threading.Thread(target=self._flusher, daemon=True)
            self._thread.start()

    def set_last_ip(self, ip):
        with self._lock:
            self.last_ip = ip

    def is_enabled(self):
        return self.enabled

    def _write_direct(self, line):
        try:
            with open(self.path, "a", encoding="utf-8") as f:
                f.write(line)
        except OSError:
            pass

    def _write_batch(self, batch):
        if not batch:
            return
        try:
            with open(self.path, "a", encoding="utf-8") as f:
                f.write("".join(batch))
        except OSError:
            pass

    def log(self, ip, msg):
        if not self.enabled:
            return
        ts = _now_stamp()
        addr = "-"
        with self._lock:
            if self.stopped:
                pass
            if ip != "":
                addr = ip
            elif self.last_ip != "":
                addr = self.last_ip
            self.last_ip = ip
        line = "[%s] %s %s\n" % (ts, addr, msg)

        if self.mode == "sync":
            self._write_direct(line)
            return

        with self._lock:
            if self.stopped:
                self._write_direct(line)
                return
            self.buf.append(line)
            self.buf_size += len(msg) + 40
            need = self.buf_size >= self.max_buf
        if need:
            self._flush()

    def _flush(self):
        with self._lock:
            if self.stopped:
                return
            batch = self.buf
            self.buf = []
            self.buf_size = 0
        self._write_batch(batch)

    def _flusher(self):
        while not self._stop.wait(self.flush_ms / 1000.0):
            self._flush()
        self._flush()

    def close(self):
        if not self.enabled:
            return
        with self._lock:
            if self.stopped:
                return
            self.stopped = True
            batch = self.buf
            self.buf = []
            self.buf_size = 0
            started = self._started
        if started:
            self._stop.set()
            self._thread.join(timeout=2)
        self._write_batch(batch)
