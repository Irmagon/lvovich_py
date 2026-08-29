"""Бенчмарк производительности сервера fioincline (аналог bench_test.go).

Измеряет ~мс/оп и ~запросов/сек для REST- и SOAP-эндпоинтов двумя способами:
  - последовательно (одно ядро/один поток);
  - параллельно (пул потоков поверх ThreadingHTTPServer).

Использование:
    python bench/benchmark.py [--n N] [--threads T]
"""

import argparse
import json
import os
import sys
import threading
import time

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from fastapi.testclient import TestClient
from fioincline import Config, create_app

SOAP_BODY = (
    '<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
    'xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline>'
    "<LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName>"
    "<Declension>dative</Declension></tns:Incline></soap:Body></soap:Envelope>"
)

ENDPOINTS = {
    "REST /api/incline full": ("post", "/api/incline",
                               json.dumps({"SurName": "Иванов", "FirstName": "Иван",
                                           "SecondName": "Иванович", "declension": "dative"}, ensure_ascii=False)),
    "REST /api/incline initials": ("post", "/api/incline",
                                   json.dumps({"SurName": "Петров", "FirstName": "Пётр",
                                               "SecondName": "Петрович", "declension": "genitive",
                                               "format": "initials"}, ensure_ascii=False)),
    "REST /api/gender": ("post", "/api/gender",
                         json.dumps({"SurName": "Смирнова", "FirstName": "Анна"}, ensure_ascii=False)),
    "REST /api/city/in": ("post", "/api/city/in",
                          json.dumps({"name": "Москва", "gender": "female"}, ensure_ascii=False)),
    "REST /api/city/from": ("post", "/api/city/from",
                            json.dumps({"name": "Москва", "gender": "female"}, ensure_ascii=False)),
    "REST /api/city/to": ("post", "/api/city/to", json.dumps({"name": "Москва"}, ensure_ascii=False)),
    "SOAP Incline": ("post", "/soap", SOAP_BODY),
}

AUTH = "Bearer testtoken999"


def make_client(tmp_log="server.log"):
    cfg = Config("config.ini")
    cfg.token = "testtoken999"
    cfg.allowed_ips = []
    app = create_app(cfg, tmp_log)
    return TestClient(app)


def _req(client, method, url, body):
    if url == "/soap":
        return client.post(url, data=body, headers={
            "Authorization": AUTH, "Content-Type": "text/xml"})
    return client.post(url, content=body.encode("utf-8"), headers={
        "Authorization": AUTH, "Content-Type": "application/json"})


def bench_serial(client, method, url, body, n):
    # прогрев
    _req(client, method, url, body)
    start = time.perf_counter()
    for _ in range(n):
        r = _req(client, method, url, body)
        if r.status_code != 200:
            raise RuntimeError("status %d" % r.status_code)
    dur = time.perf_counter() - start
    per = dur / n * 1000.0  # мс/оп
    return per, n / dur


def bench_parallel(client, method, url, body, n, threads):
    _req(client, method, url, body)
    per_thread = n // threads
    errors = []

    def worker():
        try:
            for _ in range(per_thread):
                r = _req(client, method, url, body)
                if r.status_code != 200:
                    errors.append(r.status_code)
        except Exception as e:  # noqa: BLE001
            errors.append(str(e))

    start = time.perf_counter()
    ts = [threading.Thread(target=worker) for _ in range(threads)]
    for t in ts:
        t.start()
    for t in ts:
        t.join()
    dur = time.perf_counter() - start
    if errors:
        raise RuntimeError("errors: %s" % errors[:5])
    total = per_thread * threads
    per = dur / total * 1000.0
    return per, total / dur


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--n", type=int, default=2000)
    ap.add_argument("--threads", type=int, default=8)
    ap.add_argument("--log", default=os.path.join(os.path.dirname(__file__), "..", "server.log"))
    args = ap.parse_args()

    client = make_client(args.log)
    print("=" * 70)
    print("BENCHMARK fioincline (FastAPI + uvicorn / TestClient)")
    print("n=%d  threads=%d" % (args.n, args.threads))
    print("=" * 70)
    print("\nSEQUENTIAL (one thread):")
    print("%-34s %10s %14s" % ("Endpoint", "~мс/оп", "~запросов/сек"))
    for name, (m, u, b) in ENDPOINTS.items():
        per, rps = bench_serial(client, m, u, b, args.n)
        print("%-34s %10.4f %14.0f" % (name, per, rps))

    print("\nPARALLEL (%d threads):" % args.threads)
    print("%-34s %10s %14s" % ("Endpoint", "~мс/оп", "~запросов/сек"))
    for name, (m, u, b) in ENDPOINTS.items():
        per, rps = bench_parallel(client, m, u, b, args.n, args.threads)
        print("%-34s %10.4f %14.0f" % (name, per, rps))


if __name__ == "__main__":
    main()
