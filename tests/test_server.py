"""Тесты HTTP-сервера (REST/SOAP/WSDL/Swagger/auth) — порт server_test.go."""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from fastapi.testclient import TestClient

from fioincline import Config, create_app

AUTH = "Bearer testtoken999"


def make_cfg(**kw):
    c = Config("config.ini")
    c.token = "testtoken999"
    c.allowed_ips = []
    for k, v in kw.items():
        setattr(c, k, v)
    return c


@pytest.fixture()
def client(tmp_path):
    cfg = make_cfg()
    app = create_app(cfg, str(tmp_path / "server.log"))
    return TestClient(app, raise_server_exceptions=True)


def _h(**extra):
    h = {"Authorization": AUTH}
    for k, v in extra.items():
        if k == "content_type":
            hk = "Content-Type"
        elif k == "soapaction":
            hk = "SOAPAction"
        else:
            hk = k.replace("_", "-").title()
        h[hk] = v
    return h


def test_rest_incline_full(client):
    r = client.post("/api/incline", json={"SurName": "Иванов", "FirstName": "Иван",
                                          "SecondName": "Иванович", "declension": "dative"}, headers=_h())
    assert r.status_code == 200
    assert r.text == '{"SurName":"Иванову","FirstName":"Ивану","SecondName":"Ивановичу","gender":"male"}'
    assert r.headers["content-type"].startswith("application/json")


def test_rest_incline_initials(client):
    r = client.post("/api/incline", json={"SurName": "Иванов", "FirstName": "Иван",
                                          "SecondName": "Иванович", "declension": "dative",
                                          "format": "initials"}, headers=_h())
    assert r.text == '{"SurName":"Иванову","initials":"И.И."}'


def test_rest_gender(client):
    r = client.post("/api/gender", json={"SurName": "Смирнова", "FirstName": "Анна"}, headers=_h())
    assert r.text == '{"gender":"female"}'


def test_rest_cities(client):
    for url, body, want in [
        ("/api/city/in", {"name": "Москва", "gender": "female"}, '{"name":"Москве"}'),
        ("/api/city/from", {"name": "Орел", "gender": "female"}, '{"name":"Орла"}'),
        ("/api/city/to", {"name": "Москва"}, '{"name":"Москву"}'),
    ]:
        r = client.post(url, json=body, headers=_h())
        assert r.text == want, url


def test_rest_orgs(client):
    for url, want in [
        ("/api/org/in", '{"name":"ООО «Ромашке»"}'),
        ("/api/org/from", '{"name":"ООО «Ромашки»"}'),
        ("/api/org/to", '{"name":"ООО «Ромашку»"}'),
    ]:
        r = client.post(url, json={"name": "ООО «Ромашка»"}, headers=_h())
        assert r.text == want, url


def test_auth_required(client):
    r = client.post("/api/incline", json={"SurName": "Иванов"}, headers={"Content-Type": "application/json"})
    assert r.status_code == 401


def test_ip_whitelist():
    cfg = make_cfg(allowed_ips=["10.0.0.1"])
    app = create_app(cfg, "server.log")
    cl = TestClient(app)
    r = cl.post("/api/incline", json={"SurName": "Иванов", "FirstName": "Иван",
                                      "SecondName": "Иванович", "declension": "dative"}, headers=_h())
    assert r.status_code == 403


def test_bad_json(client):
    r = client.post("/api/incline", content=b"{bad json", headers=_h(content_type="application/json"))
    import sys
    sys.stderr.write("\nDBG bad_json status=%s body=%r\n" % (r.status_code, r.text[:150]))
    assert r.status_code == 400


def test_payload_too_large(client):
    big = {"x": "a" * (100 * 1024 + 100)}
    r = client.post("/api/incline", json=big, headers=_h())
    assert r.status_code == 413


def test_soap_incline(client):
    soap = ('<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
            'xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline>'
            "<LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName>"
            "<Declension>genitive</Declension></tns:Incline></soap:Body></soap:Envelope>")
    r = client.post("/soap", data=soap, headers=_h(content_type="text/xml"))
    assert r.status_code == 200
    assert "LastName>Иванова</LastName" in r.text
    assert "FirstName>Ивана</FirstName" in r.text


def test_soap_city_fault_no_name(client):
    soap = ('<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
            'xmlns:tns="urn:LvovichService"><soap:Body><tns:CityIn></tns:CityIn></soap:Body></soap:Envelope>')
    r = client.post("/soap", data=soap, headers=_h(content_type="text/xml", soapaction="urn:CityIn"))
    assert r.status_code == 200
    assert "<Fault>Cannot read properties of null (reading 'Name')</Fault>" in r.text


def test_soap_orgs(client):
    for op, inner, want in [
        ("OrgIn", "<Name>ООО «Ромашка»</Name>", "<Name>ООО «Ромашке»</Name>"),
        ("OrgFrom", "<Name>ООО «Ромашка»</Name>", "<Name>ООО «Ромашки»</Name>"),
        ("OrgTo", "<Name>ООО «Ромашка»</Name>", "<Name>ООО «Ромашку»</Name>"),
    ]:
        soap = ('<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
                'xmlns:tns="urn:LvovichService"><soap:Body><tns:%s>%s</tns:%s></soap:Body></soap:Envelope>'
                % (op, inner, op))
        r = client.post("/soap", data=soap, headers=_h(content_type="text/xml"))
        assert want in r.text, op


def test_soap_bad_action(client):
    soap = ('<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
            'xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline><LastName>Иванов</LastName>'
            "</tns:Incline></soap:Body></soap:Envelope>")
    r = client.post("/soap", data=soap, headers=_h(content_type="text/xml", soapaction="urn:LvovichService"))
    assert r.status_code == 500
    assert "style" in r.text


def test_soap_unknown_op(client):
    soap = ('<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
            'xmlns:tns="urn:LvovichService"><soap:Body><tns:NoSuchOp><Name>X</Name></tns:NoSuchOp>'
            "</soap:Body></soap:Envelope>")
    r = client.post("/soap", data=soap, headers=_h(content_type="text/xml", soapaction="urn:Incline"))
    assert r.status_code == 500
    assert "description" in r.text


def test_wsdl(client):
    r = client.get("/wsdl", headers=_h())
    assert r.status_code == 200
    assert "text/xml" in r.headers["content-type"]


def test_swagger_ui(client):
    r = client.get("/api-docs", headers=_h())
    assert r.status_code == 200
    assert "text/html" in r.headers["content-type"]


def test_swagger_disabled():
    cfg = make_cfg(swagger=False)
    app = create_app(cfg, "server.log")
    cl = TestClient(app)
    r = cl.get("/api-docs", headers=_h())
    assert r.status_code == 404
