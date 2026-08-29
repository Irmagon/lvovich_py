"""HTTP-сервис на FastAPI: REST, SOAP, WSDL, автодокументация (Swagger UI), auth, логгер.

Порт internal/server на Python (uvicorn). Поведение — функциональный
эквивалент исходного Go/Express-сервиса.

Автодокументация — нативная генерация FastAPI:
  /api-docs     Swagger UI (автогенерация из маршрутов)
  /openapi.json OpenAPI-схема
  /redoc        ReDoc (альтернативная документация)

Включается/выключается параметром [server] swagger в config.ini.
"""

import hashlib
import base64
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pydantic import BaseModel, ConfigDict

from .config import Config
from .core import Core
from .logger import Logger
from .soap import SoapHandler
from . import jsonx as J

MAX_BODY = 100 * 1024


def weak_etag(body_bytes):
    """Express-совместимый слабый ETag: W/"<len-hex>-<base64(sha1)>"."""
    h = hashlib.sha1(body_bytes).digest()
    enc = base64.b64encode(h).decode("ascii").rstrip("=")
    return 'W/"%x-%s"' % (len(body_bytes), enc)


def _err_page_html(title, pre):
    return ('<!DOCTYPE html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n'
            "<title>%s</title>\n</head>\n<body>\n<pre>%s</pre>\n</body>\n</html>\n" % (title, pre))


# ---------------------------------------------------------------- request models (для автодокументации)
class _FioReq(BaseModel):
    model_config = ConfigDict(extra="ignore")
    SurName: str = ""
    FirstName: str = ""
    SecondName: str = ""


class InclineReq(_FioReq):
    declension: str = ""
    format: str = ""


class GenderReq(_FioReq):
    pass


class NameReq(BaseModel):
    model_config = ConfigDict(extra="ignore")
    name: str = ""
    gender: str = ""


class AppState:
    def __init__(self, cfg, log_path):
        self.cfg = cfg
        self.log = Logger(log_path, cfg.log_mode, cfg.flush_ms, cfg.buffer_kb, cfg.logging)
        self.core = Core()
        self.soap = SoapHandler(self.core, self.log)


def _client_ip(request):
    return request.client.host if request.client else ""


def _ip_allowed(cfg, ip):
    if not cfg.allowed_ips:
        return True
    return ip in cfg.allowed_ips


def _auth_ok(cfg, request):
    if not cfg.token:
        return True
    hdr = request.headers.get("Authorization", "")
    if hdr.startswith("Bearer "):
        return hdr[len("Bearer "):] == cfg.token
    return False


def _syntax_error_html(data):
    return "SyntaxError: Expected double-quoted property name in JSON at position 0 (line 1 column 1)"


def _expr_str_field(body, name):
    if body is None:
        return "", True
    v = body.get(name)
    if v is None:
        return "", True
    if v.type == J.J_STR:
        return v.str, True
    return "", False


def _model_to_jv(body):
    """Pydantic-модель -> JValue."""
    if body is None:
        return None
    return J.from_py(body.model_dump())


# ---------------------------------------------------------------- обработчики REST (возвращают Response)
def _rest_incline(request, st, body):
    if body is None or body.type != J.J_OBJ:
        return _send_json(request, 500, '{"error":"Cannot read properties of undefined (reading \'declension\')"}')
    declension, _ = _expr_str_field(body, "declension")
    fmt, _ = _expr_str_field(body, "format")
    sur, ok = _expr_str_field(body, "SurName")
    if not ok:
        return _send_json(request, 500, '{"error":"SurName.trim is not a function"}')
    first, ok = _expr_str_field(body, "FirstName")
    if not ok:
        return _send_json(request, 500, '{"error":"FirstName.trim is not a function"}')
    second, ok = _expr_str_field(body, "SecondName")
    if not ok:
        return _send_json(request, 500, '{"error":"SecondName.trim is not a function"}')
    from .lvovich import Person
    res = st.core.incline(Person(first, sur, second), declension, fmt)
    out = J.obj()
    if fmt == "initials":
        out.set("SurName", J.Str(res.sur_name))
        out.set("initials", J.Str(res.initials))
    else:
        if sur != "":
            out.set("SurName", J.Str(res.sur_name))
        if first != "":
            out.set("FirstName", J.Str(res.first_name))
        if second != "":
            out.set("SecondName", J.Str(res.second_name))
        out.set("gender", J.Null() if res.gender == "" else J.Str(res.gender))
    return _send_json(request, 200, J.j_stringify(out))


