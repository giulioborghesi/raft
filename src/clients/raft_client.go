package clients

import (
	"context"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
	"google.golang.org/grpc"
)

const (
	maxCount = 15
)

// AbstractRaftClient is a thin wrapper around around service.RaftClient. It
// provides syntactic sugar to marshall / unmarshall RPC requests / responses
type AbstractRaftClient interface {
	// AppendEntry sends an AppendEntry RPC call to a remote server
	AppendEntry([]*service.LogEntry, int64, int64, int64, int64) (int64, bool)

	// RequestVote sends a RequestVote RPC call to a remote server. The request
	// occurs within a context that must be supplied by the caller
	RequestVote(context.Context, int64, int64, int64, int64) (int64, bool, error)
}

// MakeRaftClient creates a new Raft client given a RPC connection and the
// local server ID
func MakeRaftClient(conn grpc.ClientConnInterface,
	serverID int64) *raftClient {
	c := new(raftClient)
	c.client = service.NewRaftClient(conn)
	c.serverID = serverID
	return c
}

// raftClient implements the AbstractRaftClient interface
type raftClient struct {
	client   service.RaftClient
	serverID int64
}

func (c *raftClient) AppendEntry(entries []*service.LogEntry,
	serverTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, bool) {
	// Create RPC request
	request := &service.AppendEntryRequest{Entries: entries,
		ServerTerm: serverTerm, ServerID: c.serverID,
		PrevEntryTerm: prevEntryTerm, PrevEntryIndex: prevEntryIndex,
		CommitIndex: commitIndex}

	// On error, request is retried using exponential back-off algorithm
	count := 0
	for {
		reply, err := c.client.AppendEntry(context.Background(), request)
		if err == nil {
			return reply.CurrentTerm, reply.Success
		}

		time.Sleep(time.Millisecond * (1 << count))
		count = utils.MaxInt(count+1, maxCount)
	}
}

func (c *raftClient) RequestVote(ctx context.Context, serverTerm int64,
	serverID int64, lastEntryTerm int64, lastEntryIndex int64) (int64,
	bool, error) {
	// Create RPC request and send request to remote server
	request := &service.RequestVoteRequest{ServerTerm: serverTerm,
		ServerID: serverID, LastEntryTerm: lastEntryTerm,
		LastEntryIndex: lastEntryIndex}
	reply, err := c.client.RequestVote(ctx, request)

	// Prepare results and return them to caller
	if err != nil {
		return 0, false, err
	}
	return reply.ServerTerm, reply.Success, nil
}
