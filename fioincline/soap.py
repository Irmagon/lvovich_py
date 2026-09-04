"""SOAP-сервис (порт soap.go)."""

import xml.etree.ElementTree as ET

from .jsonx import j_stringify, Str, Null

_SOAP_OPS = {
    "Incline", "GetGender",
    "CityIn", "CityFrom", "CityTo",
    "OrgIn", "OrgFrom", "OrgTo",
}

_SOAP_ACTIONS = {
    "urn:Incline": "Incline", "urn:GetGender": "GetGender",
    "urn:CityIn": "CityIn", "urn:CityFrom": "CityFrom", "urn:CityTo": "CityTo",
    "urn:OrgIn": "OrgIn", "urn:OrgFrom": "OrgFrom", "urn:OrgTo": "OrgTo",
}

_RESP_PREFIX = (
    '<?xml version="1.0" encoding="utf-8"?>'
    '<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
    'xmlns:tns="urn:LvovichService"><soap:Body><tns:'
)

_FAULT_PREFIX = (
    '<?xml version="1.0" encoding="utf-8"?>'
    '<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" '
    'xmlns:tns="urn:LvovichService"><soap:Body><soap:Fault>'
    '<soap:Code><soap:Value>SOAP-ENV:Server</soap:Value>'
    '<soap:Subcode><soap:Value>InternalServerError</soap:Value></soap:Subcode></soap:Code>'
    '<soap:Reason><soap:Text>'
)


def _local_name(tag):
    if tag.startswith("{") and "}" in tag:
        return tag.split("}", 1)[1]
    return tag


def _find_body(elem):
    if _local_name(elem.tag) == "Body":
        return elem
    for c in elem:
        b = _find_body(c)
        if b is not None:
            return b
    return None


def parse_soap_doc(body):
    if not body.strip():
        return "", None, "empty body"
    try:
        root = ET.fromstring(body)
    except ET.ParseError as e:
        return "", None, str(e)
    body_el = _find_body(root)
    if body_el is None or len(body_el) == 0:
        return "", None, "no operation element"
    op_el = body_el[0]
    return _local_name(op_el.tag), op_el, None


def _get_string(node, name):
    if node is None:
        return ""
    for c in node:
        if _local_name(c.tag) == name:
            return (c.text or "").strip()
    return ""


def _get_string_checked(node, name):
    if node is None:
        return "", False
    for c in node:
        if _local_name(c.tag) == name:
            return (c.text or "").strip(), True
    return "", False


def _xml_escape(s):
    return (s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
            .replace('"', "&quot;"))


def _append_if(fields, name, value):
    if value == "":
        return fields
    fields.append((name, value))
    return fields


def _response(resp_el, fields):

    b = [_RESP_PREFIX, resp_el, ">"]
    for name, value in fields:
        b.append("<%s>%s</%s>" % (name, _xml_escape(value), name))
    b.append("</tns:%s></soap:Body></soap:Envelope>" % (resp_el))
    return "".join(b)


def _fault(reason):
    return (_FAULT_PREFIX + _xml_escape(reason) +
            "</soap:Text></soap:Reason></soap:Fault></soap:Body></soap:Envelope>")


def _stack500(text):
    return text + "\n    at Server._process (soap:server)\n    at Server._processRequestXml (soap:server)"


