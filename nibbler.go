package nibbler

import (
	"bufio"
	"fmt"
	"io"
)

// NamedCharacterSetsMap stores sets of ASCII characters, associated with a name.  These can
// be provided to ByteNibblers when reading a string of characters from the input stream to
// determine which characters are allowed as part of the read.
type NamedCharacterSetsMap struct {
	mapOfSetsByName map[string]map[byte]bool
}

// NewNamedCharacterSetsMap creates a new empty map.
func NewNamedCharacterSetsMap() *NamedCharacterSetsMap {
	return &NamedCharacterSetsMap{
		mapOfSetsByName: make(map[string]map[byte]bool),
	}
}

// AddNamedCharacterSetFromString treats stringOfAsciiCharacters as a series of ASCII characters.
// Any rune with a value greater than 255 is ignored.  The set is added to the SetsMap with the
// provided name.
func (setsMap *NamedCharacterSetsMap) AddNamedCharacterSetFromString(nameOfSet string, stringOfASCIICharacters string) *NamedCharacterSetsMap {
	mapOfCharacters := make(map[byte]bool)
	for _, char := range stringOfASCIICharacters {
		if char > 0 && char < 255 {
			mapOfCharacters[byte(char)] = true
		}
	}

	setsMap.mapOfSetsByName[nameOfSet] = mapOfCharacters

	return setsMap
}

// AddNamedCharacterSetFromByteArray adds the bytes in byteArray as a character set with the provided name.
func (setsMap *NamedCharacterSetsMap) AddNamedCharacterSetFromByteArray(nameOfSet string, byteArray []byte) *NamedCharacterSetsMap {
	mapOfCharacters := make(map[byte]bool)
	for _, char := range byteArray {
		mapOfCharacters[byte(char)] = true
	}

	setsMap.mapOfSetsByName[nameOfSet] = mapOfCharacters

	return setsMap
}

func (setsMap *NamedCharacterSetsMap) retrieveNamedCharacterSet(nameOfSet string) map[byte]bool {
	return setsMap.mapOfSetsByName[nameOfSet]
}

// ByteNibbler is an interface for dealing with a byte buffer or byte stream one byte-at-a-time
// or in chunks based on character sets.  One can read a byte from the stream, return a read byte to the stream,
// look at the next byte from the stream without removing it, or extract bytes in a set.
// ReadByte and PeekAtNextByte should return io.EOF when the end of the stream has been reached.
type ByteNibbler interface {
	ReadByte() (byte, error)
	UnreadByte() error
	PeekAtNextByte() (byte, error)
	UseNamedCharacterSetsMap(*NamedCharacterSetsMap)
	ReadNextBytesFromSet(setName string) ([]byte, error)
}

// ByteSliceNibbler is a ByteNibbler using a static byte buffer.  A ReadByte or PeekAtNextbyte at
// the end of the slice will return io.EOF.
type ByteSliceNibbler struct {
	backingBuffer               []byte
	indexInBufferOfNextReadByte int
	namedCharacterSets          *NamedCharacterSetsMap
}

// NewByteSliceNibbler returns a new ByteSliceNibbler using the backing buffer.  Elements of the buffer
// backing array are not changed by any operation of the ByteSliceNibbler.
func NewByteSliceNibbler(buffer []byte) *ByteSliceNibbler {
	return &ByteSliceNibbler{
		backingBuffer:               buffer,
		indexInBufferOfNextReadByte: 0,
		namedCharacterSets:          nil,
	}
}

// UseNamedCharacterSetsMap receives a NamedCharacterSetsMap, to be used by ReadBytesFromSet().
func (nibbler *ByteSliceNibbler) UseNamedCharacterSetsMap(setsMap *NamedCharacterSetsMap) {
	nibbler.namedCharacterSets = setsMap
}

// ReadByte attempts to read the next byte in the backing array.  If all bytes have been read, then io.EOF
// is returned.
func (nibbler *ByteSliceNibbler) ReadByte() (byte, error) {
	if len(nibbler.backingBuffer) <= nibbler.indexInBufferOfNextReadByte {
		return 0, io.EOF
	}

	b := nibbler.backingBuffer[nibbler.indexInBufferOfNextReadByte]
	nibbler.indexInBufferOfNextReadByte++

	return b, nil
}

// UnreadByte puts the last read byte back on the nibbler pseudo-stack.  An error is returned if the
// the backing slice is empty or bytes have been unshifted back to the start of the slice.
func (nibbler *ByteSliceNibbler) UnreadByte() error {
	if nibbler.indexInBufferOfNextReadByte == 0 {
		return fmt.Errorf("already at the start of the backing buffer")
	}

	nibbler.indexInBufferOfNextReadByte--
	return nil
}

// PeekAtNextByte returns the next unread byte in the slice or io.EOF if the pseudo-stack is empty.
func (nibbler *ByteSliceNibbler) PeekAtNextByte() (byte, error) {
	if len(nibbler.backingBuffer) <= nibbler.indexInBufferOfNextReadByte {
		return 0, io.EOF
	}

	return nibbler.backingBuffer[nibbler.indexInBufferOfNextReadByte], nil
}

