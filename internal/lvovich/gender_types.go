// Константы пола и их разбор — соответствуют оригинальной библиотеке
// nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
package lvovich

// GenderConst — числовое представление пола.
type GenderConst int

// Пол (константы — как в оригинальной библиотеке).
const (
	Male        GenderConst = 1
	Female      GenderConst = 2
	Androgynous GenderConst = 4
)

// GenderNull — отсутствие/неопределённость пола.
const GenderNull GenderConst = 0

// String для диагностики.
func (g GenderConst) String() string {
	switch g {
	case Male:
		return "male"
	case Female:
		return "female"
	case Androgynous:
		return "androgynous"
	default:
		return ""
	}
}
