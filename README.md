# nibbler

Nibblers provide a means to treat a slice and a `Reader` in a uniform way when they are used as the source for a stream of data, where the stream is generally read without (much) rewinding.  Compare reading data from a byte `Reader` when compared to reading data that is in a slice.  When a `Reader` is used, a `Read()` may be needed periodically, until `EOF` is encountered.  It may not be desirable to simply read all bytes into a single backing slice, particularly if the `Reader` is attached to a source that will generate a very large amount of data.  If a caller mostly wishes to move forward in the stream (perhaps with occassional `peek` and `unread` operations where needed), a large and growing backing slice may just represent a waste of memory. 

## Overview

All Nibblers use an abstract pointer to data units.  A data unit may be, for example, a byte or a rune.  All Nibblers support the same set of basic operations: `Read`, `Unread` and `Peek`.  `Read` will read one data unit from the stream and advance the pointer by one unit; `Unread` will rewind the pointer in the data stream by one unit; and `Peek` will look at the next data unit without advancing the stream pointer.  When the end of underlying stream is reached, `Read` and `Peek` return `io.EOF`.  Attempting to `Unread` when the pointer is at the start of the stream generates an `error`.  A Nibbler may choose to arbitrary limit the number of units that can be `Unread`.  This frees Nibblers from having to back a `Reader` with a growing backing data store.

Currently, there are two Nibbler types: a `ByteNibbler` and a `UTF8Reader`.  For a `ByteNibbler`, the data unit is a `byte`.  For a `UTF8Reader`, the data unit is a `rune.

The `interface` shared by all `ByteNibbler`s is:

```golang
type ByteNibbler interface {
	ReadByte() (byte, error)
	UnreadByte() error
	PeekAtNextByte() (byte, error)
	AddNamedByteSetsMap(*NamedByteSetsMap)
	ReadNextBytesMatchingSet(setName string) ([]byte, error)
	ReadNextBytesNotMatchingSet(setName string) ([]byte, error)
	ReadFixedNumberOfBytes(countOfBytesToRead uint) ([]byte, error)
}
```

Concrete `ByteNibbler` types vary by supplied stream type.

Similarly, the `interface` shared by all `UTF8Nibbler`s is:

```golang
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
```
