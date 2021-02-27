package utils

import (
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

func ValidateResult(expected int64, actual int64, t *testing.T) {
	if expected != actual {
		t.Fatalf(valueMismatchErrFmt, expected, actual)
	}
}

func ValidateResults(expected []int64, actual []int64, t *testing.T) {
	if len(expected) != len(actual) {
		t.Fatalf(sizeMismatchErrFmt, len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Fatalf(valueMismatchAtIndexErrFmt, i, expected[i], actual[i])
		}
	}
}

func ValidateLogs(expected []*service.LogEntry, actual []*service.LogEntry,
	t *testing.T) {
	if len(expected) != len(actual) {
		t.Fatalf(sizeMismatchErrFmt, len(expected), len(actual))
	}

	for i := range expected {
		if expected[i].EntryTerm != actual[i].EntryTerm {
			t.Fatalf(valueMismatchAtIndexErrFmt, i, expected[i].EntryTerm,
				actual[i].EntryTerm)
		}

		if expected[i].Payload != actual[i].Payload {
			t.Fatalf(valueMismatchAtIndexErrFmt, i, expected[i].Payload,
				actual[i].Payload)
		}
	}
}
