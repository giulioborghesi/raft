package datasources

const (
	corruptedLogFilesErrFmt = "log files corrupted: %s and %s"
	invalidLogEntryIndexErr = "log entry index invalid"

	// Unit test errors
	bytesMismatchErrFmt  = "bytes mismatch: expected: %d, actual: %d"
	offsetMismatchErrFmt = "offsets mismatch at index %d: expected: %d, actual: %d"
	entryMismatchErrFmt  = "entry mismatch at index %d: expected: %v, actual: %v"
)
