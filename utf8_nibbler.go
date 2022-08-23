package nibblers

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// UTF8Nibbler is any nibbler that operates on UTF8 character encodings.
type UTF8Nibbler interface {
	// ReadCharacter reads and returns the next character from the nibbler stream. If the cursor is past the end of the
	// stream, returns io.EOF. If any other error occurs, that error is returned.
	ReadCharacter() (rune, error)

	// UnreadCharacter "returns" the last character to the nibbler stream. Abstractly, it moves the cursor one position
	// to the left (toward the start of the stream). Returns an error if rewind parsing fails or the cursor is already
	// at the first character. Any implementor for this interface does not need to support Unreading all the way back
	// to the start of the stream, so in general use, UnreadCharacter should not expect to be able to rewind mroe than
	// a small number of characters.
	UnreadCharacter() error

	// PeekAtNextCharacter returns the next character in the stream, but does not advance the cursor. It may return
	// io.EOF or an error in the same way and for the same reason that ReadCharacter() does.
	PeekAtNextCharacter() (rune, error)

	// Bookends instruct the Nibbler to preserve characters that are read in the backing store.  This starts a bookend
	// at the next unread character (though the character may have been peeked).
	StartBookending() error

	// This returns whatever is between the bookend start and the last character read (but not peeked).
	CharactersSinceStartOfBookend() ([]rune, error)

	// This stops the bookend at the last read character (but not peeked character) and returns a slice
	// containing the contents between the bookends.  The nibbler is permitted to discard the bookended characters from its
	// backing buffer.
	StopBookending() ([]rune, error)
}

// UTF8StringNibbler is a UTF8Nibbler that operates on golang strings, treating them as UTF8 byte streams.
type UTF8StringNibbler struct {
	backingString                     string
	indexInStringOfNextReadByte       int
	bookendStartOffsetInBackingString int // negative if no bookend start is active
}

// NewUTF8StringNibbler creates a new UTF8StringNibbler that will operate on the provided string.
func NewUTF8StringNibbler(nibbleString string) *UTF8StringNibbler {
	return &UTF8StringNibbler{
		backingString:                     nibbleString,
		indexInStringOfNextReadByte:       0,
		bookendStartOffsetInBackingString: -1,
	}
}

// ReadCharacter reads the next rune from the source string, returning it, returning io.EOF
// if the read cursor is now beyond the end of the string, or an error.  If an error is returned
// the value of the rune is undefined and the position of the pointer is undefined.
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

// UnreadCharacter moves the cursor back one UTF8 character in the string.  It returns an error
// if the cursor is already at the start of the string (i.e., pointing at the first UTF8 character)
// or if a decoding error occurs.
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

// PeekAtNextCharacter reads the next character in the source string and returns it, but
// does not advance the pointer.  If the pointer is already past the end of the string,
// io.EOF is returned.  If an error occurs, it is returned and both the value of the returned
// rune is undefined and the position of the pointer are undefined.
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

// StartBookending starts a bookend at the the next unread character.
func (nibbler *UTF8StringNibbler) StartBookending() error {
	if nibbler.indexInStringOfNextReadByte >= len(nibbler.backingString) {
		return io.EOF
	}

	if nibbler.bookendStartOffsetInBackingString >= 0 {
		return fmt.Errorf("a bookend is already active")
	}

	nibbler.bookendStartOffsetInBackingString = nibbler.indexInStringOfNextReadByte

	return nil
}

// CharactersSinceStartOfBookend returns all read characters since the start of the current bookend.
func (nibbler *UTF8StringNibbler) CharactersSinceStartOfBookend() ([]rune, error) {
	if nibbler.bookendStartOffsetInBackingString < 0 {
		return nil, fmt.Errorf("no bookend was started")
	}

	return []rune(nibbler.backingString[nibbler.bookendStartOffsetInBackingString:nibbler.indexInStringOfNextReadByte]), nil
}

// StopBookending returns a rune slice from the underlying string from the character
// after the start of bookending to the last character read (but not peeked).
func (nibbler *UTF8StringNibbler) StopBookending() ([]rune, error) {
	if nibbler.bookendStartOffsetInBackingString < 0 {
		return nil, fmt.Errorf("no bookend was started")
	}

	s := nibbler.bookendStartOffsetInBackingString
	nibbler.bookendStartOffsetInBackingString = -1

	return []rune(nibbler.backingString[s:nibbler.indexInStringOfNextReadByte]), nil
}

// UTF8RuneSliceNibbler is a concrete implementation of UTF8Nibbler. It operates on a fixed rune slice.
type UTF8RuneSliceNibbler struct {
	backingSlice                     []rune
	indexOfLastReadRune              int
	bookendStartOffsetInBackingSlice int // negative if no bookend start is active
}

