package utils

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	// Values used during test
	testString = "lorem ipsum dolor sit amet"
	testValueA = 534208
	testValueB = 3495

	// Number of bytes read / written during test
	testBytesChecksum = 28
	testBytesString   = 34
	testBytesValue    = 8

	// Initial buffer capacity
	testInitialBufferCapacity = 1024
)

func TestReadWriteChecksum(t *testing.T) {
	// Create buffer used during IO
	b := make([]byte, 0, testInitialBufferCapacity)
	buffer := bytes.NewBuffer(b)

	// Write value to buffer
	w := bufio.NewWriter(buffer)

	checksum := Int64PairCheckSum(testValueA, testValueB)
	bytesWritten, err := WriteValue(w, checksum)

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if err = w.Flush(); err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesWritten != testBytesChecksum {
		t.Fatalf(unexpectedBytesErrFmt, testBytesChecksum, bytesWritten)
	}

	// Read value from buffer
	r := bufio.NewReader(buffer)
	var actualChecksum [28]byte
	bytesRead, err := ReadValue(r, &actualChecksum)

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesRead != bytesWritten {
		t.Fatalf(unexpectedBytesErrFmt, testBytesValue, bytesWritten)
	}

	if actualChecksum != checksum {
		t.Fatalf(unexpectedValueErrFmt, checksum, actualChecksum)
	}
}
func TestReadWriteInt64(t *testing.T) {
	// Create buffer used during IO
	b := make([]byte, 0, testInitialBufferCapacity)
	buffer := bytes.NewBuffer(b)

	// Write value to buffer
	w := bufio.NewWriter(buffer)
	bytesWritten, err := WriteValue(w, int64(testValueA))

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if err = w.Flush(); err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesWritten != testBytesValue {
		t.Fatalf(unexpectedBytesErrFmt, testBytesValue, bytesWritten)
	}

	// Read value from buffer
	r := bufio.NewReader(buffer)
	var value int64
	bytesRead, err := ReadValue(r, &value)

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesRead != bytesWritten {
		t.Fatalf(unexpectedBytesErrFmt, testBytesValue, bytesWritten)
	}

	if value != testValueA {
		t.Fatalf(unexpectedValueErrFmt, testValueA, value)
	}
}

func TestReadWriteString(t *testing.T) {
	// Create buffer used during IO
	b := make([]byte, 0, testInitialBufferCapacity)
	buffer := bytes.NewBuffer(b)

	// Write string to buffer
	w := bufio.NewWriter(buffer)
	bytesWritten, err := WriteString(w, testString)

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if err = w.Flush(); err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesWritten != testBytesString {
		t.Fatalf(unexpectedBytesErrFmt, testBytesString, bytesWritten)
	}

	// Read string from buffer
	r := bufio.NewReader(buffer)
	s, bytesRead, err := ReadString(r)

	if err != nil {
		t.Fatalf(unexpectedErrFmt, err)
	}

	if bytesRead != bytesWritten {
		t.Fatalf(unexpectedBytesErrFmt, testBytesString, bytesWritten)
	}

	if s != testString {
		t.Fatalf(unexpectedValueErrFmt, testString, s)
	}
}
