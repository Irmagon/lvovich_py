"""fioincline — склонение русских ФИО, городов и организаций (Python/FastAPI)."""

from .config import Config
from .core import Core
from .logger import Logger
from .app import make_server, create_app

__all__ = ["Config", "Core", "Logger", "make_server", "create_app"]
__version__ = "1.0.0"
