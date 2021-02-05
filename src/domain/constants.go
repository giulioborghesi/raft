package domain

const (
	// Empty string
	emptyString = ""

	// Method not implemented
	notImplementedErrFmt = "method %s not implemented"

	// Server roles
	follower  = 0
	leader    = 1
	candidate = 2

	// Server state
	initialTerm          = 0
	invalidLogEntryIndex = -1
	invalidServerID      = -1
	invalidTerm          = -1
)
