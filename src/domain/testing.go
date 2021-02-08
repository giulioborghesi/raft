package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

// makeTestServerState creates an instance of serverState to be used for
// testing purposes
func makeTestServerState(currentTerm int64, votedFor int64, serverID int64,
	leaderID int64, role int, active bool) *serverState {
	// Create dao and log
	dao := datasources.MakeInMemoryServerStateDao(currentTerm, votedFor)
	log := &mockRaftLog{value: active}

	s := makeServerState(dao, log, serverID)
	s.leaderID = leaderID
	s.role = role
	return s
}

// validateServerState validates the state of the server state against its
// expected state
func validateServerState(s *serverState, expectedRole int, expectedTerm int64,
	expectedVotedFor int64, expectedLeaderID int64, t *testing.T) {
	if s.role != expectedRole {
		t.Fatalf(invalidRoleErrFmt, expectedRole, s.role)
	}

	term, votedFor := s.votedFor()
	if term != expectedTerm {
		t.Fatalf(invalidTermErrFmt, expectedTerm, term)
	}

	if votedFor != expectedVotedFor {
		t.Fatalf(invalidServerErrFmt, "voted for", expectedVotedFor, votedFor)
	}

	if s.leaderID != expectedLeaderID {
		t.Fatalf(invalidServerErrFmt, "leader", expectedLeaderID, s.leaderID)
	}
}

func validateResults(expected []int64, actual []int64, t *testing.T) {
	if len(expected) != len(actual) {
		t.Fatalf(sizeMismatchErrFmt, len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Fatalf(valueMismatchErrFmt, i, expected[i], actual[i])
		}
	}
}

// mockRaftClient implements a RaftClient object that communicates with other
// servers through direct calls to their RaftService objects instead of using
// gRPC calls
type mockRaftClient struct {
	s        AbstractRaftService
	serverID int64
	active   bool
}

func (c *mockRaftClient) AppendEntry(entries []*service.LogEntry,
	serverTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, bool) {
	for !c.active {
		time.Sleep(time.Millisecond)
	}
	return c.s.AppendEntry(entries, serverTerm, c.serverID, prevEntryTerm,
		prevEntryIndex, commitIndex)
}

func (c *mockRaftClient) RequestVote(ctx context.Context, serverTerm int64,
	serverID int64, lastEntryTerm int64, lastEntryIndex int64) (int64,
	bool, error) {
	if !c.active {
		return invalidTerm, false, fmt.Errorf("cannot vote")
	}
	currentTerm, success := c.s.RequestVote(serverTerm, serverID,
		lastEntryTerm, lastEntryIndex)
	return currentTerm, success, nil
}
