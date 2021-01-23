package domain

/*
import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const (
	initialServerTerm  = 5
	bufSize            = 1024 * 1024
	defaultCandidateID = 0
	timeoutCandidateID = 1
)

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	service.RegisterRaftServer(s, &mockRaftServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
}

func bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return lis.Dial()
}

// mockRaftServer implements a mock RaftServer that always returns true. To
// test timeout, the implementation sleeps for a duration that exceeds the
// timeout when the calling server ID is equal to timeoutCandidateID
type mockRaftServer struct {
	service.UnimplementedRaftServer
}

func (s *mockRaftServer) RequestVote(ctx context.Context,
	request *service.RequestVoteRequest) (*service.RequestVoteReply, error) {
	if request.ServerID == timeoutCandidateID {
		time.Sleep(time.Millisecond * 200)
	}
	return &service.RequestVoteReply{Success: true,
		ServerTerm: request.ServerTerm}, nil
}

func TestVoteRequestRPCSuccess(t *testing.T) {
	// Create RPC client
	options := []grpc.DialOption{grpc.WithContextDialer(bufDialer),
		grpc.WithBlock(), grpc.WithInsecure()}
	r := voteRequestor{dialOptions: options}

	// Prepare calling arguments and send vote request
	var currentTerm int64 = initialServerTerm
	result := r.requestVote("bufnet", currentTerm, defaultCandidateID)

	// We expect the request vote RPC to be rejected as the remote server term
	// is larger than the candidate server term
	if result.err != nil {
		t.Errorf("requestVote RPC failed with error: %v", result.err)
	}

	if !result.success {
		t.Errorf("requestVote RPC expected to succeed, failed instead")
	}

	if result.serverTerm != currentTerm {
		t.Errorf("requestVote RPC returned wrong remote server term: "+
			"expected %d, actual %d", currentTerm, result.serverTerm)
	}
}

func TestVoteRequestRPCTimeout(t *testing.T) {
	// Create RPC client
	options := []grpc.DialOption{grpc.WithContextDialer(bufDialer),
		grpc.WithBlock(), grpc.WithInsecure()}
	r := voteRequestor{dialOptions: options}

	// Prepare calling arguments and send vote request
	var currentTerm int64 = initialServerTerm
	result := r.requestVote("bufnet", currentTerm,
		timeoutCandidateID)

	// We expect the request vote RPC to experience a timeout failure
	if result.err == nil {
		t.Errorf("requestVote RPC expected to fail due to timeout")
	}

	if errcode := status.Code(result.err); errcode != codes.DeadlineExceeded {
		t.Errorf("requestVote RPC failed with invalid error code: "+
			"expected: %d, actual %d", codes.DeadlineExceeded, int(errcode))
	}
}
*/
