package nibblers_test

import (
	"fmt"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/blorticus-go/nibblers"
	mock "github.com/blorticus/go-test-mocks"
)

func TestUTF8StringNibbler(t *testing.T) {
	testUTF8NibblerExceptIntoFunctionsUsingType("String", t)
	testUTF8StringNibblerIntoMethodsUsingType("String", t)
}

func TestUTF8RuneSliceNibbler(t *testing.T) {
	testUTF8NibblerExceptIntoFunctionsUsingType("RuneSlice", t)
	testUTF8StringNibblerIntoMethodsUsingType("RuneSlice", t)
}

func TestUTF8ByteSliceNibbler(t *testing.T) {
	testUTF8NibblerExceptIntoFunctionsUsingType("ByteSlice", t)
	testUTF8StringNibblerIntoMethodsUsingType("ByteSlice", t)
}

func TestUTF8ReaderNibbler(t *testing.T) {
	testUTF8NibblerExceptIntoFunctionsUsingType("Reader", t)
	testUTF8StringNibblerIntoMethodsUsingType("Reader", t)
}

type utf8MatcherDiscardsTestCase struct {
	typeOfDiscard                       string // "Whitespace", "Word", "spaces", "ascii", "notAsciiAlpha"
	expectedNumberOfDiscardedCharacters int
	expectToBeAtEOFAfterDiscards        bool
	expectedNextCharacterAfterDiscard   rune
	expectEOF                           bool
}

func asciiSpaceMatcher(c rune) bool {
	return c == ' '
}

func asciiCharacterMatcher(c rune) bool {
	return c > 0 && c < 128
}

