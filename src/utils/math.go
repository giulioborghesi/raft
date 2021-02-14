package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

// Max returns the maximum of two int64 values
func MaxInt64(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

// Max returns the maximum of two int64 values
func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Int64PairCheckSum computes the checksum of a pair of int64 values
func Int64PairCheckSum(lhs int64, rhs int64) [28]byte {
	w := &bytes.Buffer{}
	binary.Write(w, binary.LittleEndian, lhs)
	binary.Write(w, binary.LittleEndian, rhs)
	checkSum := sha256.Sum224(w.Bytes())
	return checkSum
}
