package utils

// Max returns the maximum of two int64 values
func MaxI64(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
