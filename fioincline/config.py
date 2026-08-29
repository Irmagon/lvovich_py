"""Чтение config.ini (порт config.go)."""

import os
import re

DEFAULTS = {
    "address": "0.0.0.0",
    "port": 3000,
    "swagger": True,
    "token": "",
    "allowed_ips": [],
    "logging": True,
    "log_mode": "async",
    "flush_ms": 50,
    "buffer_kb": 64,
}

_KV_RE = re.compile(r"^\s*(\w+)\s*=\s*(.+?)\s*$")


class Config:
    def __init__(self, path):
        self.address = DEFAULTS["address"]
        self.port = DEFAULTS["port"]
        self.token = DEFAULTS["token"]
        self.allowed_ips = list(DEFAULTS["allowed_ips"])
        self.swagger = DEFAULTS["swagger"]
        self.logging = DEFAULTS["logging"]
        self.log_mode = DEFAULTS["log_mode"]
        self.flush_ms = DEFAULTS["flush_ms"]
        self.buffer_kb = DEFAULTS["buffer_kb"]

        if not os.path.isfile(path):
            return
        section = ""
        try:
            with open(path, encoding="utf-8") as f:
                for raw in f.read().split("\n"):
                    s = raw.strip()
                    if s.startswith("[") and s.endswith("]"):
                        section = s[1:-1]
                        continue
                    if s.startswith(";") or s == "":
                        continue
                    m = _KV_RE.match(s)
                    if not m:
                        continue
                    key, val = m.group(1), m.group(2).strip()
                    if section == "server":
                        if key == "address":
                            self.address = val
                        elif key == "port":
                            try:
                                self.port = int(val)
                            except ValueError:
                                pass
                        elif key == "swagger":
                            self.swagger = val == "true"
                    elif section == "auth":
                        if key == "token":
                            self.token = val
                        elif key == "allowed_ips":
                            self.allowed_ips = [ip.strip() for ip in val.split(",") if ip.strip()]
                    elif section == "logging":
                        if key == "enabled":
                            self.logging = val == "true"
                        elif key == "mode":
                            self.log_mode = val
                        elif key == "flush_ms":
                            try:
                                self.flush_ms = int(val)
                            except ValueError:
                                pass
                        elif key == "buffer_kb":
                            try:
                                self.buffer_kb = int(val)
                            except ValueError:
                                pass
        except (OSError, ValueError):
            pass
