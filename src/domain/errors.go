package domain

const (
	// Error messages generated during server state validation
	invalidRoleErrFmt   = "invalid role: expected: %d, actual: %d"
	invalidTermErrFmt   = "invalid term: expected: %d, actual: %d"
	invalidServerErrFmt = "invalid %s ID: expected: %d, actual: %d"

	// Fatal errors
	notAvailableErrFmt    = "method %s not available on %s"
	notImplementedErrFmt  = "method %s not implemented"
	termInconsistencyErr  = "remote server term less than local term"
	multipleLeadersErrFmt = "multiple leaders detected during term %d"

	// Unit test errors
	methodExpectedToFailErrFmt    = "method %s expected to fail"
	methodExpectedToSucceedErrFmt = "method %s expected to succeed"
)
