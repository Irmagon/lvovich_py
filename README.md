# fioincline — склонение русских ФИО, городов и организаций (Python)

Сервис склонения русских фамилий, имён, отчеств, названий городов и
организаций по падежам. Предоставляет REST, SOAP, WSDL и Swagger UI.

**Движок склонения — Python-порт библиотеки [nodkz/lvovich](https://github.com/nodkz/lvovich).**

Сервис написан на **Python 3** поверх **FastAPI** и запускается под **uvicorn**.

---

## Благодарности / Acknowledgements

Правила склонения и определения пола, а также движок падежей заимствованы из
замечательной JavaScript-библиотеки **[nodkz/lvovich](https://github.com/nodkz/lvovich)**:

> Склонение русских ФИО и городов, выдержавшее проверку временем.
> Спасибо [@nodkz](https://github.com/nodkz) и всем контрибьюторам за огромную
> работу по сбору и проверке правил русского языка.

В свою очередь, `nodkz/lvovich` наследует правила из проекта
[Petrovich](https://github.com/petrovich/petrovich-js) (правила склонения ФИО).

Проект сохраняет MIT-лицензию с указанием авторства исходных правил и кода —
см. файл [LICENSE](./LICENSE).

---

## Возможности

- Склонение фамилии, имени и отчества по 6 падежам
- Склонение названий городов (предложный, родительный, винительный)
- Склонение названий организаций (предложный, родительный, винительный)
- Определение пола по ФИО (male / female / androgynous)
- REST + SOAP + WSDL + Swagger UI
- Конфигурация через `config.ini` (порт, токен, whitelist IP, Swagger)
- Асинхронный логгер запросов

## Структура

```
fioincline/
  lvovich/           ядро склонения (типы, правила, ФИО/город/организация/пол)
  config.py          чтение config.ini
  core.py            тонкая обёртка над ядром
  jsonx.py           JSON с сохранением порядка ключей
  logger.py          асинхронный/синхронный логгер
  soap.py            SOAP-сервис
  app.py             FastAPI-приложение (маршруты, auth, swagger)
  wsdl/service.wsdl  WSDL-схема
  swagger/static/    статика Swagger UI
main.py              точка входа
tests/               pytest-тесты (ядро + сервер)
bench/benchmark.py   бенчмарк производительности
```

## Требования

- Python 3.10+
- зависимости из `requirements.txt`: `fastapi`, `uvicorn`

Установка зависимостей:

```
python -m pip install -r requirements.txt
```

## Быстрый старт

```
python main.py
```

Или напрямую через uvicorn:

```
python -m uvicorn main:app --host 0.0.0.0 --port 3000
```

Сервер запустится на адресе и порте из `config.ini` (по умолчанию `0.0.0.0:3000`).

## Endpoint-ы

| Интерфейс | Путь | Назначение |
|---|---|---|
| SOAP | `/soap` | SOAP-сервис |
| WSDL | `/wsdl` | WSDL-схема |
| REST | `/api/incline` | Склонение ФИО (JSON) |
| REST | `/api/gender` | Определение пола (JSON) |
| REST | `/api/city/in` | Город предложный (JSON) |
| REST | `/api/city/from` | Город родительный (JSON) |
| REST | `/api/city/to` | Город винительный (JSON) |
| REST | `/api/org/in` | Организация предложный (JSON) |
| REST | `/api/org/from` | Организация родительный (JSON) |
| REST | `/api/org/to` | Организация винительный (JSON) |
| Swagger | `/api-docs` | Интерактивная документация |

### Пример REST-запроса

```
POST /api/incline
Content-Type: application/json

{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}
```

Ответ:

```json
{"SurName":"Иванову","FirstName":"Ивану","SecondName":"Ивановичу","gender":"male"}
```

## Конфигурация

```ini
[server]
address = 0.0.0.0
port = 3000
swagger = true

[auth]
token = mysecret
allowed_ips = 127.0.0.1, ::1

[logging]
enabled = true
mode = async
flush_ms = 50
buffer_kb = 64
```

`[logging] enabled = false` полностью отключает запись в `server.log`
(ускоряет обработку запросов).

Полное описание API и конфигурации — в файле [`service-doc.txt`](./service-doc.txt).

## Тесты

```
python -m pytest tests
```

Или через `run-tests.cmd` (результат в `test-output/`). Прогоняются тесты ядра
по фикстуре `nodkz/lvovich` (склонение ФИО и городов, организации) и тесты
HTTP-сервера (REST/SOAP/WSDL/Swagger/auth).

## Производительность

Измерить пропускную способность можно бенчмарком (аналог Go-версии):

```
python bench/benchmark.py --n 2000 --threads 8
```

Измеряются ~мс/оп и ~запросов/сек для REST- и SOAP-эндпоинтов последовательно
и в пуле потоков. Методика и замеры — в [`service-doc.txt`](./service-doc.txt).

## Лицензия

MIT. См. файл [LICENSE](./LICENSE).
