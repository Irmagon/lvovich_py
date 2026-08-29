package server

import "fioincline/internal/lvovich"

// Person — ФИО (поля как в исходной библиотеке).
type Person = lvovich.Person

// InclineResult — результат склонения.
type InclineResult = lvovich.InclineOut

// Core — тонкая обёртка над ядром lvovich.
type Core struct{}

// NewCore создаёт ядро.
func NewCore() *Core { return &Core{} }

// Incline склоняет ФИО. decl="" трактуется как "не указано" (по умолчанию винительный).
func (c *Core) Incline(p *Person, decl, format string) InclineResult {
	return lvovich.Incline(*p, decl, format)
}

// GetGender определяет пол ("male"/"female"/"").
func (c *Core) GetGender(p *Person) string {
	return lvovich.GetGender(*p)
}

// CityIn — город в предложном падеже.
func (c *Core) CityIn(name, gender string) string {
	return lvovich.CityIn(name, gender)
}

// CityFrom — город в родительном падеже.
func (c *Core) CityFrom(name, gender string) string {
	return lvovich.CityFrom(name, gender)
}

// CityTo — город в винительном падеже.
func (c *Core) CityTo(name string) string {
	return lvovich.CityTo(name)
}

// OrgIn — организация в предложном падеже.
func (c *Core) OrgIn(name string) string {
	return lvovich.OrgIn(name)
}

// OrgFrom — организация в родительном падеже.
func (c *Core) OrgFrom(name string) string {
	return lvovich.OrgFrom(name)
}

// OrgTo — организация в винительном падеже.
func (c *Core) OrgTo(name string) string {
	return lvovich.OrgTo(name)
}