class SoapHandler:
    def __init__(self, core, log):
        self.core = core
        self.log = log

    def _log(self, op, args):
        if not self.log.is_enabled():
            return
        if args is None:
            args_json = "null"
        else:
            args_json = _args_json(args)
        self.log.log("", "%s - %s" % (op, args_json))

    def handle(self, body, soap_action):
        if not body.strip():
            return 500, _stack500("Error reading request body")
        op, op_el, err = parse_soap_doc(body)
        if err:
            return 500, _stack500(err)

        action = (soap_action or "").strip().strip('" ')
        err_kind = ""
        if action != "":
            expected = _SOAP_ACTIONS.get(action)
            if expected is None:
                err_kind = "style"
            elif expected != op:
                err_kind = "description"
        elif op not in _SOAP_OPS:
            err_kind = "description"

        if err_kind == "style":
            return 500, _fault("TypeError: Cannot read properties of undefined (reading 'style')")
        if err_kind == "description":
            return 500, _stack500("TypeError: Cannot read properties of undefined (reading 'description')")

        args = op_el if len(op_el) > 0 else None

        if op == "Incline":
            return self._incline(op_el, args)
        if op == "GetGender":
            return self._get_gender(op_el, args)
        if op in ("CityIn", "CityFrom", "CityTo"):
            return self._city(op_el, args, op)
        # OrgIn, OrgFrom, OrgTo
        return self._org(op_el, args, op)

    def _incline(self, op_el, args):
        self._log("SOAP Incline", args)
        if args is None:
            return 200, _response("InclineResponse", [("Fault", "Cannot read properties of null (reading 'LastName')")])
        last = _get_string(args, "LastName")
        first = _get_string(args, "FirstName")
        middle = _get_string(args, "MiddleName")
        decl = _get_string(args, "Declension")
        fmt = _get_string(args, "Format")
        if fmt == "":
            fmt = "full"
        from .lvovich import Person
        res = self.core.incline(Person(first, last, middle), decl, fmt)
        fields = []
        if fmt == "initials":
            fields = _append_if(fields, "LastName", res.sur_name)
            fields = _append_if(fields, "Initials", res.initials)
        else:
            fields = _append_if(fields, "LastName", res.sur_name)
            fields = _append_if(fields, "FirstName", res.first_name)
            fields = _append_if(fields, "MiddleName", res.second_name)
            fields = _append_if(fields, "Gender", res.gender)
        return 200, _response("InclineResponse", fields)

    def _get_gender(self, op_el, args):
        self._log("SOAP GetGender", args)
        if args is None:
            return 200, _response("GetGenderResponse", [("Fault", "Cannot read properties of null (reading 'LastName')")])
        last = _get_string(args, "LastName")
        first = _get_string(args, "FirstName")
        middle = _get_string(args, "MiddleName")
        from .lvovich import Person
        g = self.core.get_gender(Person(first, last, middle))
        return 200, _response("GetGenderResponse", _append_if([], "Gender", g))

    def _city(self, op_el, args, op):
        self._log("SOAP " + op, args)
        resp_el = op + "Response"
        if args is None:
            return 200, _response(resp_el, [("Fault", "Cannot read properties of null (reading 'Name')")])
        name, present = _get_string_checked(args, "Name")
        if not present:
            return 200, _response(resp_el, [("Fault", "Cannot read properties of undefined (reading 'toLowerCase')")])
        gender = _get_string(args, "Gender")
        if op == "CityIn":
            out = self.core.city_in(name, gender)
        elif op == "CityFrom":
            out = self.core.city_from(name, gender)
        else:
            out = self.core.city_to(name)
        return 200, _response(resp_el, _append_if([], "Name", out))

    def _org(self, op_el, args, op):
        self._log("SOAP " + op, args)
        resp_el = op + "Response"
        if args is None:
            return 200, _response(resp_el, [("Fault", "Cannot read properties of null (reading 'Name')")])
        name, present = _get_string_checked(args, "Name")
        if not present:
            return 200, _response(resp_el, [("Fault", "Cannot read properties of undefined (reading 'toLowerCase')")])
        if op == "OrgIn":
            out = self.core.org_in(name)
        elif op == "OrgFrom":
            out = self.core.org_from(name)
        else:
            out = self.core.org_to(name)
        return 200, _response(resp_el, _append_if([], "Name", out))


def _args_json(args):
    if args is None:
        return "null"
    first_keys = ["LastName", "FirstName", "MiddleName"]
    ordered = []
    for k in first_keys:
        for c in args:
            if _local_name(c.tag) == k:
                ordered.append(c)
                break
    rest = [c for c in args if _local_name(c.tag) not in first_keys]
    ordered.extend(rest)
    v = _elem_json(ordered)
    return j_stringify(v)


def _elem_json(children):
    from .jsonx import obj
    out = obj()
    for c in children:
        name = _local_name(c.tag)
        text = (c.text or "").strip()
        if len(c) == 0 and len(c.attrib) == 0:
            out.set(name, Str(text) if text != "" else Null())
        else:
            out.set(name, Str(text))
    return out
