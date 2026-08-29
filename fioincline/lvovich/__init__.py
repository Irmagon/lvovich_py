"""Ядро склонения русских ФИО, городов и организаций (порт nodkz/lvovich)."""

from . import types as T
from .rules import (
    RuleSet, DeclensionRule, apply_mod, apply_rule, find_rule, incline_by_rules,
)
from .gender import Person, get_gender, get_gender_const, get_fg, get_lg, get_mg
from .incline import incline, InclineOut
from .city import city_in, city_from, city_to
from .org import org_in, org_from, org_to, org_decline
from . import rules_data

__all__ = [
    "T",
    "Person", "get_gender", "get_gender_const", "get_fg", "get_lg", "get_mg",
    "incline", "InclineOut",
    "city_in", "city_from", "city_to",
    "org_in", "org_from", "org_to", "org_decline",
    "rules_data",
]
