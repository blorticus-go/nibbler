package nibblers

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

type CharacterMatchingFunction func(r rune) (runeMatches bool)

func runeIsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

type UTF8Nibbler interface {
	ReadCharacter() (rune, error)
	UnreadCharacter() error
	PeekAtNextCharacter() (rune, error)
	ReadConsecutiveCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadConsecutiveCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error)
	ReadConsecutiveCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error)
	ReadConsecutiveCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error)
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

func (nibbler *UTF8StringNibbler) ReadCharacter() (rune, error) {
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

func (nibbler *UTF8StringNibbler) ReadConsecutiveCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersMatching(nibbler, matcher)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersNotMatching(nibbler, matcher)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersNotMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersMatchingInto(runeIsWhitespace, receiver)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersNotMatching(runeIsWhitespace)
}

func (nibbler *UTF8StringNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersNotMatchingInto(runeIsWhitespace, receiver)
}

type UTF8RuneSliceNibbler struct {
	backingSlice        []rune
	indexOfLastReadRune int
}

func NewUTF8RuneSliceNibbler(runeSlice []rune) *UTF8RuneSliceNibbler {
	return &UTF8RuneSliceNibbler{
		backingSlice:        runeSlice,
		indexOfLastReadRune: -1,
	}
}

func (nibbler *UTF8RuneSliceNibbler) ReadCharacter() (rune, error) {
	if nibbler.indexOfLastReadRune == len(nibbler.backingSlice)-1 {
		return utf8.RuneError, io.EOF
	}

	nibbler.indexOfLastReadRune++
	return nibbler.backingSlice[nibbler.indexOfLastReadRune], nil
}

func (nibbler *UTF8RuneSliceNibbler) UnreadCharacter() error {
	if nibbler.indexOfLastReadRune < 0 {
		return fmt.Errorf("already at start of rune stream")
	}

	nibbler.indexOfLastReadRune--

	return nil
}

func (nibbler *UTF8RuneSliceNibbler) PeekAtNextCharacter() (rune, error) {
	if nibbler.indexOfLastReadRune == len(nibbler.backingSlice)-1 {
		return utf8.RuneError, io.EOF
	}

	return nibbler.backingSlice[nibbler.indexOfLastReadRune+1], nil
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersMatching(nibbler, matcher)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersNotMatching(nibbler, matcher)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersNotMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersMatching(runeIsWhitespace)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersMatchingInto(runeIsWhitespace, receiver)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersNotMatching(runeIsWhitespace)
}

func (nibbler *UTF8RuneSliceNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersNotMatchingInto(runeIsWhitespace, receiver)
}

type UTF8ByteSliceNibbler struct {
	underlyingStringNibbler *UTF8StringNibbler
}

func NewUTF8ByteSliceNibbler(byteSlice []byte) *UTF8ByteSliceNibbler {
	return &UTF8ByteSliceNibbler{
		underlyingStringNibbler: NewUTF8StringNibbler(string(byteSlice)),
	}
}

func (nibbler *UTF8ByteSliceNibbler) ReadCharacter() (rune, error) {
	return nibbler.underlyingStringNibbler.ReadCharacter()
}

func (nibbler *UTF8ByteSliceNibbler) UnreadCharacter() error {
	return nibbler.underlyingStringNibbler.UnreadCharacter()
}

func (nibbler *UTF8ByteSliceNibbler) PeekAtNextCharacter() (rune, error) {
	return nibbler.underlyingStringNibbler.PeekAtNextCharacter()
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveCharactersMatching(matcher)
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveCharactersNotMatching(matcher)
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveCharactersMatchingInto(matcher, receiver)
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveCharactersNotMatchingInto(matcher, receiver)
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveWhitespace()
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveWhitespaceInto(receiver)
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveWordCharacters()
}

func (nibbler *UTF8ByteSliceNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
	return nibbler.underlyingStringNibbler.ReadConsecutiveWordCharactersInto(receiver)
}

type UTF8ReaderNibbler struct {
	sourceReader                     io.Reader
	readBuffer                       []byte
	bufferOfReadBytes                []byte
	indexInReadBytesBufferOfNextRune int
}

func NewUTF8ReaderNibbler(sourceReader io.Reader) *UTF8ReaderNibbler {
	return &UTF8ReaderNibbler{
		sourceReader:                     sourceReader,
		readBuffer:                       make([]byte, 9000),
		bufferOfReadBytes:                make([]byte, 0, 9000),
		indexInReadBytesBufferOfNextRune: -1,
	}
}

func (nibbler *UTF8ReaderNibbler) readFromStreamIntoReadBuffer() (bytesRead int, err error) {
	countOfReadBytes, err := nibbler.sourceReader.Read(nibbler.readBuffer)
	if err != nil {
		return countOfReadBytes, err
	}

	if countOfReadBytes == 0 {
		return countOfReadBytes, fmt.Errorf("nothing returned from Read()")
	}

	nibbler.bufferOfReadBytes = append(nibbler.bufferOfReadBytes, nibbler.readBuffer[:countOfReadBytes]...)

	return countOfReadBytes, nil
}

