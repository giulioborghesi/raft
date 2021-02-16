package utils

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	testString                = "lorem ipsum dolor sit amet"
	testValue                 = 534208
	testBytesValue            = 8
	testBytesString           = 34
	testInitialBufferCapacity = 1024
)

func TestReadWriteInt64(t *testing.T) {
	// Create buffer used during IO
	b := make([]byte, 0, testInitialBufferCapacity)
	buffer := bytes.NewBuffer(b)

	// Write value to buffer
	w := bufio.NewWriter(buffer)
	bytesWritten, err := WriteValue(w, int64(testValue))

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

	if value != testValue {
		t.Fatalf(unexpectedValueErrFmt, testValue, value)
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
