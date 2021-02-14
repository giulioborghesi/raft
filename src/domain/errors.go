package domain

const (
	// Error messages generated during server state validation
	invalidRoleErrFmt       = "invalid role: expected: %d, actual: %d"
	invalidEntryIndexErrFmt = "invalid entry index: expected: %d, actual: %d"
	invalidTermErrFmt       = "invalid term: expected: %d, actual: %d"
	invalidServerErrFmt     = "invalid %s ID: expected: %d, actual: %d"

	// Fatal errors
	notAvailableErrFmt    = "method %s not available on %s"
	notImplementedErrFmt  = "method %s not implemented"
	termInconsistencyErr  = "remote server term less than local term"
	multipleLeadersErrFmt = "multiple leaders detected during term %d"
	invalidIndexErrFmt    = "log entry index %d is invalid"
	invalidEntryStatus    = "invalid entry status: %d"

	// Unit test errors
	methodExpectedToFailErrFmt    = "method %s expected to fail"
	methodExpectedToSucceedErrFmt = "method %s expected to succeed"
	invalidInputErrFmt            = "input data not consistent for cluster of size %d"
	sizeMismatchErrFmt            = "size mismatch: expected: %d, actual: %d"
	valueMismatchErrFmt           = "value mismatch at index %d: expected: %d, actual: %d"
)
