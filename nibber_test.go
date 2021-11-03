package nibbler_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	mock "github.com/blorticus/go-test-mocks"
	"github.com/blorticus/nibbler"
)

type nibblerExpectedResult struct {
	expectedByte  byte
	expectAnError bool
	expectEOF     bool
}

type nibblerTestCase struct {
	operation      string // "ReadByte", "UnreadByte", "PeekAtNextByte"
	expectedResult *nibblerExpectedResult
}

func (testCase *nibblerTestCase) runTestCaseAgainst(nibber nibbler.ByteNibbler) (testCaseError error) {
	var err error
	var b byte

	switch testCase.operation {
	case "ReadByte":
		b, err = nibber.ReadByte()

	case "PeekAtNextByte":
		b, err = nibber.PeekAtNextByte()

	case "UnreadByte":
		err = nibber.UnreadByte()
	}

	if testCase.expectedResult.expectEOF {
		if err == nil {
			return fmt.Errorf("expected EOF, got no error returned")
		} else if err != io.EOF {
			return fmt.Errorf("expected EOF, got different error returned (%s)", err.Error())
		}
	} else {
		if err == io.EOF {
			return fmt.Errorf("expected no EOF, got EOF")
		}

		if testCase.expectedResult.expectAnError {
			if err == nil {
				return fmt.Errorf("expected an error, no error returned")
			}
		} else if err != nil {
			return fmt.Errorf("expected no error, got an error (%s)", err.Error())
		}
	}

	if !testCase.expectedResult.expectAnError && !testCase.expectedResult.expectEOF {
		if testCase.expectedResult.expectedByte != b {
			return fmt.Errorf("expected byte value (%d), got (%d)", testCase.expectedResult.expectedByte, b)
		}
	}

	return nil
}

func TestByteSliceNibbler(t *testing.T) {
	nib := nibbler.NewByteSliceNibbler([]byte{})

	for testIndex, testCase := range []*nibblerTestCase{
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: true, expectEOF: false}},
	} {
		if err := testCase.runTestCaseAgainst(nib); err != nil {
			t.Errorf("(ByteSliceNibbler with Empty slice) (test %d) %s", testIndex+1, err.Error())
		}

	}

	nib = nibbler.NewByteSliceNibbler([]byte{0, 1, 2, 3, 4, 5})
	for testIndex, testCase := range []*nibblerTestCase{
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: true, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: true, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 1, expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 2, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 2, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 4, expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 4, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
	} {
		if err := testCase.runTestCaseAgainst(nib); err != nil {
			t.Errorf("(ByteSliceNibbler with 6 values in slice) (test %d) %s", testIndex+1, err.Error())
		}

	}
}

func TestByteReaderNibbler(t *testing.T) {
	reader := mock.NewReader().AddGoodRead([]byte{0, 1, 2, 3}).AddGoodRead([]byte{4, 5}).AddEOF()
	nibbler := nibbler.NewByteReaderNibbler(reader)
	for testIndex, testCase := range []*nibblerTestCase{
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: true, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: true, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 1, expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 2, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 2, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 4, expectAnError: false, expectEOF: false}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 3, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 4, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
		{operation: "UnreadByte", expectedResult: &nibblerExpectedResult{expectAnError: false, expectEOF: false}},
		{operation: "PeekAtNextByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 5, expectAnError: false, expectEOF: false}},
		{operation: "ReadByte", expectedResult: &nibblerExpectedResult{expectedByte: 0, expectAnError: false, expectEOF: true}},
	} {
		if err := testCase.runTestCaseAgainst(nibbler); err != nil {
			t.Errorf("(ByteReaderNibbler with 6 values then EOF) (test %d) %s", testIndex+1, err.Error())
		}

	}
}

func TestByteSliceNibblerNamedSet(t *testing.T) {
	byteSliceNibbler := nibbler.NewByteSliceNibbler([]byte{'a', 'b', 'c', ' ', '\t', 'D', '1', '2'})
	testByteNibblerWithNamedSet(byteSliceNibbler, "ByteSliceNibbler", t)

	reader := mock.NewReader().AddGoodRead([]byte{'a', 'b', 'c', ' '}).AddGoodRead([]byte{'\t', 'D', '1'}).AddGoodRead([]byte{'2'}).AddEOF()
	byteReaderNibbler := nibbler.NewByteReaderNibbler(reader)
	testByteNibblerWithNamedSet(byteReaderNibbler, "ByteReaderNibbler", t)

}

