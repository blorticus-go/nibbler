package nibblers_test

import (
	"fmt"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/blorticus/nibblers"
)

type utf8NibblerTestCase struct {
	testname                  string
	operation                 string // "Read", "Unread", "Peek", "Whitespace", "Words"
	expectedReadOrPeekRune    rune
	expectedRuneSet           []rune
	expectEOF                 bool
	expectAnErrorThatIsNotEOF bool
}

func (testCase *utf8NibblerTestCase) testAgainstNibbler(nibbler nibblers.UTF8Nibbler) error {
	switch testCase.operation {
	case "Read":
		nextReadRune, err := nibbler.ReadCaracter()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if nextReadRune != testCase.expectedReadOrPeekRune {
			return fmt.Errorf("expected rune (%c), got (%c)", testCase.expectedReadOrPeekRune, nextReadRune)
		}

	case "Unread":
		err := nibbler.UnreadCharacter()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

	case "Peek":
		peekedRune, err := nibbler.PeekAtNextCharacter()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if peekedRune != testCase.expectedReadOrPeekRune {
			return fmt.Errorf("expected rune (%c) on peek, got (%c)", testCase.expectedReadOrPeekRune, peekedRune)
		}

	case "Whitespace":
		runes, err := nibbler.ReadConsecutiveWhitespace()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err := compareTwoRuneSlices(testCase.expectedRuneSet, runes); err != nil {
			return err
		}

	case "Words":
		runes, err := nibbler.ReadConsecutiveWordCharacters()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err := compareTwoRuneSlices(testCase.expectedRuneSet, runes); err != nil {
			return err
		}

	default:
		panic(fmt.Sprintf("unexpected testCase action: %s", testCase.operation))
	}

	return nil
}

func (testCase *utf8NibblerTestCase) testReturnedError(err error) error {
	if testCase.expectEOF {
		if err == nil {
			return fmt.Errorf("expected EOF, got no returned error")
		} else if err != io.EOF {
			return fmt.Errorf("expected EOF, got error = (%s)", err.Error())
		}

		return nil
	}

	if testCase.expectAnErrorThatIsNotEOF {
		if err == nil {
			return fmt.Errorf("expected error, got no returned error")
		} else if err == io.EOF {
			return fmt.Errorf("expected error, got EOF")
		}

		return nil
	}

	if err == io.EOF {
		return fmt.Errorf("expected no error or EOF, got EOF")
	}

	if err != nil {
		return fmt.Errorf("expected no error or EOF, got error = (%s)", err.Error())
	}

	return nil
}

func compareTwoRuneSlices(expectedRunes []rune, gotRunes []rune) error {
	if len(expectedRunes) != len(gotRunes) {
		return fmt.Errorf("expected %d runes, got %d", len(expectedRunes), len(gotRunes))
	}

	for i, expectedRune := range expectedRunes {
		if gotRunes[i] != expectedRune {
			return fmt.Errorf("at index %d expected rune (%c), got (%c)", i, expectedRune, gotRunes[i])
		}
	}

	return nil
}

func stringToRuneSlice(s string) []rune {
	r := make([]rune, 0, utf8.RuneCountInString(s))

	for i := 0; i < len(s); {
		nextRune, runeLengthInBytes := utf8.DecodeRuneInString(s[i:])
		i += runeLengthInBytes
		r = append(r, nextRune)
	}

	return r
}

