package domain

import (
	"context"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
)

const (
	maxCallDuration = time.Millisecond * 100
)

// requestVoteResult is a struct used to return the result a requestVote RPC
// call to a remote server
type requestVoteResult struct {
	success    bool
	serverTerm int64
	err        error
}

// abstractVoteRequestor defines the interface of the object used for
// requesting a vote from a remote server. The recipient of a vote
// request should be binded to the requestor during object initialization
type abstractVoteRequestor interface {
	// requestVote sends a vote request to a remote server
	requestVote(int64, int64) requestVoteResult
}

// mockVoteRequestor implements a mock vote requestor that will return false
// for a configurable number of times, after which it will return true
type mockVoteRequestor struct {
	count, maxCount int
}

func (v *mockVoteRequestor) requestVote(serverTerm,
	_ int64) requestVoteResult {
	if v.count < v.maxCount {
		v.count++
		return requestVoteResult{serverTerm: serverTerm, success: false,
			err: nil}
	}
	return requestVoteResult{serverTerm: serverTerm, success: true, err: nil}
}

// voteRequestor implements the abstractVoteRequestor interface. voteRequestor
// delegates most of the implementation details to a client class; its
// responsibility is limited to marshalling / unmarshalling the input / output
type voteRequestor struct {
	client clients.AbstractRaftClient
}

// makeVoteRequestor creates an object of type voteRequestor, intialized with a
// provided client, and returns a pointer to it to the caller
func makeVoteRequestor(c clients.AbstractRaftClient) abstractVoteRequestor {
	return &voteRequestor{client: c}
}

func (v *voteRequestor) requestVote(serverTerm int64,
	serverID int64) requestVoteResult {
	d := time.Now().Add(maxCallDuration)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	serverTerm, success, err := v.client.RequestVote(ctx, serverTerm,
		serverID, invalidTermID, invalidLogID)
	return requestVoteResult{success: success, serverTerm: serverTerm, err: err}
}
