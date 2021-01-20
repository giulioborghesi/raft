package utils

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
