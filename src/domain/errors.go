package domain

const (
	// Error messages generated during server state validation
	invalidRoleErrFmt   = "invalid role: expected: %d, actual: %d"
	invalidTermErrFmt   = "invalid term: expected: %d, actual: %d"
	invalidServerErrFmt = "invalid %s ID: expected: %d, actual: %d"

	// Fatal errors
	forbiddenMethodErrFmt = "method %s cannot be called on %s"
)
