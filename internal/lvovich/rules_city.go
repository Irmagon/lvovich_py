package lvovich

// Правила склонения городов — порт cityRules.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
// Все правила без пола (Androgynous — подходит для любого пола).
// Порядок правил важен: первое совпадение выигрывает.

var frozenWords = []string{"форт-шевченко"}

var frozenParts = []string{
	"-", " ", "в", "на", "баден", "бледно", "буэнос", "вице", "гаврилов", "йошкар",
	"коста", "лос", "норд", "нью", "орехово", "принс", "сан", "санкт", "санта", "северо",
	"ситтард", "темно", "улан", "усолье", "усть", "форт", "царь", "экс", "юго", "юрьев",
	"нур", "соль",
}

// frozenPartsAfter — слова, после которых части больше не склоняются.
var frozenPartsAfter = []string{"село", "поселок", "аул", "город", "деревня", "урочище"}

var cityRules = &RuleSet{
	Exceptions: []DeclensionRule{
		{Test: []string{"сочи", "тбилиси", "хельсинки"}, Mods: []string{"", "", "", "", ""}},
		{Test: []string{"село", "озеро", "место"}, Mods: []string{"-а", "-у", "", "м", "-е"}},
		{Test: []string{"область"}, Mods: []string{"-и", "-и", "", "ю", "-и"}},
		{Test: []string{"деревня"}, Mods: []string{"-и", "-е", "-ю", "-ей", "-е"}},
		{Test: []string{"море"}, Mods: []string{"-я", "-ю", "", "м", ""}},
		{Test: []string{"холм"}, Mods: []string{"а", "у", "", "ом", "е"}},
		{Test: []string{"орел", "орёл"}, Mods: []string{"--ла", "--лу", "--ла", "--лом", "--ле"}},
		{Test: []string{"крым"}, Mods: []string{"-ма", "-му", "-ма", "-ом", "-му"}},
		{Test: []string{"бор"}, Mods: []string{"а", "у", "", "ом", "у"}},
	},
	Suffixes: []DeclensionRule{
		{Test: []string{"чёк", "чек"}, Mods: []string{"--ка", "--ку", "", "--ком", "--ке"}},
		{Test: []string{"чик", "ич"}, Mods: []string{"а", "у", "", "ом", "е"}},
		{Test: []string{"жний", "хний", "шний", "щий"}, Mods: []string{"--его", "--ему", "", "-м", "--ем"}},
		{Test: []string{"ще"}, Mods: []string{"-а", "-у", "", "м", ""}},
		{Test: []string{"щи"}, Mods: []string{"-", "-ам", "", "-ами", "-ах"}},
		{Test: []string{"чье"}, Mods: []string{"-я", "-ю", "", "м", ""}},
		{Test: []string{"ель", "пль"}, Mods: []string{"-я", "-ю", "", "-ем", "-е"}},
		{Test: []string{"чь"}, Mods: []string{"-и", "-и", "", "ю", "-и"}},
		{Test: []string{"чи"}, Mods: []string{"-ей", "-ам", "", "-ами", "-ах"}},
		{Test: []string{"ые", "ие"}, Mods: []string{"-х", "-м", "", "-ми", "-х"}},
		{Test: []string{"ый", "ий", "ое"}, Mods: []string{"--ого", "--ому", "", "-м", "--ом"}},
		{Test: []string{"ая"}, Mods: []string{"--ой", "--ой", "--ую", "--ой", "--ой"}},
		{Test: []string{"гиев"}, Mods: []string{"а", "у", "", "ым", "ом"}},
		{Test: []string{"ны", "вцы"}, Mods: []string{"-ов", "-ам", "", "-ами", "-ах"}},
		{Test: []string{"ша"}, Mods: []string{"-и", "-е", "-у", "-ей", "-е"}},
		{Test: []string{"ры", "цы", "ды", "ги"}, Mods: []string{"-", "-ам", "", "-ами", "-ах"}},
		{Test: []string{"амень"}, Mods: []string{"---ня", "---ню", "", "---нем", "---не"}},
		{Test: []string{"ьн", "нц", "мм"}, Mods: []string{"а", "у", "", "ом", "е"}},
	},
}

// init — как constantizeGenderInRules в оригинале: все правила городов androgynous.
func init() {
	for i := range cityRules.Exceptions {
		cityRules.Exceptions[i].Gender = Androgynous
	}
	for i := range cityRules.Suffixes {
		cityRules.Suffixes[i].Gender = Androgynous
	}
}
