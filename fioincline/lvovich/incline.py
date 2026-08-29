"""Склонение ФИО — порт index.ts / incline.go (nodkz/lvovich)."""

from . import types as T
from .rules import incline_by_rules
from .gender import Person, get_gender_const
from .rules_data import firstname_rules, lastname_rules, middlename_rules


def resolve_declension(declension):
    if declension is None or declension == "":
        return T.ACCUSATIVE
    return T.get_declension_const(declension)


def incline_firstname(str_, declension, gender):
    if gender == T.GENDER_NULL:
        from .gender import get_fg
        gender = get_fg(str_)
    return incline_by_rules(str_, resolve_declension(declension), gender, firstname_rules)


def incline_lastname(str_, declension, gender):
    if gender == T.GENDER_NULL:
        from .gender import get_lg
        gender = get_lg(str_)
    return incline_by_rules(str_, resolve_declension(declension), gender, lastname_rules)


def incline_middlename(str_, declension, gender):
    if gender == T.GENDER_NULL:
        from .gender import get_mg
        gender = get_mg(str_)
    return incline_by_rules(str_, resolve_declension(declension), gender, middlename_rules)


class InclineOut:
    __slots__ = ("first_name", "sur_name", "second_name", "gender", "initials")

    def __init__(self):
        self.first_name = ""
        self.sur_name = ""
        self.second_name = ""
        self.gender = ""
        self.initials = ""


def initials(p):
    res = ""
    if p.first_name != "":
        first = p.first_name.strip()
        if first != "":
            res += first[0].upper() + "."
    if p.second_name != "":
        second = p.second_name.strip()
        if second != "":
            res += second[0].upper() + "."
    return res


def incline(p, declension, format_=""):
    from .gender import get_gender
    gender = get_gender(p)
    res = InclineOut()
    res.gender = gender
    gc = get_gender_const(p)

    if p.first_name != "":
        res.first_name = incline_firstname(p.first_name.strip(), declension, gc)
    if p.sur_name != "":
        res.sur_name = incline_lastname(p.sur_name.strip(), declension, gc)
    if p.second_name != "":
        res.second_name = incline_middlename(p.second_name.strip(), declension, gc)

    if format_ == "initials":
        res.initials = initials(p)
    return res
