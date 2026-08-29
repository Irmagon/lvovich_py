"""HTTP-сервис на FastAPI: REST, SOAP, WSDL, Swagger UI, auth, логгер.

Порт internal/server на Python (uvicorn). Поведение — функциональный
эквивалент исходного Go/Express-сервиса.
"""

import hashlib
import base64
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse

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


class AppState:
    def __init__(self, cfg, log_path):
        self.cfg = cfg
        self.log = Logger(log_path, cfg.log_mode, cfg.flush_ms, cfg.buffer_kb, cfg.logging)
        self.core = Core()
        self.soap = SoapHandler(self.core, self.log)


def _client_ip(request):
    host = request.client.host if request.client else ""
    return host


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


def _expr_str_field(body, name):
    if body is None:
        return "", True
    v = body.get(name)
    if v is None:
        return "", True
    if v.type == J.J_STR:
        return v.str, True
    return "", False


def _syntax_error_html(data):
    # Функциональный эквивалент: стандартная позиция (0) достаточно для не-бит-в-бит.
    return "SyntaxError: Expected double-quoted property name in JSON at position 0 (line 1 column 1)"


def _body_log_text(body):
    if body is None:
        return "undefined"
    return J.j_stringify(body)


def _reorder_fio(v):
    if v is None or v.type != J.J_OBJ:
        return v
    keys = ["SurName", "FirstName", "SecondName"]
    out = J.obj()
    for k in keys:
        val = v.get(k)
        if val is not None:
            out.set(k, val)
    for k in v.keys:
        if k in keys:
            continue
        out.set(k, v.vals[k])
    return out


def create_app(cfg, log_path="server.log"):
    st = AppState(cfg, log_path)

    @asynccontextmanager
    async def lifespan(app):
        yield
        st.log.close()

    app = FastAPI(docs_url=None, redoc_url=None, openapi_url=None, lifespan=lifespan)

    # Статичные данные WSDL/Swagger
    base = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(base, "wsdl", "service.wsdl"), "rb") as f:
        wsdl_bytes = f.read()

    asset_dir = os.path.join(base, "swagger", "static")
    asset_ct = {
        "index.html": "text/html; charset=utf-8",
        "swagger-ui-init.js": "application/javascript; charset=utf-8",
        "swagger-ui.css": "text/css; charset=utf-8",
        "swagger-ui-bundle.js": "text/javascript; charset=utf-8",
        "swagger-ui-standalone-preset.js": "text/javascript; charset=utf-8",
        "favicon-32x32.png": "image/png",
        "favicon-16x16.png": "image/png",
    }

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

    # ---------------- REST ----------------
    async def _read_body(request):
        raw = await request.body()
        if len(raw) > MAX_BODY:
            return None, False, 413
        return raw, True, None

    @app.post("/api/incline")
    async def api_incline(request: Request):
        return await _rest_incline(request, st)

    @app.post("/api/gender")
    async def api_gender(request: Request):
        return await _rest_gender(request, st)

    for _path, _kind in [("/api/city/in", "in"), ("/api/city/from", "from"), ("/api/city/to", "to")]:
        def _mk(kind):
            async def h(request: Request):
                return await _rest_city(request, st, kind)
            return h
        app.add_api_route(_path, _mk(_kind), methods=["POST"], include_in_schema=False)

    for _path, _kind in [("/api/org/in", "in"), ("/api/org/from", "from"), ("/api/org/to", "to")]:
        def _mk(kind):
            async def h(request: Request):
                return await _rest_org(request, st, kind)
            return h
        app.add_api_route(_path, _mk(_kind), methods=["POST"], include_in_schema=False)

    # ---------------- SOAP / WSDL ----------------
    @app.get("/wsdl")
    async def wsdl_get(request: Request):
        return Response(content=wsdl_bytes, media_type="text/xml; charset=utf-8",
                        headers={"X-Powered-By": "Express"})

    @app.get("/soap")
    async def soap_get(request: Request):
        if "wsdl" in request.query_params:
            return Response(content=wsdl_bytes, media_type="application/xml",
                            headers={"X-Powered-By": "Express"})
        return Response(status_code=200, headers={"X-Powered-By": "Express"})

    @app.post("/soap")
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

    # ---------------- Swagger UI ----------------
    @app.get("/api-docs")
    @app.get("/api-docs/")
    async def api_docs(request: Request):
        if not st.cfg.swagger:
            return _err_response(request, 404, "Error", "Cannot GET /api-docs")
        return await _serve_asset(request, "index.html")

    @app.get("/api-docs/{name}")
    async def api_docs_asset(request: Request, name: str):
        if not st.cfg.swagger:
            return _err_response(request, 404, "Error", "Cannot GET /api-docs/" + name)
        return await _serve_asset(request, name)

    @app.exception_handler(404)
    async def not_found(request: Request, exc):
        return _err_response(request, 404, "Error", "Cannot %s %s" % (request.method, request.url.path))

    async def _serve_asset(request, name):
        path = os.path.join(asset_dir, name)
        if name not in asset_ct or not os.path.isfile(path):
            return _err_response(request, 404, "Error", "Cannot GET " + request.url.path)
        data = open(path, "rb").read()
        etag = weak_etag(data)
        if request.headers.get("If-None-Match") == etag:
            return Response(status_code=304, headers={"ETag": etag, "X-Powered-By": "Express"})
        headers = {"Content-Type": asset_ct[name], "X-Powered-By": "Express", "ETag": etag}
        return Response(content=data, headers=headers)

    # ---------------- helpers ----------------
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

    async def _parse_body(request):
        ct = request.headers.get("Content-Type", "")
        if not ct.startswith("application/json"):
            return None, True, None
        raw = await request.body()
        request.state.body_text = raw.decode("utf-8", "replace")
        if len(raw) > MAX_BODY:
            return None, False, _err_response(request, 413, "Error",
                                              "PayloadTooLargeError: request entity too large")
        if not raw.strip():
            return J.obj(), True, None
        try:
            v = J.parse_json(raw.decode("utf-8"))
            return v, True, None
        except (ValueError, UnicodeDecodeError):
            return None, False, _err_response(request, 400, "Error", _syntax_error_html(raw.decode("utf-8", "replace")))

    async def _rest_incline(request, st):
        body, ok, err = await _parse_body(request)
        if not ok:
            return err
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

    async def _rest_gender(request, st):
        body, ok, err = await _parse_body(request)
        if not ok:
            return err
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

    async def _rest_city(request, st, kind):
        body, ok, err = await _parse_body(request)
        if not ok:
            return err
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

    async def _rest_org(request, st, kind):
        body, ok, err = await _parse_body(request)
        if not ok:
            return err
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

    # Логирование доступа — добавим в обработчики
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
        response = await call_next(request)
        if st.log.is_enabled() and path.startswith("/api"):
            _log_access(ip, request)
        return response

    return app


def make_server(cfg=None, config_path="config.ini", log_path="server.log"):
    """Создаёт приложение; cfg=None — прочитать из config.ini."""
    if cfg is None:
        cfg = Config(config_path)
    return create_app(cfg, log_path)
