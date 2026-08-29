"""Склонение городов — порт city.ts / city.go (nodkz/lvovich)."""

from . import types as T
from .rules import find_rule, apply_rule, apply_mod, ends_with
from .incline import incline_firstname
from .rules_data import city_rules, frozen_words, frozen_parts, frozen_parts_after


def split_city(name):
    """Аналог name.split(/(\\s|-)/g): разделители (' ' и '-') сохраняются."""
    parts = []
    buf = []
    for r in name:
        if r == " " or r == "-":
            parts.append("".join(buf))
            buf = []
            parts.append(r)
        else:
            buf.append(r)
    parts.append("".join(buf))
    return parts


def is_frozen(str_, words):
    lower = str_.lower()
    for w in words:
        if w == lower:
            return True
    return False


def is_frozen_part(part, i, parts, frozen_parts, frozen_parts_after):
    if len(parts) > 1:
        if is_frozen(part, frozen_parts):
            return True
        for k in range(0, i):
            if is_frozen(parts[k], frozen_parts_after):
                return True
    return False


def decline_to(name, word_case, gender):
    if is_frozen(name, frozen_words):
        return name
    parts = split_city(name)
    out = list(parts)
    for i, part in enumerate(parts):
        if is_frozen_part(part, i, parts, frozen_parts, frozen_parts_after):
            out[i] = part
            continue
        rule = find_rule(part, T.ANDROGYNOUS, city_rules, False)
        if rule is not None:
            out[i] = apply_rule(rule, part, word_case)
        else:
            res = incline_firstname(part, word_case, T.get_gender_const(gender))
            if res == "":
                res = part
            out[i] = res
    return "".join(out)


def city_in(name, gender):
    return decline_to(name, T.PREPOSITIONAL, gender)


def city_from(name, gender):
    return decline_to(name, T.GENITIVE, gender)


def city_to(name):
    if name == "":
        return name
    parts = split_city(name)
    out = list(parts)
    for i, part in enumerate(parts):
        if is_frozen_part(part, i, parts, frozen_parts, frozen_parts_after):
            out[i] = part
            continue
        part_lower = part.lower()
        if ends_with(part_lower, "а"):
            out[i] = apply_mod(part, "-у")
        elif ends_with(part_lower, "ая"):
            out[i] = apply_mod(part, "--ую")
        elif ends_with(part_lower, "ия"):
            out[i] = apply_mod(part, "--ию")
        elif ends_with(part_lower, "я"):
            out[i] = apply_mod(part, "-ю")
        else:
            out[i] = part
    return "".join(out)
