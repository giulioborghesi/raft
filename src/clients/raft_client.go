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
	AppendEntry(int64, int64, int64) (int64, bool)
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

func (c *raftClient) AppendEntry(serverTerm int64,
	logEntryTerm int64, logEntryIndex int64) (int64, bool) {
	// Create RPC request
	request := &service.AppendEntryRequest{ServerTerm: serverTerm,
		ServerID: c.serverID}

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
