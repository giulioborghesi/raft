package utils

import (
	"encoding/hex"
	"testing"
)

const (
	testEncodedChecksumA = "351d38fe"
	testEncodedChecksumB = "453c31e5"
)

func TestInt64PairCheckSum(t *testing.T) {
	// Compute checksum for two numbers
	checksumA := Int64PairCheckSum(1565, 21)
	encodedA := hex.EncodeToString(checksumA[:])

	if encodedA[:8] != testEncodedChecksumA {
		t.Fatalf(ChecksumFailedErrFmt, testEncodedChecksumA, encodedA[:8])
	}

	// Compute checksum for inverse pair
	checksumB := Int64PairCheckSum(21, 1565)
	encodedB := hex.EncodeToString(checksumB[:])

	if encodedB[:8] != testEncodedChecksumB {
		t.Fatalf(ChecksumFailedErrFmt, testEncodedChecksumB, encodedB[:8])
	}
}
