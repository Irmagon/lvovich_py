"""Склонение названий организаций — по аналогии с city.go."""

from . import types as T
from .rules import find_rule, apply_rule, ends_with
from .incline import incline_lastname
from .rules_data import (
    org_rules, org_frozen_words, org_generic_words, org_frozen_parts, org_frozen_parts_after,
)


_QUOTE_OPEN = ["\u00ab", "\u201c", '"', "'"]
_QUOTE_CLOSE = ["\u00bb", "\u201d", '"', "'"]


def split_org(name):
    """Разбивает название организации на токены, сохраняя разделители и кавычки."""
    parts = []
    buf = []
    in_quote = ""
    quote_stack = []

    def flush():
        if buf:
            parts.append("".join(buf))
            buf.clear()

    for r in name:
        ch = r
        matched = False
        # открывающая кавычка
        for qi, qo in enumerate(_QUOTE_OPEN):
            if ch == qo and in_quote == "":
                flush()
                parts.append(ch)
                in_quote = qo
                quote_stack.append(_QUOTE_CLOSE[qi])
                matched = True
                break
        if matched:
            continue
        # закрывающая кавычка
        if quote_stack and ch == quote_stack[-1]:
            flush()
            parts.append(ch)
            quote_stack.pop()
            if not quote_stack:
                in_quote = ""
            matched = True
        if matched:
            continue
        if ch == " " or ch == "-":
            flush()
            parts.append(ch)
        else:
            buf.append(ch)
    flush()
    return parts, quote_stack


def is_generic_word(str_):
    return str_.lower() in org_generic_words


def is_frozen_org(str_):
    return str_.lower() in org_frozen_words


def is_frozen_org_part(part, i, parts):
    if len(parts) > 1:
        if part.lower() in org_frozen_parts:
            return True
        for k in range(0, i):
            if parts[k].lower() in org_frozen_parts_after:
                return True
    return False


_VOWELS = "аеёиоуыэюя"
_FEMALE_ENDINGS = ("а", "я")


def gender_by_ending(word):
    if word == "":
        return T.ANDROGYNOUS
    lower = word.lower()
    for e in _FEMALE_ENDINGS:
        if lower.endswith(e):
            return T.FEMALE
    last = lower[-1]
    if last not in _VOWELS:
        return T.MALE
    return T.ANDROGYNOUS


def is_quote(s):
    if s in _QUOTE_OPEN or s in _QUOTE_CLOSE:
        return True
    if len(s) == 1 and s in ("\u00ab", "\u00bb", "\u201c", "\u201d", '"', "'"):
        return True
    return False


def decline_org(name, word_case):
    if name == "":
        return name
    if is_frozen_org(name):
        return name

    parts, _ = split_org(name)
    if len(parts) == 0:
        return name

    generic_idx = -1
    for i, part in enumerate(parts):
        part = part.strip()
        if part == "" or part == " " or part == "-":
            continue
        if is_generic_word(part):
            generic_idx = i
            break

    start_idx = 0
    if generic_idx >= 0:
        start_idx = generic_idx + 1

    if start_idx >= len(parts):
        start_idx = generic_idx
        generic_idx = -1

    out = list(parts)
    for i, part in enumerate(parts):
        part_trimmed = part.strip()
        if part_trimmed == "" or part == " " or part == "-" or is_quote(part):
            out[i] = part
            continue
        if is_frozen_org_part(part_trimmed, i, parts):
            out[i] = part
            continue
        if i == generic_idx:
            out[i] = part
            continue
        if i >= start_idx:
            if is_frozen_org(part_trimmed):
                out[i] = part
                continue
            part_gender = gender_by_ending(part_trimmed)
            rule = find_rule(part_trimmed, part_gender, org_rules, False)
            if rule is not None:
                out[i] = apply_rule(rule, part, word_case)
            else:
                if word_case == T.ACCUSATIVE and part_gender == T.MALE:
                    out[i] = part
                else:
                    res = incline_lastname(part, word_case, part_gender)
                    if res == "":
                        res = part
                    out[i] = res
        else:
            out[i] = part
    return "".join(out)


def org_in(name):
    return decline_org(name, T.PREPOSITIONAL)


def org_from(name):
    return decline_org(name, T.GENITIVE)


def org_to(name):
    return decline_org(name, T.ACCUSATIVE)


def org_decline(name, declension):
    dec = T.get_declension_const(declension)
    if dec == T.DECL_NULL:
        return name
    return decline_org(name, dec)
