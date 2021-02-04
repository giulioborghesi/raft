package domain

import (
	"context"
	"fmt"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

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
		return invalidTermID, false, fmt.Errorf("cannot vote")
	}
	currentTerm, success := c.s.RequestVote(serverTerm, serverID,
		lastEntryTerm, lastEntryIndex)
	return currentTerm, success, nil
}
