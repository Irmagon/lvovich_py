// Определение пола — порт gender.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

import "strings"

// GenderRules — списки слов/суффиксов для каждого пола.
type GenderRules struct {
	Androgynous []string
	Male        []string
	Female      []string
}

// GenderRuleSet — набор исключений и суффиксов для определения пола.
type GenderRuleSet struct {
	Exceptions *GenderRules
	Suffixes   *GenderRules
}

// GetGenderConst переводит строку ('male') или число в константу пола.
func GetGenderConst(key interface{}) GenderConst {
	switch v := key.(type) {
	case string:
		switch v {
		case "male":
			return Male
		case "female":
			return Female
		case "androgynous":
			return Androgynous
		}
	case GenderConst:
		return v
	case int:
		return GenderConst(v)
	}
	return GenderNull
}

// ConvertGenderStr переводит строку/константу в строку пола ('male' и т.д.).
func ConvertGenderStr(cnst interface{}) string {
	switch v := cnst.(type) {
	case string:
		switch v {
		case "male", "female", "androgynous":
			return v
		}
	case GenderConst:
		switch v {
		case Male:
			return "male"
		case Female:
			return "female"
		case Androgynous:
			return "androgynous"
		}
	case int:
		switch GenderConst(v) {
		case Male:
			return "male"
		case Female:
			return "female"
		case Androgynous:
			return "androgynous"
		}
	}
	return ""
}

// Person — ФИО.
type Person struct {
	FirstName  string
	SurName    string
	SecondName string
}

// mergeGenders — объединение двух определений пола по правилам оригинальной библиотеки.
func mergeGenders(g1, g2 GenderConst) GenderConst {
	if g1 == Androgynous {
		return g2
	}
	if g2 == Androgynous {
		return g1
	}
	if g1 == g2 {
		return g1
	}
	return GenderNull
}

// getGenderByRule — пол, если ровно одна группа правил содержит совпадение.
func getGenderByRule(rules *GenderRules, matchFn func(string) bool) GenderConst {
	result := GenderNull
	groups := []struct {
		key   GenderConst
		words []string
	}{
		{Androgynous, rules.Androgynous},
		{Male, rules.Male},
		{Female, rules.Female},
	}
	found := 0
	for _, g := range groups {
		for _, w := range g.words {
			if matchFn(w) {
				found++
				result = g.key
				break
			}
		}
	}
	if found == 1 {
		return result
	}
	return GenderNull
}

// getGenderByRuleSet — определение пола по слову с учётом исключений и суффиксов.
func getGenderByRuleSet(name string, ruleSet *GenderRuleSet) GenderConst {
	if name == "" || ruleSet == nil {
		return GenderNull
	}
	nameLower := strings.ToLower(name)
	if ruleSet.Exceptions != nil {
		gender := getGenderByRule(ruleSet.Exceptions, func(some string) bool {
			if strings.HasPrefix(some, "-") {
				return endsWith(nameLower, some[1:])
			}
			return some == nameLower
		})
		if gender != GenderNull {
			return gender
		}
	}
	if ruleSet.Suffixes != nil {
		return getGenderByRule(ruleSet.Suffixes, func(some string) bool { return endsWith(nameLower, some) })
	}
	return GenderNull
}

// GetFG — пол по имени.
func GetFG(str string) GenderConst {
	return getGenderByRuleSet(str, genderRules().Firstname)
}

// GetLG — пол по фамилии.
func GetLG(str string) GenderConst {
	return getGenderByRuleSet(str, genderRules().Lastname)
}

// GetMG — пол по отчеству.
func GetMG(str string) GenderConst {
	return getGenderByRuleSet(str, genderRules().Middlename)
}

// GetFirstnameGender — пол по имени (строка).
func GetFirstnameGender(str string) string {
	return ConvertGenderStr(GetFG(str))
}

// GetLastnameGender — пол по фамилии (строка).
func GetLastnameGender(str string) string {
	return ConvertGenderStr(GetLG(str))
}

// GetMiddlenameGender — пол по отчеству (строка).
func GetMiddlenameGender(str string) string {
	return ConvertGenderStr(GetMG(str))
}

// GetGender — определение пола по ФИО (строка или "").
func GetGender(p Person) string {
	return ConvertGenderStr(getGenderConst(p))
}

// getGenderConst — определение пола по ФИО (константа).
func getGenderConst(p Person) GenderConst {
	result := GenderConst(Androgynous)

	if p.SecondName != "" {
		result = mergeGenders(result, GetMG(strings.TrimSpace(p.SecondName)))
	}
	if p.FirstName != "" {
		result = mergeGenders(result, GetFG(strings.TrimSpace(p.FirstName)))
	}
	if p.SurName != "" {
		lastGender := GetLG(strings.TrimSpace(p.SurName))
		if lastGender != GenderNull {
			result = mergeGenders(result, lastGender)
		}
	}
	return result
}