// NewUTF8RuneSliceNibbler returns a nibbler for the provided rune slice.
func NewUTF8RuneSliceNibbler(runeSlice []rune) *UTF8RuneSliceNibbler {
	return &UTF8RuneSliceNibbler{
		backingSlice:                     runeSlice,
		indexOfLastReadRune:              -1,
		bookendStartOffsetInBackingSlice: -1,
	}
}

// ReadCharacter returns the next rune from the slice. It return io.EOF if the nibbler cursor
// is past the end of the slice.
func (nibbler *UTF8RuneSliceNibbler) ReadCharacter() (rune, error) {
	if nibbler.indexOfLastReadRune == len(nibbler.backingSlice)-1 {
		return utf8.RuneError, io.EOF
	}

	nibbler.indexOfLastReadRune++
	return nibbler.backingSlice[nibbler.indexOfLastReadRune], nil
}

// UnreadCharacter returns the next rune in the underlying slice or io.EOF if the
// cursor is past the end of the slice. It returns an error if the cursor is already
// at the start of the slice.
func (nibbler *UTF8RuneSliceNibbler) UnreadCharacter() error {
	if nibbler.indexOfLastReadRune < 0 {
		return fmt.Errorf("already at start of rune stream")
	}

	nibbler.indexOfLastReadRune--

	return nil
}

// PeekAtNextCharacter returns the next character in the slice without advancing the
// cursor. It returns io.EOF if the cursor is past the end of the slice.
func (nibbler *UTF8RuneSliceNibbler) PeekAtNextCharacter() (rune, error) {
	if nibbler.indexOfLastReadRune == len(nibbler.backingSlice)-1 {
		return utf8.RuneError, io.EOF
	}

	return nibbler.backingSlice[nibbler.indexOfLastReadRune+1], nil
}

// StartBookending instruct the Nibbler to preserve characters that are read in the backing store
func (nibbler *UTF8RuneSliceNibbler) StartBookending() error {
	if nibbler.indexOfLastReadRune >= len(nibbler.backingSlice) {
		return io.EOF
	}

	if nibbler.bookendStartOffsetInBackingSlice >= 0 {
		return fmt.Errorf("a bookend is already active")
	}

	nibbler.bookendStartOffsetInBackingSlice = nibbler.indexOfLastReadRune

	return nil
}

// CharactersSinceStartOfBookend returns all read characters since the start of the current bookend.
func (nibbler *UTF8RuneSliceNibbler) CharactersSinceStartOfBookend() ([]rune, error) {
	if nibbler.bookendStartOffsetInBackingSlice < 0 {
		return nil, fmt.Errorf("no active bookend")
	}

	return nibbler.backingSlice[nibbler.bookendStartOffsetInBackingSlice : nibbler.indexOfLastReadRune+1], nil
}

// StopBookending stops the bookend at the last read character and returns a slice containing the contents of the bookend.
func (nibbler *UTF8RuneSliceNibbler) StopBookending() ([]rune, error) {
	if nibbler.bookendStartOffsetInBackingSlice < 0 {
		return nil, fmt.Errorf("no active bookend")
	}

	s := nibbler.bookendStartOffsetInBackingSlice
	nibbler.bookendStartOffsetInBackingSlice = -1

	return nibbler.backingSlice[s : nibbler.indexOfLastReadRune+1], nil
}

// UTF8ByteSliceNibbler is a concrete implementation of UTF8Nibbler, operating on a
// byte slice, which must contain only valid UTF8 sequences.
type UTF8ByteSliceNibbler struct {
	underlyingStringNibbler *UTF8StringNibbler
}

// NewUTF8ByteSliceNibbler returns a new UTF8ByteSliceNibbler operating on the
// provided byte slice.
func NewUTF8ByteSliceNibbler(byteSlice []byte) *UTF8ByteSliceNibbler {
	return &UTF8ByteSliceNibbler{
		underlyingStringNibbler: NewUTF8StringNibbler(string(byteSlice)),
	}
}

// ReadCharacter attempts to read the next UTF8 encoded character from the slice. If
// successful, returns the next rune. If the cursor is passed the end of the slice, returns
// io.EOF. If the bytes starting at the cursor are not valid UTF8, return an error.
func (nibbler *UTF8ByteSliceNibbler) ReadCharacter() (rune, error) {
	return nibbler.underlyingStringNibbler.ReadCharacter()
}

