package utils

import (
	"bufio"
	"encoding/binary"
)

// ReadString reads a string from a buffered io.Reader object
func ReadString(r *bufio.Reader) (string, int64, error) {
	// Read string length
	var rBytes int64
	if err := binary.Read(r, binary.LittleEndian, &rBytes); err != nil {
		return "", 0, err
	}

	// Read string
	s := make([]byte, rBytes)
	if err := binary.Read(r, binary.LittleEndian, &s); err != nil {
		return "", 0, err
	}

	// Update bytes read and return
	rBytes += int64(binary.Size(rBytes))
	return string(s), rBytes, nil
}

// ReadValue reads a value from a buffered io.Reader object
func ReadValue(r *bufio.Reader, v interface{}) (int64, error) {
	if err := binary.Read(r, binary.LittleEndian, v); err != nil {
		return 0, err
	}
	rBytes := int64(binary.Size(v))
	return rBytes, nil
}

// WriteString writes a string to a buffered io.Writer object
func WriteString(w *bufio.Writer, s string) (int64, error) {
	// Convert string to byte slice
	bs := []byte(s)

	// Write string length
	sBytes := int64(binary.Size(bs))
	if err := binary.Write(w, binary.LittleEndian, sBytes); err != nil {
		return 0, err
	}

	// Write string
	if err := binary.Write(w, binary.LittleEndian, bs); err != nil {
		return 0, err
	}

	// Update bytes written and return
	sBytes += int64(binary.Size(sBytes))
	return sBytes, nil
}

// WriteValue writes a generic value to a buffered io.Writer object
func WriteValue(w *bufio.Writer, value interface{}) (int64, error) {
	wBytes := int64(binary.Size(value))
	if err := binary.Write(w, binary.LittleEndian, value); err != nil {
		return 0, err
	}
	return wBytes, nil
}
