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
	operation                 string // "Read", "Unread", "Peek", "Whitespace", "Words", "Matching", "NotMatching"
	expectedReadOrPeekRune    rune
	expectedRuneSet           []rune
	expectEOF                 bool
	expectAnErrorThatIsNotEOF bool
	matcherFunction           nibblers.CharacterMatchingFunction
}

func (testCase *utf8NibblerTestCase) testAgainstNibbler(nibbler nibblers.UTF8Nibbler) error {
	switch testCase.operation {
	case "Read":
		nextReadRune, err := nibbler.ReadCaracter()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err != nil {
			return nil
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

		if err != nil {
			return nil
		}

		if peekedRune != testCase.expectedReadOrPeekRune {
			return fmt.Errorf("expected rune (%c) on peek, got (%c)", testCase.expectedReadOrPeekRune, peekedRune)
		}

	case "Whitespace":
		runes, err := nibbler.ReadConsecutiveWhitespace()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err != nil {
			return nil
		}

		if err := compareTwoRuneSlices(testCase.expectedRuneSet, runes); err != nil {
			return err
		}

	case "Words":
		runes, err := nibbler.ReadConsecutiveWordCharacters()
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err != nil {
			return nil
		}

		if err := compareTwoRuneSlices(testCase.expectedRuneSet, runes); err != nil {
			return err
		}

	case "Matching":
		runes, err := nibbler.ReadCharactersMatching(testCase.matcherFunction)
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err != nil {
			return nil
		}

		if err := compareTwoRuneSlices(testCase.expectedRuneSet, runes); err != nil {
			return err
		}

	case "NotMatching":
		runes, err := nibbler.ReadCharactersNotMatching(testCase.matcherFunction)
		if expectationFailure := testCase.testReturnedError(err); expectationFailure != nil {
			return expectationFailure
		}

		if err != nil {
			return nil
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

func matcherFunction1(r rune) bool {
	return r == '\t'
}

func matcherFunction2(r rune) bool {
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

func TestUTF8StringNibbler(t *testing.T) {
	runeString := "∀∁∂∃ ∄ ∅∆∇\t a∉∊  \r    ∋c∍∎\\  +-  ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ \n"

	nibbler := nibblers.NewUTF8StringNibbler(runeString)
	for _, testCase := range []*utf8NibblerTestCase{
		{testname: "Read [1]", operation: "Read", expectedReadOrPeekRune: '∀'},
		{testname: "Peek [1]", operation: "Peek", expectedReadOrPeekRune: '∁'},
		{testname: "Whitespace [1]", operation: "Whitespace", expectedRuneSet: []rune{}},
		{testname: "Unread [1]", operation: "Unread"},
		{testname: "Unread [2]", operation: "Unread", expectAnErrorThatIsNotEOF: true},
		{testname: "Peek [2]", operation: "Peek", expectedReadOrPeekRune: '∀'},
		{testname: "Words [1]", operation: "Words", expectedRuneSet: stringToRuneSlice("∀∁∂∃")},
		{testname: "Peek [3]", operation: "Peek", expectedReadOrPeekRune: ' '},
		{testname: "whitesapce [2]", operation: "Whitespace", expectedRuneSet: []rune{' '}},
		{testname: "Unread [3]", operation: "Unread"},
		{testname: "Peek [4]", operation: "Peek", expectedReadOrPeekRune: ' '},
		{testname: "Words [2]", operation: "Words", expectedRuneSet: []rune{}},
		{testname: "Whitesapce [3]", operation: "Whitespace", expectedRuneSet: []rune{' '}},

		{testname: "Words [2]", operation: "Words", expectedRuneSet: stringToRuneSlice("∄")},
		{testname: "Whitespace [4]", operation: "Whitespace", expectedRuneSet: stringToRuneSlice(" ")},
		{testname: "Words [3]", operation: "Words", expectedRuneSet: stringToRuneSlice("∅∆∇")},
		{testname: "Whitespace [5]", operation: "Whitespace", expectedRuneSet: stringToRuneSlice("\t ")},
		{testname: "Words [4]", operation: "Words", expectedRuneSet: stringToRuneSlice("a∉∊")},
		{testname: "Whitespace [6]", operation: "Whitespace", expectedRuneSet: stringToRuneSlice("  \r    ")},

		{testname: "NotMatching [1]", operation: "NotMatching", expectedRuneSet: stringToRuneSlice("∋c∍∎\\  +-  ∀∁∂∃ ∄ ∅∆∇"), matcherFunction: matcherFunction1},
		{testname: "NotMatching [2]", operation: "NotMatching", expectedRuneSet: stringToRuneSlice(""), matcherFunction: matcherFunction1},

		{testname: "Matching [1]", operation: "Matching", expectedRuneSet: stringToRuneSlice("\t ∈∉∊ "), matcherFunction: matcherFunction2},
		{testname: "Matching [2]", operation: "Matching", expectedRuneSet: stringToRuneSlice(""), matcherFunction: matcherFunction2},

		{testname: "Words [5]", operation: "Words", expectedRuneSet: []rune{'∀', '∁', '∂', '∃'}},
		{testname: "Whitesapce [7]", operation: "Whitespace", expectedRuneSet: []rune{' ', '\n'}},
		{testname: "Words [6]", operation: "Words", expectEOF: true},
		{testname: "Whitesapce [8]", operation: "Whitespace", expectEOF: true},
		{testname: "Peek [5]", operation: "Peek", expectEOF: true},
		{testname: "Read [2]", operation: "Read", expectEOF: true},
		{testname: "Unread [4]", operation: "Unread"},
		{testname: "Peek [6]", operation: "Peek", expectedReadOrPeekRune: '\n'},
		{testname: "Words [7]", operation: "Words", expectedRuneSet: []rune{}},
		{testname: "Whitespace [9]", operation: "Whitespace", expectedRuneSet: []rune{'\n'}},
		{testname: "Peek [7]", operation: "Peek", expectEOF: true},
		{testname: "Read [3]", operation: "Read", expectEOF: true},
	} {
		if expectationFailure := testCase.testAgainstNibbler(nibbler); expectationFailure != nil {
			t.Errorf("[%s] %s", testCase.testname, expectationFailure.Error())
		}
	}
}

type nibbleIntoTestCase struct {
	operation                 string // "Matching", "NotMatching", "Words", "Whitespace" -- all are ...Into
	matcherFunction           nibblers.CharacterMatchingFunction
	expectedRuneSet           []rune
	expectEOF                 bool
	expectAnErrorThatIsNotEOF bool
}

func (testCase *nibbleIntoTestCase) testAgainstNibblerAndReceiver(nibbler nibblers.UTF8Nibbler, receiver []rune) error {
	var runesReadIntoBuffer int
	var err error

	switch testCase.operation {
	case "Matching":
		runesReadIntoBuffer, err = nibbler.ReadCharactersMatchingInto(testCase.matcherFunction, receiver)

	case "NotMatching":
		runesReadIntoBuffer, err = nibbler.ReadCharactersNotMatchingInto(testCase.matcherFunction, receiver)

	case "Words":
		runesReadIntoBuffer, err = nibbler.ReadConsecutiveWordCharactersInto(receiver)

	case "Whitespace":
		runesReadIntoBuffer, err = nibbler.ReadConsecutiveWhitespaceInto(receiver)

	default:
		panic(fmt.Sprintf("test case operation (%s) not known", testCase.operation))
	}

	if err != nil {
		if err == io.EOF {
			if testCase.expectEOF {
				return nil
			}

			return fmt.Errorf("expected error, got EOF")
		}

		if testCase.expectEOF {
			return fmt.Errorf("expected EOF, got error = (%s)", err.Error())
		}

		if !testCase.expectAnErrorThatIsNotEOF {
			return fmt.Errorf("expected no error, got error = (%s)", err.Error())
		}

		return nil
	}

	if len(testCase.expectedRuneSet) != runesReadIntoBuffer {
		return fmt.Errorf("expected %d runes in buffer, got %d", len(testCase.expectedRuneSet), runesReadIntoBuffer)
	}

	if expectationFailure := compareTwoRuneSlices(testCase.expectedRuneSet, receiver[:runesReadIntoBuffer]); expectationFailure != nil {
		return expectationFailure
	}

	return nil
}

func TestUTF8StringNibblerIntoMethods(t *testing.T) {
	testString := "this    \t izz  schön but ∋c∍lylongin∀∁∂strings\r\n ok?"
	nibbler := nibblers.NewUTF8StringNibbler(testString)

	receiver := make([]rune, 5)

	for testCaseIndex, testCase := range []nibbleIntoTestCase{
		{operation: "Whitespace", expectedRuneSet: []rune{}},
		{operation: "Words", expectedRuneSet: stringToRuneSlice("this")},
		{operation: "Words", expectedRuneSet: []rune{}},
		{operation: "Whitespace", expectedRuneSet: stringToRuneSlice("    \t")},
		{operation: "Words", expectedRuneSet: []rune{}},
		{operation: "Whitespace", expectedRuneSet: stringToRuneSlice(" ")},
	} {
		if expectationFailure := testCase.testAgainstNibblerAndReceiver(nibbler, receiver); expectationFailure != nil {
			t.Errorf("[Test %d, %s] %s", testCaseIndex+1, testCase.operation, expectationFailure.Error())
		}
	}

	e := func(r rune) bool {
		switch r {
		case ' ', 'i', 'z':
			return true

		default:
			return false
		}
	}

	f := func(r rune) bool {
		return r == '\r'
	}

	g := func(r rune) bool {
		return r != '"'
	}

	receiver = make([]rune, 10)
	for testCaseIndex, testCase := range []nibbleIntoTestCase{
		{operation: "Whitespace", expectedRuneSet: []rune{}},
		{operation: "Matching", matcherFunction: e, expectedRuneSet: stringToRuneSlice("izz  ")},
		{operation: "NotMatching", matcherFunction: f, expectedRuneSet: stringToRuneSlice("schön but ")},
		{operation: "Words", matcherFunction: f, expectedRuneSet: stringToRuneSlice("∋c∍lylongi")},
		{operation: "Words", matcherFunction: f, expectedRuneSet: stringToRuneSlice("n∀∁∂string")},
		{operation: "NotMatching", matcherFunction: f, expectedRuneSet: stringToRuneSlice("s")},
		{operation: "Matching", matcherFunction: g, expectedRuneSet: stringToRuneSlice("\r\n ok?")},
	} {
		if expectationFailure := testCase.testAgainstNibblerAndReceiver(nibbler, receiver); expectationFailure != nil {
			t.Errorf("[Test %d, %s] %s", testCaseIndex+1, testCase.operation, expectationFailure.Error())
		}
	}

}
