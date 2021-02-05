package domain

import (
	"context"
	"fmt"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

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

// roleName returns a human-readable string describing the server role
func roleName(role int) string {
	switch role {
	case candidate:
		return "candidate"
	case follower:
		return "follower"
	case leader:
		return "leader"
	default:
		panic("unknown role")
	}
}

// mockRaftClient implements a RaftClient object that communicates with other
// servers through direct calls to their RaftService objects instead of using
// gRPC calls
type mockRaftClient struct {
	s           *raftService
	serverID    int64
	requestVote bool
}

func (c *mockRaftClient) AppendEntry(entries []*service.LogEntry,
	serverTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, bool) {
	return c.s.AppendEntry(entries, serverTerm, c.serverID, prevEntryTerm,
		prevEntryIndex, commitIndex)
}

func (c *mockRaftClient) RequestVote(ctx context.Context, serverTerm int64,
	serverID int64, lastEntryTerm int64, lastEntryIndex int64) (int64,
	bool, error) {
	if !c.requestVote {
		return invalidTerm, false, fmt.Errorf("cannot vote")
	}
	currentTerm, success := c.s.RequestVote(serverTerm, serverID,
		lastEntryTerm, lastEntryIndex)
	return currentTerm, success, nil
}
