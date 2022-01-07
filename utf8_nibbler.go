package nibblers

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/blorticus-go/stack"
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

type CharacterMatchingFunction func(r rune) (runeMatches bool)

type UTF8Nibbler interface {
	ReadCaracter() (rune, error)
	UnreadCharacter() error
	PeekAtNextCharacter() (rune, error)
	ReadCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) error
	ReadCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) error
	ReadConsecutiveWhitespace() ([]rune, error)
	ReadConsecutiveWhitespaceInto(receiver []rune) error
	ReadConsecutiveWordCharacters() ([]rune, error)
	ReadConsecutiveWordCharactersInto([]rune) error
}

type UTF8StringNibbler struct {
	backingString                      string
	indexInStringOfNextReadByte        int
	stackOfLastTenCharacterByteLengths *stack.Stack
	namedCharacterSetsMap              *NamedUnicodeCharacterSetsMap
}

func NewUTF8StringNibbler(nibbleString string) *UTF8StringNibbler {
	return &UTF8StringNibbler{
		backingString:                      nibbleString,
		indexInStringOfNextReadByte:        0,
		stackOfLastTenCharacterByteLengths: stack.NewBoundedDiscardingStack(10),
		namedCharacterSetsMap:              nil,
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
	nibbler.stackOfLastTenCharacterByteLengths.Push(sizeOfCharacterInBytes)
	return nextCharacter, nil
}

func (nibbler *UTF8StringNibbler) UnreadCharacter() error {
	sizeOfLastReadRune, stackWasEmptyBeforeRead := nibbler.stackOfLastTenCharacterByteLengths.PopInt()
	if stackWasEmptyBeforeRead {
		return fmt.Errorf("Already at the start of the stream")
	}

	nibbler.indexInStringOfNextReadByte -= sizeOfLastReadRune
	return nil
}

func (nibbler *UTF8StringNibbler) PeekAtNextCharacter() (rune, error) {
	if nibbler.indexInStringOfNextReadByte >= len(nibbler.backingString) {
		return 0, io.EOF
	}

	nextCharacter, _ := utf8.DecodeRuneInString(nibbler.backingString[nibbler.indexInStringOfNextReadByte:])
	if nextCharacter == utf8.RuneError {
		return 0, fmt.Errorf("Invalid UTF-8 string element")
	}

	return nextCharacter, nil
}

func (nibbler *UTF8StringNibbler) AddNamedCharacterSetsMap(setsMap *NamedUnicodeCharacterSetsMap) {
	nibbler.namedCharacterSetsMap = setsMap
}

func (nibbler *UTF8StringNibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	if nibbler.namedCharacterSetsMap == nil {
		return nil, fmt.Errorf("No NamedCharacterSetsMap defined")
	}

	set, setIsInMap := nibbler.namedCharacterSetsMap.mapOfSetsByName[setName]
	if !setIsInMap {
		return nil, fmt.Errorf("No set named (%s) in associated NamedCharacterSetsMap", setName)
	}

	matchingRunes := make([]rune, 0, 10)
	for {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			return matchingRunes, err
		}

		if _, runeIsInMap := set[nextRune]; !runeIsInMap {
			nibbler.UnreadCharacter()
			return matchingRunes, nil
		}

		matchingRunes = append(matchingRunes, nextRune)
	}
}

func (nibbler *UTF8StringNibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	if nibbler.namedCharacterSetsMap == nil {
		return nil, fmt.Errorf("No NamedCharacterSetsMap defined")
	}

	set, setIsInMap := nibbler.namedCharacterSetsMap.mapOfSetsByName[setName]
	if !setIsInMap {
		return nil, fmt.Errorf("No set named (%s) in associated NamedCharacterSetsMap", setName)
	}

	matchingRunes := make([]rune, 0, 10)
	for {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			return matchingRunes, err
		}

		if _, runeIsInMap := set[nextRune]; runeIsInMap {
			nibbler.UnreadCharacter()
			return matchingRunes, nil
		}

		matchingRunes = append(matchingRunes, nextRune)
	}
}

func runeIsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func (nibbler *UTF8StringNibbler) ReadCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	matchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			return matchingRunes, err
		}

		if matcher(nextRune) {
			matchingRunes = append(matchingRunes, nextRune)
		} else {
			nibbler.UnreadCharacter()
			return matchingRunes, nil
		}
	}
}

func (nibbler *UTF8StringNibbler) ReadCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	nonMatchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			return nonMatchingRunes, err
		}

		if !matcher(nextRune) {
			nonMatchingRunes = append(nonMatchingRunes, nextRune)
		} else {
			nibbler.UnreadCharacter()
			return nonMatchingRunes, nil
		}
	}
}

func (nibbler *UTF8StringNibbler) ReadCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) error {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			receiver = receiver[:i-1]
			return err
		}

		if matcher(nextRune) {
			receiver[i] = nextRune
		} else {
			nibbler.UnreadCharacter()
			receiver = receiver[:i-1]
			return nil
		}
	}

	return nil
}

func (nibbler *UTF8StringNibbler) ReadCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) error {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			receiver = receiver[:i-1]
			return err
		}

		if !matcher(nextRune) {
			receiver[i] = nextRune
		} else {
			nibbler.UnreadCharacter()
			receiver = receiver[:i-1]
			return nil
		}
	}

	return nil
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.ReadCharactersMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) error {
	return nibbler.ReadCharactersMatchingInto(runeIsWhitespace, receiver)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.ReadCharactersNotMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) error {
	return nibbler.ReadCharactersNotMatchingInto(runeIsWhitespace, receiver)
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
