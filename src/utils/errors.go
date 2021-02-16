package utils

import "testing"

const (
	// Common error messages shared by packages
	InvalidTermErrFmt     = "invalid term: expected: %d, actual: %d"
	InvalidVotedForErrFmt = "invalid voted for ID: expected: %d, actual: %d"
	ChecksumFailedErrFmt  = "invalid checksum: expected: %s, actual: %s"
	FileNotFoundErrFmt    = "file %s not found"
	UnexpectedErrFmt      = "unexpected error: expected: %v, actual: %v"

	// Local error messages
	unexpectedErrFmt      = "unexpected error: %v"
	unexpectedBytesErrFmt = "unexpected number of bytes read / written: expected: %d, actual: %d"
	unexpectedValueErrFmt = "unexpected value: expected: %v, actual: %v"
	unsupportedValueErr   = "error: value type not supported"
)

func ValidateResult(t *testing.T, fmt string, expected int64, actual int64) {
	if expected != actual {
		t.Fatalf(fmt, expected, actual)
	}
}
