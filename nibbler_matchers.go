package nibblers

import (
	"io"
	"unicode"
)

// UTF8NibblerMatcher is a wrapper around a UTF8Nibbler that performs successive character reads, comparing
// each character against a CharacterMatchingFunction. Contigiuous matching or non-matching characters (depending
// on the method) are either placed in a buffer or discarded (depdending on the method).
type UTF8NibblerMatcher struct {
	nibbler UTF8Nibbler
}

// CharacterMatchingFunction is a function that is used by *Matching and *MatchingInto methods. It accepts a rune
// and performs some sort of match against it. It returns true if the rune matches and false otherwise.
type CharacterMatchingFunction func(r rune) (runeMatches bool)

// NewUTF8NibblerMatcher creates a new UTF8Matcher using the provided Nibbler as the Read source.
func NewUTF8NibblerMatcher(nibbler UTF8Nibbler) *UTF8NibblerMatcher {
	return &UTF8NibblerMatcher{
		nibbler: nibbler,
	}
}

// ReadConsecutiveCharactersMatching reads UTF8 characters from the underlying Nibbler. It returns a slice
// containing the consecutive characters from the current Read cursor for which the CharacterMatchingFunction
// returns true. Returns an error if one occurs. If the Nibbler returns EOF on a Read and there were any
// matching characters, returns nil for the error. Otherwise, if the cursor was already at EOF, returns an empty slice
// and io.EOF.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveCharactersMatching(matchFunction CharacterMatchingFunction) ([]rune, error) {
	matchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if len(matchingRunes) == 0 {
					return nil, io.EOF
				}

				return matchingRunes, nil
			}

			return matchingRunes, err
		}

		if matchFunction(nextRune) {
			matchingRunes = append(matchingRunes, nextRune)
		} else {
			matcher.nibbler.UnreadCharacter()
			return matchingRunes, nil
		}
	}
}

// ReadConsecutiveCharactersNotMatching does that same thing as ReadConsecutiveCharactersMatching but returns
// consecutive characters for which the CharacterMatchingFunction returns false.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveCharactersNotMatching(matchFunction CharacterMatchingFunction) ([]rune, error) {
	nonMatchingRunes := make([]rune, 0, 10)

	for {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if len(nonMatchingRunes) == 0 {
					return nil, io.EOF
				}
			}
			return nonMatchingRunes, nil
		}

		if !matchFunction(nextRune) {
			nonMatchingRunes = append(nonMatchingRunes, nextRune)
		} else {
			matcher.nibbler.UnreadCharacter()
			return nonMatchingRunes, nil
		}
	}
}

// ReadConsecutiveCharactersMatchingInto does the same thing as ReadConsecutiveCharactersMatching, but places matching
// characters into receiver. If there are more consecutive characters than the length of receiver, only len(receiver)
// characters are returned, and the Nibbler pointer will be at the next character (even if it also matches). Return the
// number of consecutive matching characters. Return io.EOF only if the Nibbler cursor was already at io.EOF.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveCharactersMatchingInto(matchFunction CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if i == 0 {
					return 0, io.EOF
				}

				return i, nil
			}

			return -1, err
		}

		if matchFunction(nextRune) {
			receiver[i] = nextRune
		} else {
			matcher.nibbler.UnreadCharacter()
			return i, nil
		}
	}

	return len(receiver), nil
}

// ReadConsecutiveCharactersNotMatchingInto does the same thing as ReadConsecutiveCharactersMatchingInto, but adds
// characters for which the CharacterMatchingFunction returns false.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveCharactersNotMatchingInto(matchFunction CharacterMatchingFunction, receiver []rune) (int, error) {
	for i := 0; i < len(receiver); i++ {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if i == 0 {
					return 0, io.EOF
				}

				return i, nil
			}

			return -1, err
		}

		if !matchFunction(nextRune) {
			receiver[i] = nextRune
		} else {
			matcher.nibbler.UnreadCharacter()
			return i, nil
		}
	}

	return len(receiver), nil
}

func runeIsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// ReadConsecutiveWhitespace returns consecutive UTF8 whitespace characters.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveWhitespace() ([]rune, error) {
	return matcher.ReadConsecutiveCharactersMatching(runeIsWhitespace)
}

// ReadConsecutiveWhitespaceInto returns consecutive UTF8 whitepsace characters into the receiver.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveWhitespaceInto(receiver []rune) (int, error) {
	return matcher.ReadConsecutiveCharactersMatchingInto(runeIsWhitespace, receiver)
}

// ReadConsecutiveWordCharacters return consective UTF8 characters that are not whitespace.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveWordCharacters() ([]rune, error) {
	return matcher.ReadConsecutiveCharactersNotMatching(runeIsWhitespace)
}

// ReadConsecutiveWordCharactersInto return consective UTF8 characters that are not whitespace into the receiver.
func (matcher *UTF8NibblerMatcher) ReadConsecutiveWordCharactersInto(receiver []rune) (int, error) {
	return matcher.ReadConsecutiveCharactersNotMatchingInto(runeIsWhitespace, receiver)
}

// DiscardConsecutiveCharactersMatching advances the cursor in the Nibbler until it reaches a character that
// does not match the CharacterMatchingFunction. Return the number of discarded characters.
func (matcher *UTF8NibblerMatcher) DiscardConsecutiveCharactersMatching(matchFunction CharacterMatchingFunction) (int, error) {
	discardedCharacters := 0

	for {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if discardedCharacters == 0 {
					return 0, io.EOF
				}

				return discardedCharacters, nil
			}

			return discardedCharacters, err
		}

		if matchFunction(nextRune) {
			discardedCharacters++
		} else {
			matcher.nibbler.UnreadCharacter()
			return discardedCharacters, nil
		}
	}
}

// DiscardConsecutiveCharactersNotMatching does the same thing as DiscardConsecutiveCharactersMatching but
// advances the cursor through characters for which the CharacterMatchingFunction returns false.
func (matcher *UTF8NibblerMatcher) DiscardConsecutiveCharactersNotMatching(matchFunction CharacterMatchingFunction) (int, error) {
	discardedCharacters := 0

	for {
		nextRune, err := matcher.nibbler.ReadCharacter()
		if err != nil {
			if err == io.EOF {
				if discardedCharacters == 0 {
					return 0, io.EOF
				}

				return discardedCharacters, nil
			}

			return discardedCharacters, err
		}

		if !matchFunction(nextRune) {
			discardedCharacters++
		} else {
			matcher.nibbler.UnreadCharacter()
			return discardedCharacters, nil
		}
	}
}

// DiscardConsecutiveWhitespaceCharacters discards consecutive UTF8 whitespace characters, returning
// the number of discarded characfters.
func (matcher *UTF8NibblerMatcher) DiscardConsecutiveWhitespaceCharacters() (int, error) {
	return matcher.DiscardConsecutiveCharactersMatching(runeIsWhitespace)
}

// DiscardConsecutiveWordCharacters discards consecutive UTF8 characters that are not whitespace,
// returning the number of discarded characters.
func (matcher *UTF8NibblerMatcher) DiscardConsecutiveWordCharacters() (int, error) {
	return matcher.DiscardConsecutiveCharactersNotMatching(runeIsWhitespace)
}

// UnderlyingNibbler returns the UTF8Nibbler used by the matcher.
func (matcher *UTF8NibblerMatcher) UnderlyingNibbler() UTF8Nibbler {
	return matcher.nibbler
}
