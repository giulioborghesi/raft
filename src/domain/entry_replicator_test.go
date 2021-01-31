package domain

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func validateNextLogEntryIndexResult(expectedTerm, actualTerm int64,
	expectedIndex, actualIndex int64, expectedResult, actualResult bool,
	t *testing.T) {
	if expectedTerm != actualTerm {
		t.Fatalf("invalid term: expected: %d, actual: %d", expectedTerm,
			actualTerm)
	}

	if expectedIndex != actualIndex {
		t.Fatalf("invalid log index: expected: %d, actual: %d", expectedIndex,
			actualIndex)
	}

	if expectedResult != actualResult {
		t.Fatalf("invalid return status: expected: %v, actual: %v",
			expectedResult, actualResult)
	}
}

func TestEntryReplicatorResetState(t *testing.T) {
	// Create an entry replicator instance
	er := makeEntryReplicator(invalidServerID, nil, nil)

	// Reset entry replicator state
	var newAppendTerm int64 = int64(rand.Int())
	var newNextIndex int64 = int64(rand.Int())
	er.resetState(newAppendTerm, newNextIndex)

	if er.lastAppendTerm != newAppendTerm {
		t.Fatalf("invalid append term: expected: %d, actual: %d",
			newAppendTerm, er.lastAppendTerm)
	}

	if er.lastLeaderTerm != newAppendTerm {
		t.Fatalf("invalid leader term: expected: %d, actual: %d",
			newAppendTerm, er.lastAppendTerm)
	}

	if er.nextIndex != newNextIndex {
		t.Fatalf("invalid log next index: expected: %d, actual: %d",
			newNextIndex, er.nextIndex)
	}

	if er.matchIndex != invalidLogID {
		t.Fatalf("invalid log next index: expected: %d, actual: %d",
			invalidLogID, er.matchIndex)
	}
}

func TestEntryReplicatorStartStop(t *testing.T) {
	// Create an entry replicator instance
	er := makeEntryReplicator(invalidServerID, nil, nil)

	er.start()
	if err := er.stop(); err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	if err := er.stop(); err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}
}

func TestEntryReplicatorNextLogEntryIndex(t *testing.T) {
	// Create an entry replicator instance
	er := makeEntryReplicator(invalidServerID, nil, nil)

	// Initialize results variables
	var appendTerm int64 = math.MaxInt64
	var logNextIndex int64 = math.MaxInt64
	var active bool = true

	// Invoke goroutine asynchronously
	done := make(chan bool)
	go func() {
		appendTerm, logNextIndex, active = er.nextLogEntryIndex()
		done <- true
	}()

	// Initially, goroutine should be waiting
	select {
	case <-done:
		t.Fatalf("nextLogEntryIndex should have not completed")
	case <-time.After(100 * time.Millisecond):
	}

	// Once conditional variable is signalled, goroutine should complete
	er.Signal()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("nextLogEntryIndex should have not completed")
	}

	// Validate actual log entry state against expected state
	validateNextLogEntryIndexResult(invalidTermID, appendTerm,
		invalidLogID, logNextIndex, false, active, t)

	// Update append term
	er.active = true
	er.appendTerm = er.appendTerm + 1

	// AppendTerm is greater than last leader term, call will succeed
	appendTerm, logNextIndex, active = er.nextLogEntryIndex()
	validateNextLogEntryIndexResult(er.appendTerm, appendTerm,
		0, logNextIndex, true, active, t)
}
