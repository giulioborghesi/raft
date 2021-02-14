package utils

import "testing"

const (
	// Common error messages shared by packages
	InvalidTermErrFmt     = "invalid term: expected: %d, actual: %d"
	InvalidVotedForErrFmt = "invalid voted for ID: expected: %d, actual: %d"
	ChecksumFailedErrFmt  = "invalid checksum: expected: %s, actual: %s"

	FileNotFoundErrFmt = "file %s not found"
	UnexpectedErrFmt   = "unexpected error: expected: %s, actual: %v"
)

func ValidateResult(t *testing.T, fmt string, expected int64, actual int64) {
	if expected != actual {
		t.Fatalf(fmt, expected, actual)
	}
}