def _rest_gender(request, st, body):
    if body is not None and body.type != J.J_OBJ:
        return _send_json(request, 500, '{"error":"Cannot read properties of undefined (reading \'SurName\')"}')
    sur = first = second = ""
    if body is not None:
        sur, ok = _expr_str_field(body, "SurName")
        if not ok:
            return _send_json(request, 500, '{"error":"SurName.trim is not a function"}')
        first, ok = _expr_str_field(body, "FirstName")
        if not ok:
            return _send_json(request, 500, '{"error":"FirstName.trim is not a function"}')
        second, ok = _expr_str_field(body, "SecondName")
        if not ok:
            return _send_json(request, 500, '{"error":"SecondName.trim is not a function"}')
    from .lvovich import Person
    g = st.core.get_gender(Person(first, sur, second))
    out = J.obj()
    out.set("gender", J.Null() if g == "" else J.Str(g))
    return _send_json(request, 200, J.j_stringify(out))


def _rest_city(request, st, kind, body):
    name_v = body.get("name") if body is not None else None
    if name_v is None or not name_v.is_string():
        if kind == "to":
            return _send_json(request, 200, "{}")
        return _send_json(request, 500, '{"error":"Cannot read properties of undefined (reading \'toLowerCase\')"}')
    name, _ = name_v.str_val()
    gender, _ = _expr_str_field(body, "gender")
    if kind == "in":
        out = st.core.city_in(name, gender)
    elif kind == "from":
        out = st.core.city_from(name, gender)
    else:
        out = st.core.city_to(name)
    o = J.obj()
    o.set("name", J.Str(out))
    return _send_json(request, 200, J.j_stringify(o))


def _rest_org(request, st, kind, body):
    name_v = body.get("name") if body is not None else None
    if name_v is None or not name_v.is_string():
        return _send_json(request, 200, "{}")
    name, _ = name_v.str_val()
    if kind == "in":
        out = st.core.org_in(name)
    elif kind == "from":
        out = st.core.org_from(name)
    else:
        out = st.core.org_to(name)
    o = J.obj()
    o.set("name", J.Str(out))
    return _send_json(request, 200, J.j_stringify(o))


def _send_json(request, status, body_str):
    body_bytes = body_str.encode("utf-8")
    etag = weak_etag(body_bytes)
    if status == 200 and request.headers.get("If-None-Match") in (etag, "*"):
        return Response(status_code=304, headers={"ETag": etag, "X-Powered-By": "Express"})
    return Response(content=body_str, status_code=status,
                    media_type="application/json; charset=utf-8",
                    headers={"ETag": etag, "X-Powered-By": "Express"})


def _err_response(request, status, title, pre):
    body = _err_page_html(title, pre).encode("utf-8")
    return Response(content=body, status_code=status,
                    media_type="text/html; charset=utf-8",
                    headers={"X-Powered-By": "Express"})


