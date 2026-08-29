package lvovich

// Правила склонения имён — порт inclineRulesFirstname.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
// Порядок правил важен: первое совпадение выигрывает.

var firstnameRules = &RuleSet{
	Exceptions: []DeclensionRule{
		{Gender: Male, Test: []string{"лев"}, Mods: []string{"--ьва", "--ьву", "--ьва", "--ьвом", "--ьве"}},
		{Gender: Male, Test: []string{"пётр"}, Mods: []string{"---етра", "---етру", "---етра", "---етром", "---етре"}},
		{Gender: Male, Test: []string{"павел"}, Mods: []string{"--ла", "--лу", "--ла", "--лом", "--ле"}},
		{Gender: Male, Test: []string{"яша"}, Mods: []string{"-и", "-е", "-у", "-ей", "-е"}},
		{Gender: Male, Test: []string{"шота"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Female, Test: []string{"агидель", "жизель", "нинель", "рашель", "рахиль"}, Mods: []string{"-и", "-и", ".", "ю", "-и"}},
	},
	Suffixes: []DeclensionRule{
		{Gender: Androgynous, Test: []string{"ки"}, Mods: []string{"-ов", "-ам", ".", "-ами", "-ах"}},
		{Gender: Androgynous, Test: []string{"щи"}, Mods: []string{"-ев", "-ам", ".", "-ами", "-ах"}},
		{Gender: Androgynous, Test: []string{"е", "ё", "и", "о", "у", "ы", "э", "ю"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Male, Test: []string{"уа", "иа"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Female, Test: []string{
			"б", "в", "г", "д", "ж", "з", "й", "к", "л", "м",
			"н", "п", "р", "с", "т", "ф", "х", "ц", "ч", "ш",
			"щ", "ъ", "иа", "ль",
		}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Female, Test: []string{"ь"}, Mods: []string{"-и", "-и", ".", "ю", "-и"}},
		{Gender: Male, Test: []string{"ь"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Androgynous, Test: []string{"га", "ка", "ха", "ча", "ща", "жа"}, Mods: []string{"-и", "-е", "-у", "-ой", "-е"}},
		{Gender: Female, Test: []string{"ша"}, Mods: []string{"-и", "-е", "-у", "-ей", "-е"}},
		{Gender: Male, Test: []string{"ша", "ча", "жа"}, Mods: []string{"-и", "-е", "-у", "-ей", "-е"}},
		{Gender: Androgynous, Test: []string{"а"}, Mods: []string{"-ы", "-е", "-у", "-ой", "-е"}},
		{Gender: Female, Test: []string{"ия"}, Mods: []string{"-и", "-и", "-ю", "-ей", "-и"}},
		{Gender: Female, Test: []string{"ка", "га", "ха"}, Mods: []string{"-и", "-е", "-у", "-ой", "-е"}},
		{Gender: Female, Test: []string{"ца"}, Mods: []string{"-ы", "-е", "-у", "-ей", "-е"}},
		{Gender: Female, Test: []string{"а"}, Mods: []string{"-ы", "-е", "-у", "-ой", "-е"}},
		{Gender: Female, Test: []string{"я"}, Mods: []string{"-и", "-е", "-ю", "-ей", "-е"}},
		{Gender: Male, Test: []string{"ия"}, Mods: []string{"-и", "-и", "-ю", "-ей", "-и"}},
		{Gender: Male, Test: []string{"я"}, Mods: []string{"-и", "-е", "-ю", "-ей", "-е"}},
		{Gender: Male, Test: []string{"ий"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-и"}},
		{Gender: Male, Test: []string{"ый", "кий", "хий"}, Mods: []string{"--ого", "--ому", "--ого", "-м", "--ом"}},
		{Gender: Male, Test: []string{"ей", "й"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Male, Test: []string{"ш", "ж"}, Mods: []string{"а", "у", "а", "ем", "е"}},
		{Gender: Male, Test: []string{"ёл"}, Mods: []string{"--ла", "--лу", "--ла", "--лом", "--ле"}},
		{Gender: Male, Test: []string{"ёк"}, Mods: []string{"--ька", "--ьку", "--ька", "--ьком", "--ьке"}},
		{Gender: Male, Test: []string{
			"б", "в", "г", "д", "ж", "з", "к", "л", "м", "н",
			"п", "р", "с", "т", "ф", "х", "ц", "ч",
		}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Androgynous, Test: []string{"ния", "рия", "вия"}, Mods: []string{"-и", "-и", "-ю", "-ей", "-и"}},
	},
}