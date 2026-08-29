"""JSON-значение с сохранением порядка ключей и JS-экранированием (порт jsonx.go)."""

import json as _json

# Типы: obj, arr, str, num, bool, null, undef
J_OBJ = "o"
J_ARR = "a"
J_STR = "s"
J_NUM = "n"
J_BOOL = "b"
J_NULL = "z"
J_UNDEF = "u"


class JValue:
    __slots__ = ("type", "keys", "vals", "str", "raw", "items")

    def __init__(self, type_=J_UNDEF):
        self.type = type_
        self.keys = []      # порядок ключей объекта
        self.vals = {}      # dict ключ->JValue
        self.str = ""       # для строк
        self.raw = ""       # для num/bool
        self.items = []     # для массивов

    def set(self, k, val):
        self.keys.append(k)
        self.vals[k] = val

    def get(self, k):
        if self.type != J_OBJ:
            return None
        return self.vals.get(k)

    def is_string(self):
        return self.type == J_STR

    def str_val(self):
        if self.type == J_STR:
            return self.str, True
        return "", False


def obj():
    return JValue(J_OBJ)


def arr():
    return JValue(J_ARR)


def Str(s):
    v = JValue(J_STR)
    v.str = s
    return v


def Num(raw):
    v = JValue(J_NUM)
    v.raw = raw
    return v


def BoolVal(raw):
    v = JValue(J_BOOL)
    v.raw = raw
    return v


def Null():
    return JValue(J_NULL)


def Undef():
    return JValue(J_UNDEF)


def _escape(b, s):
    for r in s:
        if r == '"':
            b.append('\\"')
        elif r == "\\":
            b.append("\\\\")
        elif r == "\b":
            b.append("\\b")
        elif r == "\f":
            b.append("\\f")
        elif r == "\n":
            b.append("\\n")
        elif r == "\r":
            b.append("\\r")
        elif r == "\t":
            b.append("\\t")
        elif ord(r) < 0x20:
            b.append("\\u%04x" % ord(r))
        else:
            b.append(r)


def _write(b, v):
    if v is None or v.type in (J_NULL, J_UNDEF):
        b.append("null")
        return
    t = v.type
    if t == J_STR:
        b.append('"')
        _escape(b, v.str)
        b.append('"')
    elif t in (J_NUM, J_BOOL):
        b.append(v.raw)
    elif t == J_ARR:
        b.append("[")
        for i, it in enumerate(v.items):
            if i > 0:
                b.append(",")
            _write(b, it)
        b.append("]")
    elif t == J_OBJ:
        b.append("{")
        first = True
        for k in v.keys:
            val = v.vals[k]
            if val is not None and val.type == J_UNDEF:
                continue
            if not first:
                b.append(",")
            first = False
            b.append('"')
            _escape(b, k)
            b.append('":')
            _write(b, val)
        b.append("}")


def j_stringify(v):
    b = []
    _write(b, v)
    return "".join(b)


class _Parser:
    def __init__(self, s):
        self.s = s
        self.pos = 0

    def skip_ws(self):
        while self.pos < len(self.s) and self.s[self.pos] in " \t\n\r":
            self.pos += 1

    def parse_value(self):
        self.skip_ws()
        if self.pos >= len(self.s):
            raise ValueError("unexpected end")
        c = self.s[self.pos]
        if c == "{":
            return self.parse_object()
        if c == "[":
            return self.parse_array()
        if c == '"':
            return Str(self.parse_string())
        if self.s.startswith("true", self.pos):
            self.pos += 4
            return BoolVal("true")
        if self.s.startswith("false", self.pos):
            self.pos += 5
            return BoolVal("false")
        if self.s.startswith("null", self.pos):
            self.pos += 4
            return Null()
        # число
        start = self.pos
        while self.pos < len(self.s) and self.s[self.pos] in "0123456789-+.eE":
            self.pos += 1
        if self.pos > start:
            return Num(self.s[start:self.pos])
        raise ValueError(f"unexpected char at {self.pos}")

    def parse_object(self):
        v = obj()
        self.pos += 1  # '{'
        while True:
            self.skip_ws()
            if self.pos >= len(self.s):
                raise ValueError("unexpected end of object")
            if self.s[self.pos] == "}":
                self.pos += 1
                return v
            if self.s[self.pos] != '"':
                raise ValueError("expected string key")
            k = self.parse_string()
            self.skip_ws()
            if self.pos >= len(self.s) or self.s[self.pos] != ":":
                raise ValueError("expected ':'")
            self.pos += 1
            val = self.parse_value()
            v.set(k, val)
            self.skip_ws()
            if self.pos >= len(self.s):
                raise ValueError("unexpected end of object")
            if self.s[self.pos] == ",":
                self.pos += 1
                continue
            if self.s[self.pos] == "}":
                self.pos += 1
                return v
            raise ValueError("expected ',' or '}'")

    def parse_array(self):
        v = arr()
        self.pos += 1  # '['
        while True:
            self.skip_ws()
            if self.pos >= len(self.s):
                raise ValueError("unexpected end of array")
            if self.s[self.pos] == "]":
                self.pos += 1
                return v
            item = self.parse_value()
            v.items.append(item)
            self.skip_ws()
            if self.pos >= len(self.s):
                raise ValueError("unexpected end of array")
            if self.s[self.pos] == ",":
                self.pos += 1
                continue
            if self.s[self.pos] == "]":
                self.pos += 1
                return v
            raise ValueError("expected ',' or ']'")

    def parse_string(self):
        if self.pos >= len(self.s) or self.s[self.pos] != '"':
            raise ValueError("expected string")
        self.pos += 1
        out = []
        while self.pos < len(self.s):
            c = self.s[self.pos]
            if c == '"':
                self.pos += 1
                return "".join(out)
            if c == "\\":
                self.pos += 1
                if self.pos >= len(self.s):
                    raise ValueError("bad escape")
                e = self.s[self.pos]
                self.pos += 1
                if e == '"':
                    out.append('"')
                elif e == "\\":
                    out.append("\\")
                elif e == "/":
                    out.append("/")
                elif e == "b":
                    out.append("\b")
                elif e == "f":
                    out.append("\f")
                elif e == "n":
                    out.append("\n")
                elif e == "r":
                    out.append("\r")
                elif e == "t":
                    out.append("\t")
                elif e == "u":
                    hexstr = self.s[self.pos:self.pos + 4]
                    if len(hexstr) != 4:
                        raise ValueError("bad unicode escape")
                    try:
                        out.append(chr(int(hexstr, 16)))
                    except ValueError:
                        raise ValueError("bad unicode escape")
                    self.pos += 4
                else:
                    raise ValueError(f"bad escape {e}")
                continue
            out.append(c)
            self.pos += 1
        raise ValueError("unterminated string")


def parse_json(s):
    return _Parser(s).parse_value()


# Парсинг с сохранением порядка через стандартный json -> JValue (обратная совместимость).
def from_py(data):
    """Переводит Python-объект (dict/list/...) в JValue с сохранением порядка."""
    if data is None:
        return Null()
    if isinstance(data, dict):
        v = obj()
        for k, val in data.items():
            v.set(str(k), from_py(val))
        return v
    if isinstance(data, list):
        v = arr()
        v.items = [from_py(x) for x in data]
        return v
    if isinstance(data, bool):
        return BoolVal("true" if data else "false")
    if isinstance(data, (int, float)):
        return Num(_json.dumps(data))
    return Str(str(data))