func (nibbler *UTF8ReaderNibbler) triggerReadFromStreamIntoBufferIfNeeded() error {
	if nibbler.indexInReadBytesBufferOfNextRune >= len(nibbler.bufferOfReadBytes) {
		if _, err := nibbler.readFromStreamIntoReadBuffer(); err != nil {
			return err
		}
	} else if nibbler.indexInReadBytesBufferOfNextRune < 0 {
		if _, err := nibbler.readFromStreamIntoReadBuffer(); err != nil {
			return err
		}

		nibbler.indexInReadBytesBufferOfNextRune = 0
	}

	return nil
}

func (nibbler *UTF8ReaderNibbler) ReadCharacter() (rune, error) {
	if err := nibbler.triggerReadFromStreamIntoBufferIfNeeded(); err != nil {
		return utf8.RuneError, err
	}

	nextRuneInByteStream, numberOfBytesConsumedByRune := utf8.DecodeRune(nibbler.bufferOfReadBytes[nibbler.indexInReadBytesBufferOfNextRune:])
	if nextRuneInByteStream != utf8.RuneError {
		nibbler.indexInReadBytesBufferOfNextRune += numberOfBytesConsumedByRune
		return nextRuneInByteStream, nil
	}

	for bytesAddedToReadBuffer := 0; bytesAddedToReadBuffer <= 4; {
		countOfReadBytes, err := nibbler.readFromStreamIntoReadBuffer()
		if err != nil {
			return utf8.RuneError, err
		}

		nextRuneInByteStream, numberOfBytesConsumedByRune := utf8.DecodeRune(nibbler.bufferOfReadBytes[nibbler.indexInReadBytesBufferOfNextRune:])
		if nextRuneInByteStream != utf8.RuneError {
			nibbler.indexInReadBytesBufferOfNextRune += numberOfBytesConsumedByRune
			return nextRuneInByteStream, nil
		}

		bytesAddedToReadBuffer += countOfReadBytes
	}

	return utf8.RuneError, fmt.Errorf("invalid UTF-8 encoding in stream")
}

func (nibbler *UTF8ReaderNibbler) UnreadCharacter() error {
	if nibbler.indexInReadBytesBufferOfNextRune <= 0 {
		return fmt.Errorf("already at start of stream")
	}

	previousRuneInReadBuffer, bytesRequiredForPreviousRune := utf8.DecodeLastRune(nibbler.bufferOfReadBytes[:nibbler.indexInReadBytesBufferOfNextRune])
	if previousRuneInReadBuffer == utf8.RuneError || bytesRequiredForPreviousRune == 0 {
		return fmt.Errorf("UTF-8 decode failure")
	}

	nibbler.indexInReadBytesBufferOfNextRune -= bytesRequiredForPreviousRune

	return nil
}

func (nibbler *UTF8ReaderNibbler) PeekAtNextCharacter() (rune, error) {
	nextRune, err := nibbler.ReadCharacter()
	if err != nil {
		return utf8.RuneError, err
	}

	if err := nibbler.UnreadCharacter(); err != nil {
		return utf8.RuneError, err
	}

	return nextRune, nil
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveCharactersMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersMatching(nibbler, matcher)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveCharactersNotMatching(matcher CharacterMatchingFunction) ([]rune, error) {
	return readConsecutiveCharactersNotMatching(nibbler, matcher)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveCharactersMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveCharactersNotMatchingInto(matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	return readConsecutiveCharactersNotMatchingInto(nibbler, matcher, receiver)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveWhitespace() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersMatching(runeIsWhitespace)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersMatchingInto(runeIsWhitespace, receiver)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveWordCharacters() ([]rune, error) {
	return nibbler.ReadConsecutiveCharactersNotMatching(runeIsWhitespace)
}

func (nibbler *UTF8ReaderNibbler) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
	return nibbler.ReadConsecutiveCharactersNotMatchingInto(runeIsWhitespace, receiver)
}

func readConsecutiveCharactersMatching(nibbler UTF8Nibbler, matcherFunction CharacterMatchingFunction) ([]rune, error) {
	matchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if len(matchingRunes) == 0 {
					return nil, io.EOF
				}

				return matchingRunes, nil
			}

			return matchingRunes, err
		}

		if matcherFunction(nextRune) {
			matchingRunes = append(matchingRunes, nextRune)
		} else {
			nibbler.UnreadCharacter()
			return matchingRunes, nil
		}
	}
}

func readConsecutiveCharactersNotMatching(nibbler UTF8Nibbler, matcherFunction CharacterMatchingFunction) ([]rune, error) {
	nonMatchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if len(nonMatchingRunes) == 0 {
					return nil, io.EOF
				}
			}
			return nonMatchingRunes, nil
		}

		if !matcherFunction(nextRune) {
			nonMatchingRunes = append(nonMatchingRunes, nextRune)
		} else {
			nibbler.UnreadCharacter()
			return nonMatchingRunes, nil
		}
	}
}

func readConsecutiveCharactersMatchingInto(nibbler UTF8Nibbler, matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCharacter()
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

func readConsecutiveCharactersNotMatchingInto(nibbler UTF8Nibbler, matcher CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := nibbler.ReadCharacter()
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
