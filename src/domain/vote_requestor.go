package domain

import (
	"context"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"google.golang.org/grpc"
)

const (
	maxCallDuration = time.Millisecond * 100
)

// voteRequestor specifies the interface of the object used to request a vote
// from a remote server
type voteRequestor interface {
	// requestVote sends a vote request to the remote server at the address
	// specified by the string argument
	requestVote(string, int64, int64) bool
}

// rpcVoteRequestor implements the voteRequestor interface using gRPC
type rpcVoteRequestor struct {
	dialOptions []grpc.DialOption
}

func (v *rpcVoteRequestor) requestVote(address string, serverTerm int64,
	serverID int64) (bool, int64, error) {
	// Prepare gRPC call arguments
	request := new(service.RequestVoteRequest)
	request.ServerTerm = serverTerm
	request.ServerID = serverID

	// Create a connection context with a deadline
	d := time.Now().Add(maxCallDuration)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	// Connect to remote server
	conn, err := grpc.DialContext(ctx, address, v.dialOptions...)
	if err != nil {
		return false, -1, err
	}
	defer conn.Close()

	// Create gRPC client and send vote request to remote server
	c := service.NewRaftClient(conn)
	result, err := c.RequestVote(ctx, request)
	if err != nil {
		return false, -1, err
	}
	return result.Success, result.ServerTerm, nil
}
