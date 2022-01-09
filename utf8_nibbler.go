package nibblers

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

type CharacterMatchingFunction func(r rune) (runeMatches bool)

type UTF8Nibbler interface {
	ReadCaracter() (rune, error)
	UnreadCharacter() error
	PeekAtNextCharacter() (rune, error)
	ReadCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error)
	ReadCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error)
	ReadConsecutiveWhitespace() ([]rune, error)
	ReadConsecutiveWhitespaceInto(receiver []rune) (int, error)
	ReadConsecutiveWordCharacters() ([]rune, error)
	ReadConsecutiveWordCharactersInto([]rune) (int, error)
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
		return utf8.RuneError, io.EOF
	}

	nextCharacter, sizeOfCharacterInBytes := utf8.DecodeRuneInString(nibbler.backingString[nibbler.indexInStringOfNextReadByte:])
	if nextCharacter == utf8.RuneError {
		return utf8.RuneError, fmt.Errorf("invalid UTF-8 string element")
	}

	nibbler.indexInStringOfNextReadByte += sizeOfCharacterInBytes
	return nextCharacter, nil
}

func (nibbler *UTF8StringNibbler) UnreadCharacter() error {
	s := nibbler.backingString[:nibbler.indexInStringOfNextReadByte]
	previousRune, sizeOfPreviousRune := utf8.DecodeLastRuneInString(s)

	if previousRune == utf8.RuneError {
		if sizeOfPreviousRune == 0 {
			return fmt.Errorf("already at start of string")
		}

		return fmt.Errorf("invalid UTF-8 string element")
	}

	nibbler.indexInStringOfNextReadByte -= sizeOfPreviousRune

	return nil
}

func (nibbler *UTF8StringNibbler) PeekAtNextCharacter() (rune, error) {
	if nibbler.indexInStringOfNextReadByte >= len(nibbler.backingString) {
		return utf8.RuneError, io.EOF
	}

	nextCharacter, _ := utf8.DecodeRuneInString(nibbler.backingString[nibbler.indexInStringOfNextReadByte:])
	if nextCharacter == utf8.RuneError {
		return 0, fmt.Errorf("invalid UTF-8 string element")
	}

	return nextCharacter, nil
}

func runeIsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func (nibbler *UTF8StringNibbler) ReadCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	matchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			if err == io.EOF {
				if len(matchingRunes) == 0 {
					return nil, io.EOF
				}

				return matchingRunes, nil
			}

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
			if err == io.EOF {
				if len(nonMatchingRunes) == 0 {
					return nil, io.EOF
				}
			}
			return nonMatchingRunes, nil
		}

		if !matcher(nextRune) {
			nonMatchingRunes = append(nonMatchingRunes, nextRune)
		} else {
			nibbler.UnreadCharacter()
			return nonMatchingRunes, nil
		}
	}
}

func (nibbler *UTF8StringNibbler) ReadCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			if err == io.EOF {
				if i == 0 {
					return 0, io.EOF
				}

				return i, nil
			}

			return -1, err
		}

		if matcher(nextRune) {
			receiver[i] = nextRune
		} else {
			nibbler.UnreadCharacter()
			return i, nil
		}
	}

	return len(receiver), nil
}

func (nibbler *UTF8StringNibbler) ReadCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCaracter()
		if err != nil {
			if err == io.EOF {
				if i == 0 {
					return 0, io.EOF
				}

				return i, nil
			}

			return -1, err
		}

		if !matcher(nextRune) {
			receiver[i] = nextRune
		} else {
			nibbler.UnreadCharacter()
			return i, nil
		}
	}

	return len(receiver), nil
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.ReadCharactersMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return nibbler.ReadCharactersMatchingInto(runeIsWhitespace, receiver)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.ReadCharactersNotMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
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

func (nibbler *UTF8ReaderNibbler) ReadNextCharactersMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}

func (nibbler *UTF8ReaderNibbler) ReadNextCharactersNotMatchingSet(setName string) ([]rune, error) {
	return nil, nil
}
