"""Константы падежей и пола — соответствуют оригинальной библиотеке nodkz/lvovich."""

MALE = 1
FEMALE = 2
ANDROGYNOUS = 4
GENDER_NULL = 0

GENDER_STR = {
    MALE: "male",
    FEMALE: "female",
    ANDROGYNOUS: "androgynous",
    GENDER_NULL: "",
}

# Падежи: именительный, родительный, дательный, винительный, творительный, предложный.
NOMINATIVE = 1
GENITIVE = 2
DATIVE = 3
ACCUSATIVE = 4
INSTRUMENTAL = 5
PREPOSITIONAL = 6
DECL_NULL = 0

DECLENSION_STR = {
    "nominative": NOMINATIVE,
    "genitive": GENITIVE,
    "dative": DATIVE,
    "accusative": ACCUSATIVE,
    "instrumental": INSTRUMENTAL,
    "prepositional": PREPOSITIONAL,
}


def get_gender_const(key):
    """Переводит строку ('male') или число в константу пола."""
    if isinstance(key, str):
        return {"male": MALE, "female": FEMALE, "androgynous": ANDROGYNOUS}.get(key, GENDER_NULL)
    if isinstance(key, int):
        return key
    return GENDER_NULL


def convert_gender_str(cnst):
    """Переводит константу/строку в строку пола ('male' и т.д.)."""
    if isinstance(cnst, str):
        if cnst in ("male", "female", "androgynous"):
            return cnst
        return ""
    return GENDER_STR.get(cnst, "")


def get_declension_const(key):
    """Переводит строку ('nominative') или число в константу падежа."""
    if isinstance(key, str):
        return DECLENSION_STR.get(key, DECL_NULL)
    if isinstance(key, int):
        return key
    return DECL_NULL
