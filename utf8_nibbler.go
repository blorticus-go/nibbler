package nibblers

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type simpleSizeConstrainedStackOfUints struct {
	underlyingSlice   []uint
	indexOfHead       int
	currentStackDepth int
	maxStackDepth     int
}

func newSimpleSizeConstrainedStackOfUints(maximuimStackLength uint) *simpleSizeConstrainedStackOfUints {
	if maximuimStackLength < 2 {
		panic("simpleSizeConstrainedStackOfUints length must be at least 2")
	}

	return &simpleSizeConstrainedStackOfUints{
		underlyingSlice:   make([]uint, maximuimStackLength),
		indexOfHead:       -1,
		maxStackDepth:     int(maximuimStackLength),
		currentStackDepth: 0,
	}
}

func (stack *simpleSizeConstrainedStackOfUints) Push(value uint) {
	if stack.indexOfHead == stack.maxStackDepth-1 {
		stack.indexOfHead = 0
	} else {
		stack.indexOfHead++
	}

	stack.underlyingSlice[stack.indexOfHead] = value

	if stack.currentStackDepth < stack.maxStackDepth {
		stack.currentStackDepth++
	}
}

func (stack *simpleSizeConstrainedStackOfUints) Pop() (value uint, stackIsAlreadyEmpty bool) {
	if stack.currentStackDepth == 0 {
		return 0, true
	}

	value = stack.underlyingSlice[stack.indexOfHead]
	if stack.indexOfHead == 0 {
		stack.indexOfHead = len(stack.underlyingSlice) - 1
	} else {
		stack.indexOfHead--
	}

	stack.currentStackDepth--

	return value, false
}

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

type UTF8StringNibbler struct {
	backingString               string
	indexInStringOfNextReadByte int
}

func NewUTF8StringNibbler(nibbleString string) *UTF8StringNibbler {
	return &UTF8StringNibbler{
		backingString:               nibbleString,
		indexInStringOfNextReadByte: 0,
	}
}

func (nibbler *UTF8StringNibbler) ReadCaracter() (rune, error) {
	if nibbler.indexInStringOfNextReadByte >= len(nibbler.backingString) {
		return 0, io.EOF
	}

	nextCharacter, sizeOfCharacterInBytes := utf8.DecodeRuneInString(nibbler.backingString[nibbler.indexInStringOfNextReadByte:])
	if nextCharacter == utf8.RuneError {
		return 0, fmt.Errorf("Invalid UTF-8 string element")
	}

	nibbler.indexInStringOfNextReadByte += sizeOfCharacterInBytes
	return nextCharacter, nil
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
