// Движок склонения по правилам — порт inclineRules.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

import (
	"strings"
	"unicode/utf8"
)

// DeclensionRule — правило склонения (суффикс или исключение).
type DeclensionRule struct {
	// Gender — ожидаемый пол: Male, Female, Androgynous.
	// Androgynous означает "подходит для любого пола" (как в оригинале).
	Gender GenderConst
	Test   []string
	// Mods — 5 модификаторов: родительный, дательный, винительный,
	// творительный, предложный.
	Mods []string
	Tags []string
}

// RuleSet — набор исключений и правил по суффиксам.
type RuleSet struct {
	Exceptions []DeclensionRule
	Suffixes   []DeclensionRule
}

// applyMod — применяет модификатор к слову (как в оригинале):
// '.' — ничего, '-' — удалить последний символ, прочий символ — добавить.
func applyMod(str, mod string) string {
	out := str
	for _, r := range mod {
		switch r {
		case '.':
			// без изменений
		case '-':
			out = trimLastRune(out)
		default:
			out += string(r)
		}
	}
	return out
}

// trimLastRune удаляет последний символ (руну) строки.
func trimLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

// getModByIdx — модификатор по индексу падежа, иначе '.'.
func getModByIdx(mods []string, i int) string {
	if len(mods) > 0 && len(mods) >= i+1 {
		return mods[i]
	}
	return "."
}

// applyRule — применение правила к слову в указанном падеже.
func applyRule(rule DeclensionRule, str string, declension DeclenType) string {
	var mod string
	switch declension {
	case Nominative:
		mod = "."
	case Genitive:
		mod = getModByIdx(rule.Mods, 0)
	case Dative:
		mod = getModByIdx(rule.Mods, 1)
	case Accusative:
		mod = getModByIdx(rule.Mods, 2)
	case Instrumental:
		mod = getModByIdx(rule.Mods, 3)
	case Prepositional:
		mod = getModByIdx(rule.Mods, 4)
	default:
		mod = "."
	}
	return applyMod(str, mod)
}

// findExactRule — поиск правила среди переданного списка.
// Совпадает правило с тем же полом (или androgynous) И совпадающим тестом.
// Правила с tags пропускаются, если ни один tag не запрошен.
func findExactRule(rules []DeclensionRule, gender GenderConst, matchFn func(string) bool, tags []string) *DeclensionRule {
	for i := range rules {
		rule := &rules[i]

		if len(rule.Tags) > 0 {
			found := false
			for _, t := range rule.Tags {
				for _, rt := range tags {
					if t == rt {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		if rule.Gender != Androgynous && gender != rule.Gender {
			continue
		}

		for _, t := range rule.Test {
			if matchFn(t) {
				return rule
			}
		}
	}
	return nil
}

// FindRule — поиск правила для слова: сначала исключения, затем суффиксы.
func FindRule(str string, gender GenderConst, ruleSet *RuleSet, firstWord bool) *DeclensionRule {
	if str == "" {
		return nil
	}
	strLower := strings.ToLower(str)

	tags := []string{}
	if firstWord {
		tags = append(tags, "firstWord")
	}

	if len(ruleSet.Exceptions) > 0 {
		rule := findExactRule(ruleSet.Exceptions, gender, func(some string) bool { return some == strLower }, tags)
		if rule != nil {
			return rule
		}
	}

	if len(ruleSet.Suffixes) > 0 {
		return findExactRule(ruleSet.Suffixes, gender, func(some string) bool { return endsWith(strLower, some) }, tags)
	}
	return nil
}

// inclineByRules — склонение слова по правилам с учётом пола.
func inclineByRules(str string, declension interface{}, gender GenderConst, ruleSet *RuleSet) string {
	dec := GetDeclensionConst(declension)
	g := GetGenderConst(gender)

	parts := strings.Split(str, "-")
	result := make([]string, 0, len(parts))
	for i, part := range parts {
		isFirstWord := i == 0 && len(parts) > 1
		rule := FindRule(part, g, ruleSet, isFirstWord)
		if rule != nil {
			result = append(result, applyRule(*rule, part, dec))
		} else {
			result = append(result, part)
		}
	}
	return strings.Join(result, "-")
}