func TestUTF8StringNibbler(t *testing.T) {
	runeString := "∀∁∂∃ ∄ ∅∆∇\t a∉∊  \r    ∋c∍∎\\  +-  ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ \n"

	nibbler := nibblers.NewUTF8StringNibbler(runeString)
	for _, testCase := range []*utf8NibblerTestCase{
		{testname: "First read from string", operation: "Read", expectedReadOrPeekRune: '∀'},
		{testname: "First peek from string", operation: "Peek", expectedReadOrPeekRune: '∁'},
		{testname: "First read whitesapce from string", operation: "Whitespace", expectedRuneSet: []rune{}},
		{testname: "First unread from string", operation: "Unread"},
		{testname: "Second unread from string", operation: "Unread", expectAnErrorThatIsNotEOF: true},
		{testname: "Second peek from string", operation: "Peek", expectedReadOrPeekRune: '∀'},
		{testname: "First read words from string", operation: "Words", expectedRuneSet: []rune{'∀', '∁', '∂', '∃'}},
		{testname: "Third peek from string", operation: "Peek", expectedReadOrPeekRune: ' '},
		{testname: "Second read whitesapce from string", operation: "Whitespace", expectedRuneSet: []rune{' '}},
		{testname: "Third unread from string", operation: "Unread"},
		{testname: "Fourth peek from string", operation: "Peek", expectedReadOrPeekRune: ' '},
		{testname: "Second read words from string", operation: "Words", expectedRuneSet: []rune{}},
		{testname: "Third read whitesapce from string", operation: "Whitespace", expectedRuneSet: []rune{' '}},
	} {
		if expectationFailure := testCase.testAgainstNibbler(nibbler); expectationFailure != nil {
			t.Errorf("[%s] %s", testCase.testname, expectationFailure.Error())
		}
	}

	for testIndex, expectedWordsThenSpaces := range [][][]rune{
		{{'∄'}, {' '}},
		{{'∅', '∆', '∇'}, {'\t', ' '}},
		{{'a', '∉', '∊'}, {' ', ' ', '\r', ' ', ' ', ' ', ' '}},
	} {
		expectedWords := expectedWordsThenSpaces[0]
		expectedWhitespace := expectedWordsThenSpaces[1]

		runes, err := nibbler.ReadConsecutiveWordCharacters()
		if err != nil {
			t.Errorf("[Alternating word/space Test Index %d] on ReadConsecutiveWorCharacters() expected no error, got err = (%s)", testIndex, err.Error())
		}

		if err := compareTwoRuneSlices(expectedWords, runes); err != nil {
			t.Errorf("[Alternating word/space Test Index %d] on ReadConsecutiveWorCharacters() %s", testIndex, err.Error())
		}

		runes, err = nibbler.ReadConsecutiveWhitespace()
		if err != nil {
			t.Errorf("[Alternating word/space Test Index %d] on ReadConsecutiveWhitespace() expected no error, got err = (%s)", testIndex, err.Error())
		}

		if err := compareTwoRuneSlices(expectedWhitespace, runes); err != nil {
			t.Errorf("[Alternating word/space Test Index %d] on ReadConsecutiveWhitespace() %s", testIndex, err.Error())
		}
	}

	f := func(r rune) bool {
		return r == '\t'
	}

	runes, err := nibbler.ReadCharactersNotMatching(f)
	if err != nil {
		t.Errorf("[ReadCharactersNotMatching(\\t)] expected no error, got error = (%s)", err.Error())
	}

	if err := compareTwoRuneSlices(stringToRuneSlice("∋c∍∎\\  +-  ∀∁∂∃ ∄ ∅∆∇"), runes); err != nil {
		t.Errorf("[ReadCharactersNotMatching(\\t)] %s", err.Error())
	}

	runes, err = nibbler.ReadCharactersNotMatching(f)
	if err != nil {
		t.Errorf("[Second ReadCharactersNotMatching(\\t)] expected no error, got error = (%s)", err.Error())
	}

	if err := compareTwoRuneSlices([]rune{}, runes); err != nil {
		t.Errorf("[Second ReadCharactersNotMatching(\\t)] %s", err.Error())
	}

	f = func(r rune) bool {
		switch r {
		case ' ':
			return true
		case '∈':
			return true
		case '∉':
			return true
		case '∊':
			return true
		case 't':
			return true
		case '\r':
			return true
		case '\t':
			return true
		default:
			return false
		}
	}

	runes, err = nibbler.ReadCharactersMatching(f)
	if err != nil {
		t.Errorf("[ReadCharactersMatching()] expected no error, got error = (%s)", err.Error())
	}

	if err := compareTwoRuneSlices(stringToRuneSlice("\t ∈∉∊ "), runes); err != nil {
		t.Errorf("[ReadCharactersMatching()] %s", err.Error())
	}

	runes, err = nibbler.ReadCharactersMatching(f)
	if err != nil {
		t.Errorf("[Second ReadCharactersMatching()] expected no error, got error = (%s)", err.Error())
	}

	if err := compareTwoRuneSlices([]rune{}, runes); err != nil {
		t.Errorf("[Second ReadCharactersMatching()] %s", err.Error())
	}

	for _, testCase := range []*utf8NibblerTestCase{
		{testname: "Third read words from string", operation: "Words", expectedRuneSet: []rune{'∀', '∁', '∂', '∃'}},
		{testname: "Fourth read whitesapce from string", operation: "Whitespace", expectedRuneSet: []rune{' ', '\n'}},
	} {
		if expectationFailure := testCase.testAgainstNibbler(nibbler); expectationFailure != nil {
			t.Errorf("[%s] %s", testCase.testname, expectationFailure.Error())
		}
	}

}
