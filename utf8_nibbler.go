package nibblers

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

func (nibbler *UTF8StringNibbler) ReadCaracter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8StringNibbler) UnreadCharacter() error {
	return nil
}

func (nibbler *UTF8StringNibbler) PeekAtNextCharacter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8StringNibbler) AddNamedCharacterSetsMap(*NamedUnicodeCharacterSetsMap) {

}

func (nibbler *UTF8StringNibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

func (nibbler *UTF8StringNibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

type UTF8RuneSliceNibbler struct{}

func (nibbler *UTF8RuneSliceNibbler) ReadCaracter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8RuneSliceNibbler) UnreadCharacter() error {
	return nil
}

func (nibbler *UTF8RuneSliceNibbler) PeekAtNextCharacter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8RuneSliceNibbler) AddNamedCharacterSetsMap(*NamedUnicodeCharacterSetsMap) {

}

func (nibbler *UTF8RuneSliceNibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

func (nibbler *UTF8RuneSliceNibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

type UTF8ByteSliceibbler struct{}

func (nibbler *UTF8ByteSliceibbler) ReadCaracter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8ByteSliceibbler) UnreadCharacter() error {
	return nil
}

func (nibbler *UTF8ByteSliceibbler) PeekAtNextCharacter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8ByteSliceibbler) AddNamedCharacterSetsMap(*NamedUnicodeCharacterSetsMap) {

}

func (nibbler *UTF8ByteSliceibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

func (nibbler *UTF8ByteSliceibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

type UTF8ReaderNibbler struct{}

func (nibbler *UTF8ReaderNibbler) ReadCaracter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8ReaderNibbler) UnreadCharacter() error {
	return nil
}

func (nibbler *UTF8ReaderNibbler) PeekAtNextCharacter() (rune, error) {
	return 0, nil
}

func (nibbler *UTF8ReaderNibbler) AddNamedCharacterSetsMap(*NamedUnicodeCharacterSetsMap) {

}

func (nibbler *UTF8ReaderNibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

func (nibbler *UTF8ReaderNibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}
