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
	return sha256.Sum224(w.Bytes())
}

// Int64StringCheckSum computes the checksum of a string and an int64 value
func Int64StringCheckSum(s string, val int64) [28]byte {
	w := &bytes.Buffer{}
	binary.Write(w, binary.LittleEndian, s)
	binary.Write(w, binary.LittleEndian, val)
	return sha256.Sum224(w.Bytes())
}
