package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	notImplementedErrFmt = "%s not implemented"
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

// unimplementedRaftService provides a basic implementation of a Raft service
// that can be specialized by its user. It is especially useful for unit tests,
// where we only want to change the behavior of a few functions
type unimplementedRaftService struct{}

func (s *unimplementedRaftService) AppendEntry([]*service.LogEntry, int64,
	int64, int64, int64, int64) (int64, bool) {
	panic(fmt.Sprintf(notImplementedErrFmt, "AppendEntry"))
}

func (s *unimplementedRaftService) ApplyCommandAsync(*service.LogEntry) (string,
	int64, error) {
	panic(fmt.Sprintf(notImplementedErrFmt, "ApplyCommandAsync"))
}

func (s *unimplementedRaftService) CommandStatus(string) (logEntryStatus,
	int64, error) {
	panic(fmt.Sprintf(notImplementedErrFmt, "CommandStatus"))
}

func (s *unimplementedRaftService) entryInfo(int64) (int64, int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "entryInfo"))
}

func (s *unimplementedRaftService) entries(int64) ([]*service.LogEntry,
	int64, int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "entries"))
}

func (s *unimplementedRaftService) lastModified() time.Time {
	panic(fmt.Sprintf(notImplementedErrFmt, "lastModified"))
}

func (s *unimplementedRaftService) processAppendEntryEvent(_, _, _ int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "processAppendEntryEvent"))
}

func (s *unimplementedRaftService) RequestVote(int64, int64, int64,
	int64) (int64, bool) {
	panic(fmt.Sprintf(notImplementedErrFmt, "RequestVote"))
}

func (s *unimplementedRaftService) sendHeartbeat(time.Duration) {
	panic(fmt.Sprintf(notImplementedErrFmt, "sendHeartbeat"))
}

func (s *unimplementedRaftService) StartElection(time.Duration) {
	panic(fmt.Sprintf(notImplementedErrFmt, "StartElection"))
}
