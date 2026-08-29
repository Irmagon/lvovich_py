// Склонение названий организаций — по аналогии с city.go.
package lvovich

import (
	"strings"
)

// splitOrg — разбивает название организации на токены, сохраняя разделители
// (пробел, дефис, кавычки). Кавычки всех видов (« », " ") нормализуются
// во внутреннее представление, но при сборке возвращаются в исходном виде.
func splitOrg(name string) (tokens []string, quotes []string) {
	type qt struct {
		open, close string
		normal      string
	}
	quoteTypes := []qt{
		{"\u00ab", "\u00bb", "\u00ab"}, // « »
		{"\u201c", "\u201d", "\u201c"}, // " "
		{"\"", "\"", "\""},            // " (ASCII)
		{"'", "'", "'"},               // ' (ASCII)
	}

	var parts []string
	var buf []rune
	flush := func() {
		if len(buf) > 0 {
			parts = append(parts, string(buf))
			buf = buf[:0]
		}
	}

	inQuote := ""
	quoteStack := []string{}

	for _, r := range name {
		ch := string(r)
		matched := false
		// открывающая кавычка
		for _, qt := range quoteTypes {
			if ch == qt.open && inQuote == "" {
				flush()
				parts = append(parts, ch)
				inQuote = qt.normal
				quoteStack = append(quoteStack, qt.close)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// закрывающая кавычка
		if len(quoteStack) > 0 && ch == quoteStack[len(quoteStack)-1] {
			flush()
			parts = append(parts, ch)
			quoteStack = quoteStack[:len(quoteStack)-1]
			if len(quoteStack) == 0 {
				inQuote = ""
			}
			matched = true
		}
		if matched {
			continue
		}
		if r == ' ' || r == '-' {
			flush()
			parts = append(parts, string(r))
		} else {
			buf = append(buf, r)
		}
	}
	flush()

	return parts, quoteStack
}

// isGenericWord — true, если слово — родовое (не склоняется).
func isGenericWord(str string) bool {
	lower := strings.ToLower(str)
	for _, w := range orgGenericWords {
		if w == lower {
			return true
		}
	}
	return false
}

// isFrozenOrg — true, если слово полностью несклоняемое.
func isFrozenOrg(str string) bool {
	lower := strings.ToLower(str)
	for _, w := range orgFrozenWords {
		if w == lower {
			return true
		}
	}
	return false
}

// isFrozenOrgPart — true, если часть названия не склоняется (предлог/союз).
func isFrozenOrgPart(part string, i int, parts []string) bool {
	if len(parts) > 1 {
		if isFrozen(part, orgFrozenParts) {
			return true
		}
		for k := 0; k < i; k++ {
			if isFrozen(parts[k], orgFrozenPartsAfter) {
				return true
			}
		}
	}
	return false
}

// genderByEnding — определяет грамматический род слова по окончанию
// (для выбора правила fallback при склонении значимой части названия).
func genderByEnding(word string) GenderConst {
	if word == "" {
		return Androgynous
	}
	lower := strings.ToLower(word)
	// окончания на гласную — женский/средний (не склоняются как мужской род)
	femaleEndings := []string{"а", "я"}
	for _, e := range femaleEndings {
		if strings.HasSuffix(lower, e) {
			return Female
		}
	}
	// окончания на согласную — мужской род
	vowels := "аеёиоуыэюя"
	last := lower[len(lower)-1]
	if !strings.ContainsRune(vowels, rune(last)) {
		return Male
	}
	// прочие гласные (о, е, и, у, ю, ы, э) — средний/несклоняемый
	return Androgynous
}

// declineOrg — склонение названия организации по падежу.
func declineOrg(name string, wordCase DeclenType) string {
	if name == "" {
		return name
	}

	// Проверка на полную заморозку
	if isFrozenOrg(name) {
		return name
	}

	parts, _ := splitOrg(name)
	if len(parts) == 0 {
		return name
	}

	// Найти родовое слово (первый не-разделитель, не-кавычка)
	genericIdx := -1
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == " " || part == "-" {
			continue
		}
		if isGenericWord(part) {
			genericIdx = i
			break
		}
	}

	// Если есть родовое слово — склоняем всё после него (значимую часть)
	// Если нет — склоняем всё название
	startIdx := 0
	if genericIdx >= 0 {
		startIdx = genericIdx + 1
	}

	// Если после родового слова ничего нет — склоняем родовое слово
	if startIdx >= len(parts) {
		startIdx = genericIdx
		genericIdx = -1
	}

	out := make([]string, len(parts))
	for i, part := range parts {
		partTrimmed := strings.TrimSpace(part)
		if partTrimmed == "" || part == " " || part == "-" || isQuote(part) {
			out[i] = part
			continue
		}
		// Разделители, предлоги, союзы — не склоняем
		if isFrozenOrgPart(partTrimmed, i, parts) {
			out[i] = part
			continue
		}
		// Родовое слово — не склоняем
		if i == genericIdx {
			out[i] = part
			continue
		}
		// Значимая часть — склоняем
		if i >= startIdx {
			// Проверка на замороженное слово
			if isFrozenOrg(partTrimmed) {
				out[i] = part
				continue
			}
			partGender := genderByEnding(partTrimmed)
			rule := FindRule(partTrimmed, partGender, orgRules, false)
			if rule != nil {
				out[i] = applyRule(*rule, part, wordCase)
			} else {
				// Для неодушевлённых мужского рода винительный = именительный
				if wordCase == Accusative && partGender == Male {
					out[i] = part
				} else {
					res := InclineLastname(part, wordCase, partGender)
					if res == "" {
						res = part
					}
					out[i] = res
				}
			}
		} else {
			out[i] = part
		}
	}
	return strings.Join(out, "")
}

// isQuote — true, если токен — кавычка.
func isQuote(s string) bool {
	switch s {
	case "\u00ab", "\u00bb", "\u201c", "\u201d", "\"", "'":
		return true
	}
	if len(s) == 1 {
		r := []rune(s)[0]
		return r == '\u00ab' || r == '\u00bb' || r == '\u201c' || r == '\u201d' || r == '"' || r == '\''
	}
	return false
}

// OrgIn — организация в предложном падеже ("в «Ромашке»").
func OrgIn(name string) string {
	return declineOrg(name, Prepositional)
}

// OrgFrom — организация в родительном падеже ("из «Ромашки»").
func OrgFrom(name string) string {
	return declineOrg(name, Genitive)
}

// OrgTo — организация в винительном падеже ("в «Ромашку»").
func OrgTo(name string) string {
	return declineOrg(name, Accusative)
}

// OrgDecline — склонение организации в указанный падеж.
// declension — строка или DeclenType.
func OrgDecline(name string, declension interface{}) string {
	dec := GetDeclensionConst(declension)
	if dec == DeclenNull {
		return name
	}
	return declineOrg(name, dec)
}