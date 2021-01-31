package domain

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

// mockRaftServiceEntryReplicator implements a mock Raft service to be used
// for unit testing entry replicator
type mockRaftServiceEntryReplicator struct {
	unimplementedRaftService
	leaderTerm            int64
	entryTerm             int64
	countEntryInfo        int64
	countAppendEntryEvent int64
}

func (s *mockRaftServiceEntryReplicator) entryInfo(int64) (int64, int64) {
	s.countEntryInfo++
	return s.leaderTerm, s.entryTerm
}

func (s *mockRaftServiceEntryReplicator) processAppendEntryEvent(_, _,
	_ int64) {
	s.countAppendEntryEvent++
}

// validateNextLogEntryIndexResult validates the results of a call to
// nextLogEntryIndex by providing the expected results
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

	if er.leaderTerm != newAppendTerm {
		t.Fatalf("invalid append term: expected: %d, actual: %d",
			newAppendTerm, er.leaderTerm)
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

func TestEntryReplicatorUpdateMatchIndex(t *testing.T) {
	// Create an entry replicator instance
	service := &mockRaftServiceEntryReplicator{leaderTerm: 5, entryTerm: 4}
	er := makeEntryReplicator(invalidServerID, nil, service)

	// appendTerm is equal to leader term, return true
	success := er.updateMatchIndex(er.leaderTerm, invalidLogID)
	if !success {
		t.Fatalf("updateMatchIndex expected to succeed")
	}

	// appendTerm is less than entry term, return false
	logNextIndex := int64(5)
	success = er.updateMatchIndex(service.leaderTerm-1, logNextIndex)
	if success {
		t.Fatalf("updateMatchIndex expected to fail")
	}

	// Check number of API calls
	if service.countEntryInfo != 1 {
		t.Fatalf("expected %d calls to entryInfo, got %d",
			1, service.countEntryInfo)
	}

	if service.countAppendEntryEvent != 1 {
		t.Fatalf("expected %d calls to AppendEntryEvent, got %d",
			1, service.countAppendEntryEvent)
	}

	// Reset counters
	service.countEntryInfo = 0
	service.countAppendEntryEvent = 0
}