// ReadNextBytesFromSet reads bytes in the stream as long as they match the characters in the
// setName (which, in turn, must be supplied to the NamedCharacterSetsMap provided in UseNamedCharacterSetsMap).
// Return an error if no named character sets map has been provided, if the setName provided is
// not in that map, or if the stream read produces an error.  Note that this error may be io.EOF.
// Whether or not an error is returned, the assembled slice of bytes read from the stream is also returned.
// After this method returns, the nibbler's next byte is the one after the last character in the returned set.
func (nibbler *ByteSliceNibbler) ReadNextBytesFromSet(setName string) ([]byte, error) {
	if nibbler.namedCharacterSets == nil {
		return nil, fmt.Errorf("no character set with that name is defined")
	}

	setMap := nibbler.namedCharacterSets.retrieveNamedCharacterSet(setName)
	if setMap == nil {
		return nil, fmt.Errorf("no character set with that name is defined")
	}

	bytesInSet := make([]byte, 0, 20)

	for {
		character, err := nibbler.ReadByte()
		if err != nil {
			return bytesInSet, err
		}

		if _, characterIsInMap := setMap[character]; characterIsInMap {
			bytesInSet = append(bytesInSet, character)
		} else {
			_ = nibbler.UnreadByte()
			return bytesInSet, nil
		}
	}
}

// ByteReaderNibbler is a ByteNibbler that uses an io.Reader as its dynamic backing stream.
// There is no guarantee that the internal buffer representing the pseudo queue grows to
// the size of all bytes read, so if UnreadByte() is called repeatedly in succession, it may
// eventually return an error and may not allow a return of every byte previously read.
type ByteReaderNibbler struct {
	backingReader               io.Reader
	internalBuffer              []byte
	readBuffer                  []byte
	indexOfNextReadByteInBuffer int
	namedCharacterSets          *NamedCharacterSetsMap
}

// NewByteReaderNibbler returns a ByteReaderNibbler.
func NewByteReaderNibbler(streamReader io.Reader) *ByteReaderNibbler {
	return &ByteReaderNibbler{
		backingReader:               bufio.NewReader(streamReader),
		readBuffer:                  make([]byte, 9000),
		internalBuffer:              make([]byte, 0, 18000),
		indexOfNextReadByteInBuffer: 0,
		namedCharacterSets:          nil,
	}
}

// UseNamedCharacterSetsMap receives a NamedCharacterSetsMap, to be used by ReadBytesFromSet().
func (nibbler *ByteReaderNibbler) UseNamedCharacterSetsMap(setsMap *NamedCharacterSetsMap) {
	nibbler.namedCharacterSets = setsMap
}

func (nibbler *ByteReaderNibbler) readFromStreamAndAppendToInternalBuffer() error {
	bytesReadFromStream, err := nibbler.backingReader.Read(nibbler.readBuffer)
	if err != nil {
		return err
	}

	if bytesReadFromStream > 0 {
		nibbler.internalBuffer = append(nibbler.internalBuffer, nibbler.readBuffer[:bytesReadFromStream]...)
	} else {
		return fmt.Errorf("read of stream returned no bytes, no eof, and no error")
	}

	return nil
}

// ReadByte reads the next byte from the stream.  Return io.EOF if the end of the stream
// has been reached.
func (nibbler *ByteReaderNibbler) ReadByte() (byte, error) {
	if nibbler.indexOfNextReadByteInBuffer >= len(nibbler.internalBuffer) {
		if err := nibbler.readFromStreamAndAppendToInternalBuffer(); err != nil {
			return 0, err
		}
	}

	b := nibbler.internalBuffer[nibbler.indexOfNextReadByteInBuffer]
	nibbler.indexOfNextReadByteInBuffer++

	return b, nil
}

// UnreadByte returns the last read byte back to the buffered stream.  A subsequent ReadByte()
// will return this same byte.  An error is returned if the stream is empty or if the last
// read byte is the first byte in the stream.
func (nibbler *ByteReaderNibbler) UnreadByte() error {
	if nibbler.indexOfNextReadByteInBuffer < 1 {
		return fmt.Errorf("already at the start of the stream buffer")
	}

	nibbler.indexOfNextReadByteInBuffer--
	return nil
}

// PeekAtNextByte looks at the next byte in the stream and returns it without advancing
// the byte return pointer.  Thus, a subsequent call to ReadNext() will return the same
// byte.  Return io.EOF if currently at the end of the stream.
func (nibbler *ByteReaderNibbler) PeekAtNextByte() (byte, error) {
	if nibbler.indexOfNextReadByteInBuffer >= len(nibbler.internalBuffer) {
		if err := nibbler.readFromStreamAndAppendToInternalBuffer(); err != nil {
			return 0, err
		}
	}

	return nibbler.internalBuffer[nibbler.indexOfNextReadByteInBuffer], nil
}

// ReadNextBytesFromSet reads bytes in the stream as long as they match the characters in the
// setName (which, in turn, must be supplied to the NamedCharacterSetsMap provided in UseNamedCharacterSetsMap).
// Return an error if no named character sets map has been provided, if the setName provided is
// not in that map, or if the stream read produces an error.  Note that this error may be io.EOF.
// Whether or not an error is returned, the assembled slice of bytes read from the stream is also returned.
// After this method returns, the nibbler's next byte is the one after the last character in the returned set.
func (nibbler *ByteReaderNibbler) ReadNextBytesFromSet(setName string) ([]byte, error) {
	if nibbler.namedCharacterSets == nil {
		return nil, fmt.Errorf("no character set with that name is defined")
	}

	setMap := nibbler.namedCharacterSets.retrieveNamedCharacterSet(setName)
	if setMap == nil {
		return nil, fmt.Errorf("no character set with that name is defined")
	}

	bytesInSet := make([]byte, 0, 20)

	for {
		character, err := nibbler.ReadByte()
		if err != nil {
			return bytesInSet, err
		}

		if _, characterIsInMap := setMap[character]; characterIsInMap {
			bytesInSet = append(bytesInSet, character)
		} else {
			_ = nibbler.UnreadByte()
			return bytesInSet, nil
		}
	}
}