def create_app(cfg, log_path="server.log"):
    st = AppState(cfg, log_path)

    if cfg.swagger:
        app = FastAPI(title="fioincline",
                      description="Склонение русских ФИО, городов и организаций",
                      version="1.0.0",
                      docs_url="/api-docs",
                      openapi_url="/openapi.json",
                      redoc_url="/redoc",
                      lifespan=_make_lifespan(st))
    else:
        app = FastAPI(docs_url=None, openapi_url=None, redoc_url=None,
                      lifespan=_make_lifespan(st))

    # WSDL-схема для SOAP (функциональная часть, не документация).
    base = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(base, "wsdl", "service.wsdl"), "rb") as f:
        wsdl_bytes = f.read()

    def _log_access(ip, request, prefix=None):
        if not st.log.is_enabled():
            return
        body = ""
        try:
            body = request.state.body_text
        except AttributeError:
            body = ""
        pre = (prefix + " ") if prefix else ""
        st.log.log(ip, "%s%s %s - %s" % (pre, request.method, request.url.path, body))

    # ---------------- SOAP / WSDL ----------------
    @app.get("/wsdl", include_in_schema=False)
    async def wsdl_get(request: Request):
        return Response(content=wsdl_bytes, media_type="text/xml; charset=utf-8",
                        headers={"X-Powered-By": "Express"})

    @app.get("/soap", include_in_schema=False)
    async def soap_get(request: Request):
        if "wsdl" in request.query_params:
            return Response(content=wsdl_bytes, media_type="application/xml",
                            headers={"X-Powered-By": "Express"})
        return Response(status_code=200, headers={"X-Powered-By": "Express"})

    @app.post("/soap", include_in_schema=False)
    async def soap_post(request: Request):
        ip = _client_ip(request)
        if not _auth_ok(st.cfg, request):
            return _send_json(request, 401, '{"error":"Unauthorized"}')
        raw = await request.body()
        action = request.headers.get("SOAPAction", "")
        status, body = st.soap.handle(raw.decode("utf-8", "replace"), action)
        _log_access(ip, request)
        return Response(content=body, status_code=status, media_type="text/xml",
                        headers={"X-Powered-By": "Express"})

    # ---------------- REST (Pydantic-модели -> автодокументация) ----------------
    @app.post("/api/incline", response_model=dict,
              summary="Склонение ФИО",
              description="Склоняет фамилию, имя и отчество по заданному падежу. "
                          "declension: nominative|genitive|dative|accusative|instrumental|prepositional, "
                          "format: full|initials.")
    async def api_incline(request: Request, body: InclineReq):
        return _rest_incline(request, st, _model_to_jv(body))

    @app.post("/api/gender", response_model=dict,
              summary="Определение пола",
              description="Определяет пол по ФИО (male / female / androgynous).")
    async def api_gender(request: Request, body: GenderReq):
        return _rest_gender(request, st, _model_to_jv(body))

    @app.post("/api/city/in", response_model=dict, summary="Город в предложном падеже")
    async def api_city_in(request: Request, body: NameReq):
        return _rest_city(request, st, "in", _model_to_jv(body))

    @app.post("/api/city/from", response_model=dict, summary="Город в родительном падеже")
    async def api_city_from(request: Request, body: NameReq):
        return _rest_city(request, st, "from", _model_to_jv(body))

    @app.post("/api/city/to", response_model=dict, summary="Город в винительном падеже")
    async def api_city_to(request: Request, body: NameReq):
        return _rest_city(request, st, "to", _model_to_jv(body))

    @app.post("/api/org/in", response_model=dict, summary="Организация в предложном падеже")
    async def api_org_in(request: Request, body: NameReq):
        return _rest_org(request, st, "in", _model_to_jv(body))

    @app.post("/api/org/from", response_model=dict, summary="Организация в родительном падеже")
    async def api_org_from(request: Request, body: NameReq):
        return _rest_org(request, st, "from", _model_to_jv(body))

    @app.post("/api/org/to", response_model=dict, summary="Организация в винительном падеже")
    async def api_org_to(request: Request, body: NameReq):
        return _rest_org(request, st, "to", _model_to_jv(body))

    # ---------------- ошибки валидации тела (400 вместо 422) ----------------
    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(request, exc):
        return _err_response(request, 400, "Error", _syntax_error_html(""))

    # ---------------- middleware: auth / IP / 413 / лог ----------------
    @app.middleware("http")
    async def auth_ip_middleware(request, call_next):
        ip = _client_ip(request)
        path = request.url.path
        if path == "/wsdl" and request.method == "GET":
            pass
        elif path.startswith("/api") or path.startswith("/soap"):
            if not _ip_allowed(st.cfg, ip):
                return _send_json(request, 403, '{"error":"Forbidden"}')
            if not _auth_ok(st.cfg, request):
                return _send_json(request, 401, '{"error":"Unauthorized"}')
        # Ограничение размера тела для JSON-REST (413).
        if path.startswith("/api") and request.method == "POST":
            content_length = request.headers.get("content-length")
            if content_length and content_length.isdigit() and int(content_length) > MAX_BODY:
                return _err_response(request, 413, "Error",
                                     "PayloadTooLargeError: request entity too large")
            raw = await request.body()
            request.state.body_text = raw.decode("utf-8", "replace")
            if len(raw) > MAX_BODY:
                return _err_response(request, 413, "Error",
                                     "PayloadTooLargeError: request entity too large")
        response = await call_next(request)
        if st.log.is_enabled() and path.startswith("/api"):
            _log_access(ip, request)
        return response

    return app


def _make_lifespan(st):
    @asynccontextmanager
    async def lifespan(app):
        yield
        st.log.close()
    return lifespan


def make_server(cfg=None, config_path="config.ini", log_path="server.log"):
    """Создаёт приложение; cfg=None — прочитать из config.ini."""
    if cfg is None:
        cfg = Config(config_path)
    return create_app(cfg, log_path)
