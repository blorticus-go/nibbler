package nibblers_test

import (
	"fmt"
	"io"
	"testing"

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

func TestUTF8StringNibbler(t *testing.T) {
	runeString := "∀∁∂∃ ∄ ∅∆∇\t ∈∉∊  \r    ∋∌∍∎∏  +-  ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊     ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ \t∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ "

	nibbler := nibblers.NewUTF8StringNibbler(runeString)
	for _, testCase := range []*utf8NibblerTestCase{
		{testname: "First read from string", operation: "Read", expectedReadOrPeekRune: '∀'},
		{testname: "First read whitesapce from string", operation: "Whitespace", expectedRuneSet: []rune{}},
		{testname: "First unread from string", operation: "Unread"},
		{testname: "Second unread from string", operation: "Unread", expectAnErrorThatIsNotEOF: true},
	} {
		if expectationFailure := testCase.testAgainstNibbler(nibbler); expectationFailure != nil {
			t.Errorf("[%s] %s", testCase.testname, expectationFailure.Error())
		}
	}
}
