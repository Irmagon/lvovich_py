"""Движок склонения по правилам — порт inclineRules.ts / rules.go (nodkz/lvovich)."""

from . import types as T


class DeclensionRule:
    __slots__ = ("gender", "test", "mods", "tags")

    def __init__(self, gender, test, mods, tags=None):
        self.gender = gender
        self.test = test
        self.mods = mods
        self.tags = tags or []


class RuleSet:
    __slots__ = ("exceptions", "suffixes")

    def __init__(self, exceptions=None, suffixes=None):
        self.exceptions = exceptions or []
        self.suffixes = suffixes or []


def trim_last_rune(s):
    """Удаляет последний символ строки."""
    if not s:
        return s
    return s[:-1]


def apply_mod(str_, mod):
    """Применяет модификатор к слову: '.' — ничего, '-' — удалить последний символ,
    прочий символ — добавить."""
    out = str_
    for r in mod:
        if r == ".":
            continue
        elif r == "-":
            out = trim_last_rune(out)
        else:
            out += r
    return out


def get_mod_by_idx(mods, i):
    if mods and len(mods) >= i + 1:
        return mods[i]
    return "."


def apply_rule(rule, str_, declension):
    if declension == T.NOMINATIVE:
        mod = "."
    elif declension == T.GENITIVE:
        mod = get_mod_by_idx(rule.mods, 0)
    elif declension == T.DATIVE:
        mod = get_mod_by_idx(rule.mods, 1)
    elif declension == T.ACCUSATIVE:
        mod = get_mod_by_idx(rule.mods, 2)
    elif declension == T.INSTRUMENTAL:
        mod = get_mod_by_idx(rule.mods, 3)
    elif declension == T.PREPOSITIONAL:
        mod = get_mod_by_idx(rule.mods, 4)
    else:
        mod = "."
    return apply_mod(str_, mod)


def ends_with(str_, search):
    return str_.endswith(search)


def starts_with(str_, search, pos=0):
    return str_.startswith(search, pos)


def find_exact_rule(rules, gender, match_fn, tags):
    for rule in rules:
        if rule.tags:
            found = False
            for t in rule.tags:
                if t in tags:
                    found = True
                    break
            if not found:
                continue
        if rule.gender != T.ANDROGYNOUS and gender != rule.gender:
            continue
        for t in rule.test:
            if match_fn(t):
                return rule
    return None


def find_rule(str_, gender, rule_set, first_word=False):
    if str_ == "":
        return None
    str_lower = str_.lower()

    tags = []
    if first_word:
        tags.append("firstWord")

    if rule_set.exceptions:
        rule = find_exact_rule(rule_set.exceptions, gender, lambda some: some == str_lower, tags)
        if rule is not None:
            return rule

    if rule_set.suffixes:
        return find_exact_rule(rule_set.suffixes, gender, lambda some: ends_with(str_lower, some), tags)
    return None


def incline_by_rules(str_, declension, gender, rule_set):
    dec = T.get_declension_const(declension)
    g = T.get_gender_const(gender)

    parts = str_.split("-")
    result = []
    for i, part in enumerate(parts):
        is_first_word = i == 0 and len(parts) > 1
        rule = find_rule(part, g, rule_set, is_first_word)
        if rule is not None:
            result.append(apply_rule(rule, part, dec))
        else:
            result.append(part)
    return "-".join(result)
