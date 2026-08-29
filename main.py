"""Точка входа сервера fioincline (uvicorn).

Запуск:  python main.py   (или:  uvicorn main:app --host 0.0.0.0 --port 3000)
или после установки:  fioincline
"""
import os

from fioincline import Config, create_app


def _find_config():
    for cand in ("config.ini", os.path.join(os.path.dirname(os.path.abspath(__file__)), "config.ini")):
        if os.path.isfile(cand):
            return cand
    return "config.ini"


config_path = _find_config()
_cfg = Config(config_path)
_log_path = os.path.join(os.path.dirname(os.path.abspath(config_path)), "server.log")
app = create_app(_cfg, _log_path)


def run():
    """Консольный запуск через entry point (fioincline)."""
    import uvicorn

    uvicorn.run(app, host=_cfg.address, port=_cfg.port, log_level="info")


if __name__ == "__main__":
    run()
