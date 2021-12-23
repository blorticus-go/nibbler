package nibbler

import (
	"unicode/utf8"
)

type NamedUnicodeCharacterSetsMap struct {
	mapOfSetsByName map[string]map[rune]bool
}

func NewUnicodeCharacterSetsMap() *NamedUnicodeCharacterSetsMap {
	return &NamedUnicodeCharacterSetsMap{
		mapOfSetsByName: make(map[string]map[rune]bool),
	}
}

func (setsMap *NamedUnicodeCharacterSetsMap) AddNamedUnicodeCharactersSetFromString(nameOfSet string, utf8CharacterSet string) *NamedUnicodeCharacterSetsMap {
	mapOfCharacters := make(map[rune]bool)

	characterSetAsByteArray := []byte(utf8CharacterSet)
	var nextRune rune

	for i, byteWidthOfRune := 0, 0; i < len(utf8CharacterSet); i += byteWidthOfRune {
		if nextRune, byteWidthOfRune = utf8.DecodeRune(characterSetAsByteArray[i:]); nextRune == utf8.RuneError {
			break
		}
		mapOfCharacters[nextRune] = true
	}

	setsMap.mapOfSetsByName[nameOfSet] = mapOfCharacters

	return setsMap
}

func (setsMap *NamedUnicodeCharacterSetsMap) AddNamedUnicodeCharacterSetFromRuneArray(nameOfSet string, runearray []rune) *NamedUnicodeCharacterSetsMap {
	mapOfCharacters := make(map[rune]bool)
	for _, r := range runearray {
		mapOfCharacters[r] = true
	}

	setsMap.mapOfSetsByName[nameOfSet] = mapOfCharacters

	return setsMap
}

type UTF8Nibbler interface {
	ReadCaracter() (rune, error)
	UnreadCharacter() error
	PeekAtNextCharacter() (rune, error)
	AddNamedCharacterSetsMap(*NamedUnicodeCharacterSetsMap)
	ReadNextCharactersMatchingSet(setName string) ([]rune, error)
	ReadNextCharactersNotMatchingSet(setName string) ([]rune, error)
}

type UTF8StringNibbler struct{}

type UTF8RuneSliceNibbler struct{}

type UTF8ByteSliceibbler struct{}

type UTF8ReaderNibbler struct{}
