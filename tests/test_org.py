"""Тесты склонения организаций (порт org_test.go)."""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from fioincline.lvovich import org_in, org_from, org_to, org_decline

DECLINE = [
    ("ООО «Ромашка»", "genitive", "ООО «Ромашки»"),
    ("ООО «Ромашка»", "dative", "ООО «Ромашке»"),
    ("ООО «Ромашка»", "accusative", "ООО «Ромашку»"),
    ("ООО «Ромашка»", "instrumental", "ООО «Ромашкой»"),
    ("ООО «Ромашка»", "prepositional", "ООО «Ромашке»"),
    ("ООО \"Ромашка\"", "genitive", "ООО \"Ромашки\""),
    ("ООО Ромашка", "genitive", "ООО Ромашки"),
    ("компания «Свет»", "genitive", "компания «Света»"),
    ("банк «Открытие»", "genitive", "банк «Открытия»"),
    ("Сбербанк", "genitive", "Сбербанка"),
    ("МТС", "genitive", "МТС"),
    ("ПАО «МТС»", "genitive", "ПАО «МТС»"),
    ("«Красное и белое»", "genitive", "«Красного и белого»"),
    ("завод имени Ленина", "genitive", "завод имени Ленина"),
]


def test_org_decline():
    for name, dec, want in DECLINE:
        assert org_decline(name, dec) == want, f"org_decline({name!r}, {dec!r})"


def test_org_in_from_to():
    cases = [
        ("ООО «Ромашка»", "ООО «Ромашке»", "ООО «Ромашки»", "ООО «Ромашку»"),
        ("Сбербанк", "Сбербанке", "Сбербанка", "Сбербанк"),
        ("МТС", "МТС", "МТС", "МТС"),
        ("банк «Открытие»", "банк «Открытии»", "банк «Открытия»", "банк «Открытие»"),
    ]
    for name, in_w, from_w, to_w in cases:
        assert org_in(name) == in_w, f"org_in({name!r})"
        assert org_from(name) == from_w, f"org_from({name!r})"
        assert org_to(name) == to_w, f"org_to({name!r})"
