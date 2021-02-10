package domain

import "time"

const (
	// Empty string
	emptyString = ""

	// Server roles
	follower  = 0
	leader    = 1
	candidate = 2

	// Server state
	initialTerm          = 0
	invalidLogEntryIndex = -1
	invalidServerID      = -1
	invalidTerm          = -1

	// Unit test constants
	testStartingTerm = 15
	testSleepTime    = 10 * time.Millisecond
)