func asciiAlphaMatcher(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (testCase *utf8MatcherDiscardsTestCase) testAgainstMatcher(matcher *nibblers.UTF8NibblerMatcher) error {
	var discardError error
	var numberOfDiscardedCharacters int

	switch testCase.typeOfDiscard {
	case "Whitespace":
		numberOfDiscardedCharacters, discardError = matcher.DiscardConsecutiveWhitespaceCharacters()

	case "Word":
		numberOfDiscardedCharacters, discardError = matcher.DiscardConsecutiveWordCharacters()

	case "spaces":
		numberOfDiscardedCharacters, discardError = matcher.DiscardConsecutiveCharactersMatching(asciiSpaceMatcher)

	case "ascii":
		numberOfDiscardedCharacters, discardError = matcher.DiscardConsecutiveCharactersMatching(asciiCharacterMatcher)

	case "notAsciiAlpha":
		numberOfDiscardedCharacters, discardError = matcher.DiscardConsecutiveCharactersNotMatching(asciiAlphaMatcher)

	default:
		return fmt.Errorf("invalid test case type (%s) provided", testCase.typeOfDiscard)
	}

	if testCase.expectEOF {
		if discardError == nil {
			return fmt.Errorf("expected io.EOF, got no error")
		}

		if discardError != io.EOF {
			return fmt.Errorf("expected io.EOF, got error = (%s)", discardError.Error())
		}

		return nil
	}

	if discardError != nil {
		if discardError == io.EOF {
			return fmt.Errorf("did not expect io.EOF, but got it")
		}

		return fmt.Errorf("did not expect error, but got error = (%s)", discardError.Error())
	}

	if numberOfDiscardedCharacters != testCase.expectedNumberOfDiscardedCharacters {
		return fmt.Errorf("expected (%d) discarded characters, got (%d)", testCase.expectedNumberOfDiscardedCharacters, numberOfDiscardedCharacters)
	}

	characterAfterDiscards, err := matcher.UnderlyingNibbler().PeekAtNextCharacter()
	if err != nil {
		if err == io.EOF {
			if !testCase.expectToBeAtEOFAfterDiscards {
				return fmt.Errorf("expected no io.EOF on peek after discards, but got it")
			}

			return nil
		}

		return fmt.Errorf("expected no error on peek after discards, but got error = (%s)", err.Error())
	}

	if testCase.expectToBeAtEOFAfterDiscards {
		return fmt.Errorf("expected io.EOF on peek after discards, but got none")
	}

	if characterAfterDiscards != testCase.expectedNextCharacterAfterDiscard {
		return fmt.Errorf("expected (%c) on peek after discards, but got (%c)", testCase.expectedNextCharacterAfterDiscard, characterAfterDiscards)
	}

	return nil
}

func TestUTF8NibblerDiscards(t *testing.T) {
	runeString := "∀∁∂∃ ∄ ∅∆∇\t z∉∊  \r    ∋∍∎\\c  \t \r\n+-∀abc∃"

	nibbler := nibblers.NewUTF8StringNibbler(runeString)
	matcher := nibblers.NewUTF8NibblerMatcher(nibbler)

	for testCaseIndex, testCase := range []*utf8MatcherDiscardsTestCase{
		{
			typeOfDiscard:                       "Whitespace",
			expectedNumberOfDiscardedCharacters: 0,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   '∀',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "ascii",
			expectedNumberOfDiscardedCharacters: 0,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   '∀',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Word",
			expectedNumberOfDiscardedCharacters: 4,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   ' ',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Whitespace",
			expectedNumberOfDiscardedCharacters: 1,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   '∄',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "notAsciiAlpha",
			expectedNumberOfDiscardedCharacters: 7,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   'z',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Whitespace",
			expectedNumberOfDiscardedCharacters: 0,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   'z',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "ascii",
			expectedNumberOfDiscardedCharacters: 1,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   '∉',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "notAsciiAlpha",
			expectedNumberOfDiscardedCharacters: 13,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   'c',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Word",
			expectedNumberOfDiscardedCharacters: 1,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   ' ',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Whitespace",
			expectedNumberOfDiscardedCharacters: 6,
			expectToBeAtEOFAfterDiscards:        false,
			expectedNextCharacterAfterDiscard:   '+',
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Word",
			expectedNumberOfDiscardedCharacters: 7,
			expectToBeAtEOFAfterDiscards:        true,
			expectEOF:                           false,
		},
		{
			typeOfDiscard:                       "Word",
			expectedNumberOfDiscardedCharacters: 0,
			expectEOF:                           true,
		},
	} {
		if err := testCase.testAgainstMatcher(matcher); err != nil {
			t.Errorf("on test case %d: %s", testCaseIndex+1, err.Error())
		}
	}
}

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
	matcher := nibblers.NewUTF8NibblerMatcher(nibbler)

	switch testCase.operation {
	case "Read":
		nextReadRune, err := nibbler.ReadCharacter()
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
		runes, err := matcher.ReadConsecutiveWhitespace()
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
		runes, err := matcher.ReadConsecutiveWordCharacters()
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
		runes, err := matcher.ReadConsecutiveCharactersMatching(testCase.matcherFunction)
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
		runes, err := matcher.ReadConsecutiveCharactersNotMatching(testCase.matcherFunction)
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

func testUTF8NibblerExceptIntoFunctionsUsingType(typeOfNibbler string, t *testing.T) {
	runeString := "∀∁∂∃ ∄ ∅∆∇\t a∉∊  \r    ∋c∍∎\\  +-  ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ \n"

	var nibbler nibblers.UTF8Nibbler

	switch typeOfNibbler {
	case "String":
		nibbler = nibblers.NewUTF8StringNibbler(runeString)
	case "RuneSlice":
		nibbler = nibblers.NewUTF8RuneSliceNibbler(stringToRuneSlice(runeString))
	case "ByteSlice":
		nibbler = nibblers.NewUTF8ByteSliceNibbler([]byte(runeString))
	case "Reader":
		reader := mock.NewReader().
			AddGoodRead([]byte(runeString[:9])).
			AddGoodRead([]byte(runeString[9:10])).
			AddGoodRead([]byte(runeString[10:11])).
			AddGoodRead([]byte(runeString[11:28])).
			AddGoodRead([]byte(runeString[28:])).AddEOF()
		nibbler = nibblers.NewUTF8ReaderNibbler(reader)
	default:
		panic(fmt.Sprintf("invalid typeOfNibbler (%s) for testUTF8NibblerExceptIntoFunctionsUsingType", typeOfNibbler))
	}

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

	matcher := nibblers.NewUTF8NibblerMatcher(nibbler)

	switch testCase.operation {
	case "Matching":
		runesReadIntoBuffer, err = matcher.ReadConsecutiveCharactersMatchingInto(testCase.matcherFunction, receiver)

	case "NotMatching":
		runesReadIntoBuffer, err = matcher.ReadConsecutiveCharactersNotMatchingInto(testCase.matcherFunction, receiver)

	case "Words":
		runesReadIntoBuffer, err = matcher.ReadConsecutiveWordCharactersInto(receiver)

	case "Whitespace":
		runesReadIntoBuffer, err = matcher.ReadConsecutiveWhitespaceInto(receiver)

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

func testUTF8StringNibblerIntoMethodsUsingType(typeOfNibbler string, t *testing.T) {
	runeString := "this    \t izz  schön but ∋c∍lylongin∀∁∂strings\r\n ok?"

	var nibbler nibblers.UTF8Nibbler

	switch typeOfNibbler {
	case "String":
		nibbler = nibblers.NewUTF8StringNibbler(runeString)
	case "RuneSlice":
		nibbler = nibblers.NewUTF8RuneSliceNibbler(stringToRuneSlice(runeString))
	case "ByteSlice":
		nibbler = nibblers.NewUTF8ByteSliceNibbler([]byte(runeString))
	case "Reader":
		reader := mock.NewReader().
			AddGoodRead([]byte(runeString[:27])).
			AddGoodRead([]byte(runeString[27:47])).
			AddGoodRead([]byte(runeString[47:])).AddEOF()
		nibbler = nibblers.NewUTF8ReaderNibbler(reader)

	default:
		panic(fmt.Sprintf("invalid typeOfNibbler (%s) for testUTF8StringNibblerIntoMethodsUsingType", typeOfNibbler))
	}

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
		{operation: "NotMatching", matcherFunction: f, expectEOF: true},
		{operation: "Matching", matcherFunction: g, expectEOF: true},
	} {
		if expectationFailure := testCase.testAgainstNibblerAndReceiver(nibbler, receiver); expectationFailure != nil {
			t.Errorf("[Test %d, %s] %s", testCaseIndex+1, testCase.operation, expectationFailure.Error())
		}
	}

}

type BookendTester struct {
	nibbler                                      nibblers.UTF8Nibbler
	methodForCreatingTheNibblerThatUsesTheString func(s string) nibblers.UTF8Nibbler
	pendingError                                 error
	pendingRuneSet                               []rune
}

func NewBookendTester(methodForCreatingTheNibblerThatUsesTheString func(s string) nibblers.UTF8Nibbler) *BookendTester {
	return &BookendTester{
		methodForCreatingTheNibblerThatUsesTheString: methodForCreatingTheNibblerThatUsesTheString,
	}
}

func (tester *BookendTester) createNibblerWithString(s string) {
	tester.nibbler = tester.methodForCreatingTheNibblerThatUsesTheString(s)
}

type BookendTestStep func(nibblers.UTF8Nibbler) ([]rune, error)

func (tester *BookendTester) After(steps []BookendTestStep) *BookendTester {
	for _, step := range steps {
		runeSet, err := step(tester.nibbler)
		if err != nil {
			tester.pendingError = err

			if err != io.EOF {
				return tester
			}
		}

		if len(runeSet) > 0 {
			tester.pendingRuneSet = runeSet
		}
	}

	return tester
}

func (tester *BookendTester) ExpectTheRuneSet(expectedRuneSet []rune) error {
	if tester.pendingError != nil {
		e := tester.pendingError
		tester.pendingError = nil
		tester.pendingRuneSet = nil
		return e
	}

	if string(tester.pendingRuneSet) != string(expectedRuneSet) {
		gotRuneSet := tester.pendingRuneSet
		tester.pendingError = nil
		tester.pendingRuneSet = nil
		return fmt.Errorf("expected the rune set (%s), got (%s)", string(expectedRuneSet), string(gotRuneSet))
	}

	return nil
}

func (tester *BookendTester) ExpectEOF() error {
	if tester.pendingError == nil {
		return fmt.Errorf("expected EOF, got no EOF")
	}

	if tester.pendingError != io.EOF {
		return fmt.Errorf("expected EOF, got (%s)", tester.pendingError.Error())
	}

	return nil
}

type BookendTesterChain struct {
	tester               *BookendTester
	lastExpectationError error
}

func (tester *BookendTester) Expect() *BookendTesterChain {
	return &BookendTesterChain{
		tester:               tester,
		lastExpectationError: nil,
	}
}

func (chain *BookendTesterChain) Failed() bool {
	return chain.lastExpectationError != nil
}

func (chain *BookendTesterChain) Error() string {
	return chain.lastExpectationError.Error()
}

func (chain *BookendTesterChain) TheRuneSet(expectedRuneSet []rune) *BookendTesterChain {
	if chain.tester.pendingRuneSet == nil || len(chain.tester.pendingRuneSet) == 0 {
		if expectedRuneSet == nil || len(expectedRuneSet) == 0 {
			return chain
		} else {
			chain.lastExpectationError = fmt.Errorf("expected the rune set (%s), got an empty rune set", string(expectedRuneSet))
			return chain
		}
	}

	if string(chain.tester.pendingRuneSet) != string(expectedRuneSet) {
		chain.lastExpectationError = fmt.Errorf("expected the rune set (%s), got (%s)", string(expectedRuneSet), string(chain.tester.pendingRuneSet))
		return chain
	}

	return chain
}

func (chain *BookendTesterChain) And() *BookendTesterChain {
	return chain
}

func (chain *BookendTesterChain) TheEOFIndicator() *BookendTesterChain {
	if chain.tester.pendingError == nil {
		chain.lastExpectationError = fmt.Errorf("expected EOF, got no error")
	} else if chain.tester.pendingError != io.EOF {
		chain.lastExpectationError = fmt.Errorf("expected EOF, got the error (%s)", chain.tester.pendingError.Error())
	}

	return chain
}

func StartingABookend(nibbler nibblers.UTF8Nibbler) ([]rune, error) {
	err := nibbler.StartBookending()
	return nil, err
}

func StoppingTheBookend(nibbler nibblers.UTF8Nibbler) ([]rune, error) {
	return nibbler.StopBookending()
}

type ReadCharacterIntermediate struct {
	Character  BookendTestStep
	Characters BookendTestStep
}

func Reading(numberOfCharactersRead int) *ReadCharacterIntermediate {
	f := func(nibbler nibblers.UTF8Nibbler) ([]rune, error) {
		for i := 0; i < numberOfCharactersRead; i++ {
			if _, err := nibbler.ReadCharacter(); err != nil {
				return nil, err
			}
		}

		return nil, nil
	}

	return &ReadCharacterIntermediate{
		Character:  f,
		Characters: f,
	}
}

type PeekCharacterIntermediate struct {
	Character  BookendTestStep
	Characters BookendTestStep
}

func PeekingAt(numberOfPeeks int) *PeekCharacterIntermediate {
	f := func(nibbler nibblers.UTF8Nibbler) ([]rune, error) {
		for i := 0; i < numberOfPeeks; i++ {
			if _, err := nibbler.PeekAtNextCharacter(); err != nil {
				return nil, err
			}
		}

		return nil, nil
	}

	return &PeekCharacterIntermediate{
		Character:  f,
		Characters: f,
	}
}

func TestUTF8NibblerBookending(t *testing.T) {
	tester := NewBookendTester(func(nibblerString string) nibblers.UTF8Nibbler {
		return nibblers.NewUTF8StringNibbler(nibblerString)
	})

	if err := bookendTestsForUTF8Nibblers(tester); err != nil {
		t.Errorf(err.Error())
	}

	tester = NewBookendTester(func(nibblerString string) nibblers.UTF8Nibbler {
		reader := mock.NewReader().AddGoodRead([]byte(nibblerString)).AddEOF()
		return nibblers.NewUTF8ReaderNibbler(reader)
	})

	if err := bookendTestsForUTF8Nibblers(tester); err != nil {
		t.Errorf(err.Error())
	}
}

func bookendTestsForUTF8Nibblers(tester *BookendTester) error {
	tester.createNibblerWithString("∋c∍lylongi schön but \r\n ok? おはよう")

	theExpectation := tester.After([]BookendTestStep{
		StartingABookend,
		StoppingTheBookend,
	}).Expect().TheRuneSet([]rune{})

	if theExpectation.Failed() {
		return fmt.Errorf("on bookend start and stop at beginnging: %s", theExpectation.Error())
	}

	theExpectation = tester.After([]BookendTestStep{
		StartingABookend,
		Reading(8).Characters,
		StoppingTheBookend,
	}).Expect().TheRuneSet([]rune("∋c∍lylon"))

	if theExpectation.Failed() {
		return fmt.Errorf("on bookending start through character 8: %s", theExpectation.Error())
	}

	theExpectation = tester.After([]BookendTestStep{
		Reading(4).Characters,
		StartingABookend,
		Reading(4).Characters,
		PeekingAt(1).Character,
		StoppingTheBookend,
	}).Expect().TheRuneSet([]rune("chön"))

	if theExpectation.Failed() {
		return fmt.Errorf("on bookending characters 12 through 16 followed by a peek: %s", theExpectation.Error())
	}

	theExpectations := tester.After([]BookendTestStep{
		PeekingAt(1).Character,
		Reading(4).Characters,
		PeekingAt(1).Character,
		StartingABookend,
		Reading(8).Characters,
		PeekingAt(1).Character,
		Reading(8).Characters,
		StoppingTheBookend,
	}).Expect().TheEOFIndicator().And().TheRuneSet([]rune(" \r\n ok? おはよう"))
	if theExpectations.Failed() {
		return fmt.Errorf("on bookending characters 16 through the end followed by a peek: %s", theExpectations.Error())
	}

	tester.createNibblerWithString("∋c∍lylongi schön but \r\n ok? おはよう")
	nibbler := tester.nibbler

	runeSets := make([][]rune, 8)
	errors := make([]error, 8)

	if err := nibbler.StartBookending(); err != nil {
		return fmt.Errorf("on StartBookending(): %s", err.Error())
	}

	runeSets[0], errors[0] = nibbler.CharactersSinceStartOfBookend()

	if err := readCharactersFromNibbler(nibbler, 4); err != nil {
		return fmt.Errorf("on reading characters [1], error = (%s)", err.Error())
	}
	runeSets[1], errors[1] = nibbler.CharactersSinceStartOfBookend()
	runeSets[2], errors[2] = nibbler.CharactersSinceStartOfBookend()

	if err := readCharactersFromNibbler(nibbler, 4); err != nil {
		return fmt.Errorf("on reading characters [2], error = (%s)", err.Error())
	}
	runeSets[3], errors[3] = nibbler.CharactersSinceStartOfBookend()

	if err := readCharactersFromNibbler(nibbler, 8); err != nil {
		return fmt.Errorf("on reading characters [3], error = (%s)", err.Error())
	}
	runeSets[4], errors[4] = nibbler.CharactersSinceStartOfBookend()

	if err := readCharactersFromNibbler(nibbler, 16); err != nil {
		return fmt.Errorf("on reading characters [4], error = (%s)", err.Error())
	}
	runeSets[5], errors[5] = nibbler.CharactersSinceStartOfBookend()

	if err := readCharactersFromNibbler(nibbler, 1); err != nil && err != io.EOF {
		return fmt.Errorf("on reading characters [5], error = (%s)", err.Error())
	}
	runeSets[6], errors[6] = nibbler.CharactersSinceStartOfBookend()

	runeSets[7], errors[7] = nibbler.StopBookending()

	for _, err := range errors {
		if err != nil {
			return fmt.Errorf("expected no error, got (%s)", err.Error())
		}
	}

	for i, expectedRuneSliceString := range []string{"", "∋c∍l", "∋c∍l", "∋c∍lylon", "∋c∍lylongi schön", "∋c∍lylongi schön but \r\n ok? おはよう", "∋c∍lylongi schön but \r\n ok? おはよう", "∋c∍lylongi schön but \r\n ok? おはよう"} {
		if string(runeSets[i]) != expectedRuneSliceString {
			return fmt.Errorf("expected runeSet (%s), got (%s)", expectedRuneSliceString, string(runeSets[i]))
		}
	}

	return nil
}

func readCharactersFromNibbler(nibbler nibblers.UTF8Nibbler, numberToRead int) error {
	for i := 0; i < numberToRead; i++ {
		if _, err := nibbler.ReadCharacter(); err != nil {
			return err
		}
	}

	return nil
}
