package lvovich

// Правила склонения отчеств — порт inclineRulesMiddlename.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
// Порядок правил важен: первое совпадение выигрывает.

var middlenameRules = &RuleSet{
	Exceptions: []DeclensionRule{
		{Gender: Androgynous, Test: []string{"борух"}, Mods: []string{".", ".", ".", ".", "."}, Tags: []string{"first_word"}},
	},
	Suffixes: []DeclensionRule{
		{Gender: Male, Test: []string{"мич", "ьич", "кич"}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Male, Test: []string{"ич"}, Mods: []string{"а", "у", "а", "ем", "е"}},
		{Gender: Female, Test: []string{"на"}, Mods: []string{"-ы", "-е", "-у", "-ой", "-е"}},
	},
}