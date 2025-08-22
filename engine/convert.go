package engine

func ConvertLangKeyMap(langKeyMap LangKeyMap) KeyLangMap {
	var keyLangMap KeyLangMap

	keyLangMap = make(KeyLangMap)

	for lang, keyMap := range langKeyMap {
		for key, translation := range keyMap {
			_, ok := keyLangMap[key]

			if !ok {
				keyLangMap[key] = make(map[Lang]Translation)
			}

			keyLangMap[key][lang] = translation
		}
	}

	return keyLangMap
}

func ConvertKeyLangMap(keyLangMap KeyLangMap) LangKeyMap {
	var langKeyMap LangKeyMap

	langKeyMap = make(LangKeyMap)

	for key, langMap := range keyLangMap {
		for lang, translation := range langMap {
			_, ok := langKeyMap[lang]

			if !ok {
				langKeyMap[lang] = make(map[Key]Translation)
			}

			langKeyMap[lang][key] = translation
		}
	}

	return langKeyMap
}
