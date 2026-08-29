"""Тесты ядра lvovich по фикстуре nodkz/lvovich."""
import json
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest

from fioincline.lvovich import (
    Person, get_gender, incline, city_in, city_from, city_to,
)

FIXTURE = os.path.join(os.path.dirname(__file__), "..", "fioincline", "testdata", "fixture.json")


@pytest.fixture(scope="module")
def fx():
    with open(FIXTURE, encoding="utf-8") as f:
        return json.load(f)


def test_fio_fixture(fx):
    for i, f in enumerate(fx["fio"]):
        inp = f["input"]
        p = Person(inp.get("FirstName", ""), inp.get("SurName", ""), inp.get("SecondName", ""))
        assert get_gender(p) == (f.get("gender") or ""), f"gender #{i}"
        for decl, exp in f.get("cases", {}).items():
            if decl == "initials":
                r = incline(p, "genitive", "initials")
                assert r.sur_name == (exp.get("SurName") or ""), f"#{i}/initials SurName"
                assert r.initials == (exp.get("initials") or ""), f"#{i}/initials"
            else:
                r = incline(p, decl, "")
                assert r.first_name == (exp.get("FirstName") or ""), f"#{i}/{decl} FirstName"
                assert r.sur_name == (exp.get("SurName") or ""), f"#{i}/{decl} SurName"
                assert r.second_name == (exp.get("SecondName") or ""), f"#{i}/{decl} SecondName"
                assert r.gender == (exp.get("gender") or ""), f"#{i}/{decl} gender"


def test_cities_fixture(fx):
    for i, c in enumerate(fx["cities"]):
        name = c["input"]
        assert city_in(name, "") == (c["in"] or ""), f"city#{i} in"
        assert city_in(name, "female") == (c["inF"] or ""), f"city#{i} inF"
        assert city_from(name, "") == (c["from"] or ""), f"city#{i} from"
        assert city_from(name, "female") == (c["fromF"] or ""), f"city#{i} fromF"
        assert city_to(name) == (c["to"] or ""), f"city#{i} to"
