package domain

import (
	"context"
	"fmt"
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
	serverTerm int64, serverID int64, _, _ int64) (int64, bool, error) {
	panic(fmt.Sprintf(notImplementedErrFmt, "RequestVote"))
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
	s := makeEntrySender(invalidTerm, c)

	// Send entries to server. First fourth calls are expected to fail
	appendTerm := prevEntryTerm + 2
	prevEntryIndex := matchPrevEntryIndex + 4
	for i := 0; i < 4; i++ {
		serverTerm, success := s.sendEntries(nil, appendTerm,
			prevEntryTerm, prevEntryIndex, invalidLogEntryIndex)

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
		prevEntryTerm, prevEntryIndex, invalidLogEntryIndex)
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
			invalidLogEntryIndex)
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
