// Склонение городов — порт city.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

import (
	"strings"
)

// splitCity — аналог JS name.split(/(\s|-)/g): разделители (' ' и '-')
// сохраняются в результате как отдельные элементы.
func splitCity(name string) []string {
	var parts []string
	var buf []rune
	flush := func() {
		parts = append(parts, string(buf))
		buf = buf[:0]
	}
	for _, r := range name {
		if r == ' ' || r == '-' {
			flush()
			parts = append(parts, string(r))
		} else {
			buf = append(buf, r)
		}
	}
	flush()
	return parts
}

func isFrozen(str string, words []string) bool {
	lower := strings.ToLower(str)
	for _, w := range words {
		if w == lower {
			return true
		}
	}
	return false
}

func isFrozenPart(part string, i int, parts []string) bool {
	if len(parts) > 1 {
		if isFrozen(part, frozenParts) {
			return true
		}
		for k := 0; k < i; k++ {
			if isFrozen(parts[k], frozenPartsAfter) {
				return true
			}
		}
	}
	return false
}

// declineTo — склонение названия города по частям (предложный/родительный).
func declineTo(name string, wordCase DeclenType, gender string) string {
	if isFrozen(name, frozenWords) {
		return name
	}
	parts := splitCity(name)
	out := make([]string, len(parts))
	for i, part := range parts {
		if isFrozenPart(part, i, parts) {
			out[i] = part
			continue
		}
		rule := FindRule(part, Androgynous, cityRules, false)
		if rule != nil {
			out[i] = applyRule(*rule, part, wordCase)
		} else {
			res := InclineFirstname(part, wordCase, GetGenderConst(gender))
			if res == "" {
				res = part
			}
			out[i] = res
		}
	}
	return strings.Join(out, "")
}

// CityIn — город в предложном падеже ("в Москве").
func CityIn(name string, gender string) string {
	return declineTo(name, Prepositional, gender)
}

// CityFrom — город в родительном падеже ("из Москвы").
func CityFrom(name string, gender string) string {
	return declineTo(name, Genitive, gender)
}

// CityTo — город в винительном падеже ("в Москву").
func CityTo(name string) string {
	if name == "" {
		return name
	}
	parts := splitCity(name)
	out := make([]string, len(parts))
	for i, part := range parts {
		if isFrozenPart(part, i, parts) {
			out[i] = part
			continue
		}
		partLower := strings.ToLower(part)
		switch {
		case endsWith(partLower, "а"):
			out[i] = applyMod(part, "-у")
		case endsWith(partLower, "ая"):
			out[i] = applyMod(part, "--ую")
		case endsWith(partLower, "ия"):
			out[i] = applyMod(part, "--ию")
		case endsWith(partLower, "я"):
			out[i] = applyMod(part, "-ю")
		default:
			out[i] = part
		}
	}
	return strings.Join(out, "")
}