func testByteNibblerWithNamedSet(nib nibbler.ByteNibbler, nameOfNibblerType string, t *testing.T) {
	set := nibbler.NewNamedCharacterSetsMap().AddNamedCharacterSetFromString("set-abcdefg", "abcdefg").AddNamedCharacterSetFromByteArray("set-12", []byte{'1', '2'})

	_, err := nib.ReadNextBytesFromSet("set-abc")
	if err == nil {
		t.Errorf("(%s) (test 1) on ReadNextBytesFromSet before adding set: expected error, go no error", nameOfNibblerType)
	}

	nib.UseNamedCharacterSetsMap(set)

	if _, err = nib.ReadNextBytesFromSet("foo"); err == nil {
		t.Errorf("(%s) (test 2) on ReadNextBytesFromSet from non-existent set: expected error, go no error", nameOfNibblerType)
	}

	returnedByteSlice, err := nib.ReadNextBytesFromSet("set-12")
	if err != nil {
		t.Errorf("(%s) (test 3) on ReadNextBytesFromSet: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}

	if len(returnedByteSlice) != 0 {
		t.Errorf("(%s) (test 3) on ReadNextBytesFromSet: expected empty byte array, got (%d) bytes", nameOfNibblerType, len(returnedByteSlice))
	}

	returnedByteSlice, err = nib.ReadNextBytesFromSet("set-abcdefg")
	if err != nil {
		t.Errorf("(%s) (test 4) on ReadNextBytesFromSet: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}

	if bytes.Compare(returnedByteSlice, []byte{'a', 'b', 'c'}) != 0 {
		t.Errorf("(%s) (test 4) on ReadNextBytesFromSet: expected 'abc', got (%s)", nameOfNibblerType, string(returnedByteSlice))
	}

	b, err := nib.ReadByte()
	if err != nil {
		t.Errorf("(%s) (test 5) on first ReadByte: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}
	if b != ' ' {
		t.Errorf("(%s) (test 5) on first ReadByte: expected ' ', got %s", nameOfNibblerType, string(b))
	}

	b, err = nib.ReadByte()
	if err != nil {
		t.Errorf("(%s) (test 5) on second ReadByte: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}
	if b != '\t' {
		t.Errorf("(%s) (test 5) on second ReadByte: expected '\t', got %s", nameOfNibblerType, string(b))
	}

	returnedByteSlice, err = nib.ReadNextBytesFromSet("set-abcdefg")
	if err != nil {
		t.Errorf("(%s) (test 6) on ReadNextBytesFromSet: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}
	if len(returnedByteSlice) > 0 {
		t.Errorf("(%s) (test 6) on ReadNextBytesFromSet: expected empty byte array, got (%d) bytes", nameOfNibblerType, len(returnedByteSlice))
	}

	b, err = nib.ReadByte()
	if err != nil {
		t.Errorf("(%s) (test 7) on ReadByte: expected no error, got error = (%s)", nameOfNibblerType, err.Error())
	}
	if b != 'D' {
		t.Errorf("(%s) (test 7) on ReadByte: expected 'D', got %s", nameOfNibblerType, string(b))
	}

	returnedByteSlice, err = nib.ReadNextBytesFromSet("set-12")
	if err == nil {
		t.Errorf("(%s) (test 8) on ReadNextBytesFromSet: expected EOF, got no error", nameOfNibblerType)
	} else if err != io.EOF {
		t.Errorf("(%s) (test 8) on ReadNextBytesFromSet: expected EOF, got error = (%s)", nameOfNibblerType, err.Error())
	}
	if bytes.Compare(returnedByteSlice, []byte{'1', '2'}) != 0 {
		t.Errorf("(%s) (test 8) on ReadNextBytesFromSet: expected '12', got (%s)", nameOfNibblerType, string(returnedByteSlice))
	}

}
