"""Точка входа сервера fioincline (uvicorn).

Запуск:  python main.py   (или:  uvicorn main:app --host 0.0.0.0 --port 3000)
"""
import os
import sys

from fioincline import Config, create_app

config_path = "config.ini"
if not os.path.isfile(config_path):
    exe_dir = os.path.dirname(os.path.abspath(__file__))
    cand = os.path.join(exe_dir, "config.ini")
    if os.path.isfile(cand):
        config_path = cand
_cfg = Config(config_path)
_log_path = os.path.join(os.path.dirname(os.path.abspath(config_path)), "server.log")
app = create_app(_cfg, _log_path)

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host=_cfg.address, port=_cfg.port, log_level="info")
