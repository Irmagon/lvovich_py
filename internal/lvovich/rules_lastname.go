package lvovich

// Правила склонения фамилий — порт inclineRulesLastname.ts
// библиотеки nodkz/lvovich (https://github.com/nodkz/lvovich), MIT.
// Порядок правил важен: первое совпадение выигрывает.

var lastnameRules = &RuleSet{
	Exceptions: []DeclensionRule{
		{
			Gender: Androgynous,
			Test: []string{
				"бонч", "абдул", "белиц", "гасан", "дюссар", "дюмон", "книппер", "корвин", "ван", "шолом",
				"тер", "призван", "мелик", "вар", "фон",
			},
			Mods: []string{".", ".", ".", ".", "."},
			Tags: []string{"first_word"},
		},
		{
			Gender: Androgynous,
			Test:   []string{"дюма", "тома", "дега", "люка", "ферма", "гамарра", "петипа", "шандра", "скаля", "каруана"},
			Mods:   []string{".", ".", ".", ".", "."},
		},
		{
			Gender: Androgynous,
			Test:   []string{"гусь", "ремень", "камень", "онук", "богода", "нечипас", "долгопалец", "маненок", "рева", "кива"},
			Mods:   []string{".", ".", ".", ".", "."},
		},
		{Gender: Androgynous, Test: []string{"вий", "сой", "цой", "хой"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
	},
	Suffixes: []DeclensionRule{
		{Gender: Female, Test: []string{
			"б", "в", "г", "д", "ж", "з", "й", "к", "л", "м",
			"н", "п", "р", "с", "т", "ф", "х", "ц", "ч", "ш",
			"щ", "ъ", "ь",
		}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Androgynous, Test: []string{"орота"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Female, Test: []string{"ска", "цка"}, Mods: []string{"-ой", "-ой", "-ую", "-ой", "-ой"}},
		{Gender: Female, Test: []string{"цкая", "ская", "ная", "ая"}, Mods: []string{"--ой", "--ой", "--ую", "--ой", "--ой"}},
		{Gender: Female, Test: []string{"яя"}, Mods: []string{"--ей", "--ей", "--юю", "--ей", "--ей"}},
		{Gender: Male, Test: []string{"иной", "уй"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Androgynous, Test: []string{"ца"}, Mods: []string{"-ы", "-е", "-у", "-ей", "-е"}},
		{Gender: Male, Test: []string{"рих"}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Androgynous, Test: []string{"ия"}, Mods: []string{"-и", "-и", "-ю", "-ей", "-и"}},
		{Gender: Androgynous, Test: []string{"иа", "аа", "оа", "уа", "ыа", "еа", "юа", "эа"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Androgynous, Test: []string{"о", "е", "э", "и", "ы", "у", "ю"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Male, Test: []string{"их", "ых"}, Mods: []string{".", ".", ".", ".", "."}},
		{Gender: Female, Test: []string{"ова", "ева", "на", "ёва"}, Mods: []string{"-ой", "-ой", "-у", "-ой", "-ой"}},
		{Gender: Androgynous, Test: []string{"га", "ка", "ха", "ча", "ща", "жа", "ша"}, Mods: []string{"-и", "-е", "-у", "-ой", "-е"}},
		{Gender: Androgynous, Test: []string{"а"}, Mods: []string{"-ы", "-е", "-у", "-ой", "-е"}},
		{Gender: Male, Test: []string{"ь"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Androgynous, Test: []string{"я"}, Mods: []string{"-и", "-е", "-ю", "-ей", "-е"}},
		{Gender: Male, Test: []string{"обей"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Male, Test: []string{"ей"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Male, Test: []string{"ян", "ан", "йн"}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Male, Test: []string{"ынец", "овец"}, Mods: []string{"--ца", "--цу", "--ца", "--цом", "--це"}},
		{Gender: Male, Test: []string{"нец", "обец"}, Mods: []string{"--ца", "--цу", "--ца", "--цем", "--це"}},
		{Gender: Male, Test: []string{"ай"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-е"}},
		{Gender: Male, Test: []string{"гой", "кой"}, Mods: []string{"-го", "-му", "-го", "--ым", "-м"}},
		{Gender: Male, Test: []string{"ой"}, Mods: []string{"-го", "-му", "-го", "--ым", "-м"}},
		{Gender: Male, Test: []string{"ах", "ив", "шток"}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Male, Test: []string{"ший", "щий", "жий", "ний"}, Mods: []string{"--его", "--ему", "--его", "-м", "--ем"}},
		{Gender: Male, Test: []string{"ый", "кий", "хий"}, Mods: []string{"--ого", "--ому", "--ого", "-м", "--ом"}},
		{Gender: Male, Test: []string{"ий"}, Mods: []string{"-я", "-ю", "-я", "-ем", "-и"}},
		{Gender: Male, Test: []string{"ок"}, Mods: []string{"--ка", "--ку", "--ка", "--ком", "--ке"}},
		{Gender: Male, Test: []string{"иец", "еец"}, Mods: []string{"--йца", "--йцу", "--йца", "--йцом", "--йце"}},
		{Gender: Male, Test: []string{"ец"}, Mods: []string{"--ца", "--цу", "--ца", "--цом", "--це"}},
		{Gender: Male, Test: []string{"ц", "ч", "ш", "щ"}, Mods: []string{"а", "у", "а", "ем", "е"}},
		{Gender: Male, Test: []string{
			"ен", "нн", "он", "ун", "б", "г", "д", "ж", "з", "к",
			"л", "м", "п", "р", "с", "т", "ф", "х",
		}, Mods: []string{"а", "у", "а", "ом", "е"}},
		{Gender: Male, Test: []string{"в", "н"}, Mods: []string{"а", "у", "а", "ым", "е"}},
	},
}