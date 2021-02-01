package domain

import (
	"context"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// mockRaftClientEntryReplicator is an entry replicator used to unit test
// entryReplicator. mockRaftClientEntryReplicator returns a match only when
// the previous entry term and the previous entry index match the values
// used to initialize the client
type mockRaftClientEntryReplicator struct {
	prevEntryTerm  int64
	prevEntryIndex int64
}

func (c *mockRaftClientEntryReplicator) AppendEntry(
	entries []*service.LogEntry, serverTerm int64, prevEntryTerm int64,
	prevEntryIndex int64, commitIndex int64) (int64, bool) {
	if prevEntryTerm == c.prevEntryTerm && prevEntryIndex == c.prevEntryIndex {
		return serverTerm, true
	}
	return serverTerm, false
}

func (c *mockRaftClientEntryReplicator) RequestVote(ctx context.Context,
	serverTerm int64, serverID int64) (int64, bool, error) {
	panic("method not implemented")
}

// mockRaftServiceEntryAppender implements a Raft service to be used for unit
// testing entryAppender
type mockRaftServiceEntryAppender struct {
	unimplementedRaftService
	appendTerm     int64
	prevEntryTerms []int64
}

func (s *mockRaftServiceEntryAppender) entryInfo(prevEntryIndex int64) (int64,
	int64) {
	return s.appendTerm, s.prevEntryTerms[prevEntryIndex]
}

func TestEntrySenderSendEntries(t *testing.T) {
	// Test data initialization
	prevEntryTerm := int64(4)
	matchPrevEntryIndex := int64(3)

	// Create entry sender
	c := &mockRaftClientEntryReplicator{prevEntryTerm: prevEntryTerm,
		prevEntryIndex: matchPrevEntryIndex}
	s := makeEntrySender(invalidTermID, c)

	// Send entries to server. First fourth calls are expected to fail
	appendTerm := prevEntryTerm + 2
	prevEntryIndex := matchPrevEntryIndex + 4
	for i := 0; i < 4; i++ {
		serverTerm, success := s.sendEntries(nil, appendTerm,
			prevEntryTerm, prevEntryIndex, invalidLogID)

		if success {
			t.Fatalf("sendEntries is expected to fail")
		}

		if serverTerm != appendTerm {
			t.Fatalf("invalid server term: expected: %d, actual: %d",
				appendTerm, serverTerm)
		}

		prevEntryIndex = s.nextEntryIndex() - 1
	}

	// Fifth call will succeed
	serverTerm, success := s.sendEntries(nil, appendTerm,
		prevEntryTerm, prevEntryIndex, invalidLogID)
	if !success {
		t.Fatalf("sendEntries is expected to succeed")
	}

	if serverTerm != appendTerm {
		t.Fatalf("invalid server term: expected: %d, actual: %d",
			appendTerm, serverTerm)
	}

	// Next entry index should have not changed because no log entry was sent
	nextEntryIndex := s.nextEntryIndex()
	if nextEntryIndex != prevEntryIndex+1 {
		t.Fatalf("invalid next entry index: expected: %d, actual: %d",
			prevEntryIndex+1, nextEntryIndex)
	}

	// Trigger a panic by calling send entries with an invalid entry index
	sendEntries := func() {
		s.sendEntries(nil, appendTerm, prevEntryTerm, prevEntryIndex-1,
			invalidLogID)
	}
	utils.AssertPanic(t, "sendEntries", sendEntries)
}

func TestEntryReplicatorUpdateMatchIndex(t *testing.T) {
	// Test data
	appendTerm := int64(5)
	prevEntryTerm := int64(4)
	matchPrevEntryIndex := int64(3)
	prevEntryTerms := []int64{0, 0, 3, 4, 4, 5, 5}

	// Create entry replicator
	c := &mockRaftClientEntryReplicator{prevEntryTerm: prevEntryTerm,
		prevEntryIndex: matchPrevEntryIndex}
	s := &mockRaftServiceEntryAppender{appendTerm: appendTerm,
		prevEntryTerms: prevEntryTerms}
	r := newEntryReplicator(invalidServerID, c, s)

	// Update match indices
	nextIndex := int64(7)
	currentTerm, nextEntryIndex := r.updateMatchIndex(appendTerm, nextIndex)

	if currentTerm != appendTerm {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			appendTerm, currentTerm)
	}

	if nextEntryIndex != matchPrevEntryIndex+1 {
		t.Fatalf("invalid next entry index: expected: %d, actual: %d",
			matchPrevEntryIndex+1, nextEntryIndex)
	}

	// A second call to matchIndices should be a no-op because the append term
	// has not changed
	currentTerm, nextEntryIndex = r.updateMatchIndex(appendTerm, nextIndex)

	if currentTerm != appendTerm {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			appendTerm, currentTerm)
	}

	if nextEntryIndex != matchPrevEntryIndex+1 {
		t.Fatalf("invalid next entry index: expected: %d, actual: %d",
			matchPrevEntryIndex+1, nextEntryIndex)
	}
}

/*

I now want to implement a unit test

*/

// mockRaftServiceEntryReplicator implements a mock Raft service to be used
// for unit testing entry replicator
/*type mockRaftServiceEntryReplicator struct {
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
		appendTerm, logNextIndex, active = er.nextEntryIndex()
		done <- true
	}()

	// Initially, goroutine should be waiting
	select {
	case <-done:
		t.Fatalf("nextLogEntryIndex should have not completed")
	case <-time.After(100 * time.Millisecond):
	}

	// Update append term
	er.active = true
	er.appendTerm = er.appendTerm + 1

	// AppendTerm is greater than last leader term, call will succeed
	appendTerm, logNextIndex, active = er.nextEntryIndex()
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
}*/
