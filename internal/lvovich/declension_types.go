// Константы падежей и их разбор — соответствуют оригинальной библиотеке
// nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

// DeclenType — числовое представление падежа.
type DeclenType int

// Падежи (константы — как в оригинальной библиотеке): именительный, родительный,
// дательный, винительный, творительный, предложный.
const (
	Nominative    DeclenType = 1
	Genitive      DeclenType = 2
	Dative        DeclenType = 3
	Accusative    DeclenType = 4
	Instrumental  DeclenType = 5
	Prepositional DeclenType = 6
)

// DeclenNull — отсутствие/неопределённость падежа.
const DeclenNull DeclenType = 0

// GetDeclensionConst переводит строку ('nominative') или число в константу падежа.
func GetDeclensionConst(key interface{}) DeclenType {
	switch v := key.(type) {
	case string:
		switch v {
		case "nominative":
			return Nominative
		case "genitive":
			return Genitive
		case "dative":
			return Dative
		case "accusative":
			return Accusative
		case "instrumental":
			return Instrumental
		case "prepositional":
			return Prepositional
		}
	case DeclenType:
		return v
	case int:
		return DeclenType(v)
	}
	return DeclenNull
}