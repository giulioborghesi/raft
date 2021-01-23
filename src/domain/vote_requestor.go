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

// requestVoteResult is a struct used to store the result a requestVote RPC
// call to a remote server
type requestVoteResult struct {
	success    bool
	serverTerm int64
	err        error
}

// abstractVoteRequestor defines the interface of an object to be used for
// requesting a vote from a remote server
type abstractVoteRequestor interface {
	// requestVote sends a vote request to the remote server at the address
	// specified by the string argument
	requestVote(string, int64, int64) requestVoteResult
}

// voteRequestor implements the voteRequestor interface using gRPC
type voteRequestor struct {
	dialOptions []grpc.DialOption
}

func (v *voteRequestor) requestVote(address string, serverTerm int64,
	serverID int64) requestVoteResult {
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
		return requestVoteResult{err: err}
	}
	defer conn.Close()

	// Create gRPC client and send vote request to remote server
	c := service.NewRaftClient(conn)
	result, err := c.RequestVote(ctx, request)
	if err != nil {
		return requestVoteResult{err: err}
	}

	// requestVote RPC was successful, return result to caller
	return requestVoteResult{success: result.Success,
		serverTerm: result.ServerTerm, err: nil}
}
