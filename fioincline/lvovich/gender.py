"""Определение пола — порт gender.ts / gender.go (nodkz/lvovich)."""

from . import types as T


class GenderRules:
    __slots__ = ("androgynous", "male", "female")

    def __init__(self, androgynous=None, male=None, female=None):
        self.androgynous = androgynous or []
        self.male = male or []
        self.female = female or []


class GenderRuleSet:
    __slots__ = ("exceptions", "suffixes")

    def __init__(self, exceptions=None, suffixes=None):
        self.exceptions = exceptions
        self.suffixes = suffixes


class Person:
    __slots__ = ("first_name", "sur_name", "second_name")

    def __init__(self, first_name="", sur_name="", second_name=""):
        self.first_name = first_name
        self.sur_name = sur_name
        self.second_name = second_name


def merge_genders(g1, g2):
    if g1 == T.ANDROGYNOUS:
        return g2
    if g2 == T.ANDROGYNOUS:
        return g1
    if g1 == g2:
        return g1
    return T.GENDER_NULL


def get_gender_by_rule(rules, match_fn):
    result = T.GENDER_NULL
    groups = [
        (T.ANDROGYNOUS, rules.androgynous),
        (T.MALE, rules.male),
        (T.FEMALE, rules.female),
    ]
    found = 0
    for key, words in groups:
        for w in words:
            if match_fn(w):
                found += 1
                result = key
                break
    if found == 1:
        return result
    return T.GENDER_NULL


def get_gender_by_rule_set(name, rule_set):
    if name == "" or rule_set is None:
        return T.GENDER_NULL
    name_lower = name.lower()
    if rule_set.exceptions is not None:
        def exc_match(some):
            if some.startswith("-"):
                return name_lower.endswith(some[1:])
            return some == name_lower
        gender = get_gender_by_rule(rule_set.exceptions, exc_match)
        if gender != T.GENDER_NULL:
            return gender
    if rule_set.suffixes is not None:
        return get_gender_by_rule(rule_set.suffixes, lambda some: name_lower.endswith(some))
    return T.GENDER_NULL


def get_fg(str_):
    return get_gender_by_rule_set(str_, gender_rules().firstname)


def get_lg(str_):
    return get_gender_by_rule_set(str_, gender_rules().lastname)


def get_mg(str_):
    return get_gender_by_rule_set(str_, gender_rules().middlename)


def get_gender_const(p):
    result = T.ANDROGYNOUS
    if p.second_name != "":
        result = merge_genders(result, get_mg(p.second_name.strip()))
    if p.first_name != "":
        result = merge_genders(result, get_fg(p.first_name.strip()))
    if p.sur_name != "":
        last_gender = get_lg(p.sur_name.strip())
        if last_gender != T.GENDER_NULL:
            result = merge_genders(result, last_gender)
    return result


def get_gender(p):
    return T.convert_gender_str(get_gender_const(p))


from .rules_data import gender_rules  # noqa: E402