// UnreadCharacter attempts to "return" the last read character to the slice. Abstractly,
// it moves the cursor one UTF8 encoded character closer to the start of the underlying slice.
// Return an error if the byte sequence before the cursor do not form a valid UTF8 encoding or
// if the cursor is already at the start of the slice.
func (nibbler *UTF8ByteSliceNibbler) UnreadCharacter() error {
	return nibbler.underlyingStringNibbler.UnreadCharacter()
}

// PeekAtNextCharacter is logically the same as ReadCharacter() followed by UnreadCharacter().
func (nibbler *UTF8ByteSliceNibbler) PeekAtNextCharacter() (rune, error) {
	return nibbler.underlyingStringNibbler.PeekAtNextCharacter()
}

// StartBookending instruct the Nibbler to preserve characters that are read in the backing store.
func (nibbler *UTF8ByteSliceNibbler) StartBookending() error {
	return nibbler.underlyingStringNibbler.StartBookending()
}

// CharactersSinceStartOfBookend returns all read characters since the start of the current bookend.
func (nibbler *UTF8ByteSliceNibbler) CharactersSinceStartOfBookend() ([]rune, error) {
	returnedSubString, err := nibbler.underlyingStringNibbler.CharactersSinceStartOfBookend()
	if err != nil {
		return nil, err
	}

	return []rune(returnedSubString), nil
}

// StopBookending stops the bookend at the last read character and returns a slice containing the contents of the bookend.
func (nibbler *UTF8ByteSliceNibbler) StopBookending() ([]rune, error) {
	returnedSubString, err := nibbler.underlyingStringNibbler.StopBookending()
	if err != nil {
		return nil, err
	}

	return []rune(returnedSubString), nil
}

// UTF8ReaderNibbler is a concrete implementation of UTF8Nibbler, operating on an io.Reader().
// It will trigger Read() when necessary to read more characters from the stream, until it reaches
// io.EOF or an error on Read().
type UTF8ReaderNibbler struct {
	sourceReader                     io.Reader
	readBuffer                       []byte
	bufferOfReadBytes                []byte
	indexInReadBytesBufferOfNextRune int
	indexInBufferOfBookendStart      int
}

// NewUTF8ReaderNibbler returns a new UTF8ReaderNibbler using the provided reader as the source. The
// io.Reader must only returns validly encoded UTF8 encoded bytes.
func NewUTF8ReaderNibbler(sourceReader io.Reader) *UTF8ReaderNibbler {
	return &UTF8ReaderNibbler{
		sourceReader:                     sourceReader,
		readBuffer:                       make([]byte, 9000),
		bufferOfReadBytes:                make([]byte, 0, 9000),
		indexInReadBytesBufferOfNextRune: 0,
		indexInBufferOfBookendStart:      -1,
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
	}

	return nil
}

// ReadCharacter attempts to read the next UTF8 encoded character from the underlying reader. If it
// succeeds the corresponding rune is returned.  If the reader returns io.EOF, return that. If
// the next set of bytes read are not a valid UTF8 encoding, return an error.
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

// UnreadCharacter attempts to "return" the last read UTF8 sequence to the stream. The intermediate stored
// buffer is constrained so UnreadCharacter() may not be able to reach the start of the stream. If the cursor
// is at the start of the intermediate buffer, return an error.
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

// PeekAtNextCharacter is logically the same as ReadCharacter() followed by UnreadCharacter().
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

// StartBookending instruct the Nibbler to preserve characters that are read in the backing store.
func (nibbler *UTF8ReaderNibbler) StartBookending() error {
	if nibbler.indexInBufferOfBookendStart >= 0 {
		return fmt.Errorf("a bookend is already active")
	}

	nibbler.indexInBufferOfBookendStart = nibbler.indexInReadBytesBufferOfNextRune

	return nil
}

// CharactersSinceStartOfBookend returns all read characters since the start of the current bookend.
func (nibbler *UTF8ReaderNibbler) CharactersSinceStartOfBookend() ([]rune, error) {
	if nibbler.indexInBufferOfBookendStart < 0 {
		return nil, fmt.Errorf("no bookmark is active")
	}

	return []rune(string(nibbler.bufferOfReadBytes[nibbler.indexInBufferOfBookendStart:nibbler.indexInReadBytesBufferOfNextRune])), nil
}

// StopBookending stops the bookend at the last read character and returns a slice containing the contents of the bookend.
func (nibbler *UTF8ReaderNibbler) StopBookending() ([]rune, error) {
	if nibbler.indexInBufferOfBookendStart < 0 {
		return nil, fmt.Errorf("no bookmark is active")
	}

	s := nibbler.indexInBufferOfBookendStart
	nibbler.indexInBufferOfBookendStart = -1

	return []rune(string(nibbler.bufferOfReadBytes[s:nibbler.indexInReadBytesBufferOfNextRune])), nil
}
