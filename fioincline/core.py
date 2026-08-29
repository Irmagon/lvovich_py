"""Тонкая обёртка над ядром lvovich (порт core.go)."""

from .lvovich import (
    Person, get_gender, incline, city_in, city_from, city_to, org_in, org_from, org_to,
)


class Core:
    def incline(self, p, decl, format_):
        return incline(p, decl, format_)

    def get_gender(self, p):
        return get_gender(p)

    def city_in(self, name, gender):
        return city_in(name, gender)

    def city_from(self, name, gender):
        return city_from(name, gender)

    def city_to(self, name):
        return city_to(name)

    def org_in(self, name):
        return org_in(name)

    def org_from(self, name):
        return org_from(name)

    def org_to(self, name):
        return org_to(name)
