// Склонение ФИО — порт index.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// resolveDeclension — приведение переданного падежа к константе.
// undefined/пустое значение трактуется как винительный (по умолчанию оригинала).
func resolveDeclension(declension interface{}) DeclenType {
	if declension == nil {
		return Accusative
	}
	switch v := declension.(type) {
	case string:
		if v == "" {
			return Accusative
		}
	case DeclenType:
		if v == DeclenNull {
			return Accusative
		}
	}
	return GetDeclensionConst(declension)
}

// InclineFirstname — склонение имени.
func InclineFirstname(str string, declension interface{}, gender GenderConst) string {
	if gender == GenderNull {
		gender = GetFG(str)
	}
	return inclineByRules(str, resolveDeclension(declension), gender, firstnameRules)
}

// InclineLastname — склонение фамилии.
func InclineLastname(str string, declension interface{}, gender GenderConst) string {
	if gender == GenderNull {
		gender = GetLG(str)
	}
	return inclineByRules(str, resolveDeclension(declension), gender, lastnameRules)
}

// InclineMiddlename — склонение отчества.
func InclineMiddlename(str string, declension interface{}, gender GenderConst) string {
	if gender == GenderNull {
		gender = GetMG(str)
	}
	return inclineByRules(str, resolveDeclension(declension), gender, middlenameRules)
}

// InclineOut — результат склонения ФИО.
type InclineOut struct {
	FirstName  string
	SurName    string
	SecondName string
	Gender     string
	// Initials заполняется только в формате initials.
	Initials string
}

// initials — инициалы вида "И.И." (первая буква имени/отчества с точкой).
func initials(p Person) string {
	res := ""
	if p.FirstName != "" {
		first := strings.TrimSpace(p.FirstName)
		if first != "" {
			r, _ := utf8.DecodeRuneInString(first)
			res += string(unicode.ToUpper(r)) + "."
		}
	}
	if p.SecondName != "" {
		second := strings.TrimSpace(p.SecondName)
		if second != "" {
			r, _ := utf8.DecodeRuneInString(second)
			res += string(unicode.ToUpper(r)) + "."
		}
	}
	return res
}

// Incline — склонение ФИО по падежам.
// format: "full" (по умолчанию) или "initials".
func Incline(p Person, declension interface{}, format string) InclineOut {
	gender := GetGender(p)
	res := InclineOut{Gender: gender}
	gc := GetGenderConst(gender)

	if p.FirstName != "" {
		res.FirstName = InclineFirstname(strings.TrimSpace(p.FirstName), declension, gc)
	}
	if p.SurName != "" {
		res.SurName = InclineLastname(strings.TrimSpace(p.SurName), declension, gc)
	}
	if p.SecondName != "" {
		res.SecondName = InclineMiddlename(strings.TrimSpace(p.SecondName), declension, gc)
	}

	if format == "initials" {
		res.Initials = initials(p)
	}
	return res
}