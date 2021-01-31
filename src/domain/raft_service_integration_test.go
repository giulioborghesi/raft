package domain

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	testClusterSize                     = 3
	testServiceIntegrationLeaderID      = 1
	testServiceIntegrationOtherServerID = 2
)

// localRaftClient provides an implementation of RaftClient to be used for
// unit tests
type localRaftClient struct {
	s           *raftService
	serverID    int64
	requestVote bool
}

func (c *localRaftClient) AppendEntry(entries []*service.LogEntry,
	serverTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, bool) {
	return c.s.AppendEntry(entries, serverTerm, c.serverID, prevEntryTerm,
		prevEntryIndex, commitIndex)
}

func (c *localRaftClient) RequestVote(ctx context.Context,
	serverTerm int64, serverID int64) (int64, bool, error) {
	if !c.requestVote {
		return invalidTermID, false, fmt.Errorf("cannot vote")
	}
	currentTerm, success := c.s.RequestVote(serverTerm, serverID)
	return currentTerm, success, nil
}

func createMockRaftService(vr []abstractVoteRequestor,
	er []abstractEntryReplicator) *raftService {
	// Create a server state instance
	dao := datasources.MakeTestServerStateDao()
	log := &mockRaftLog{value: true}
	s := makeServerState(dao, log, testServiceIntegrationLeaderID)

	// Create and initialize Raft service
	service := &raftService{state: s, roles: make(map[int]serverRole)}
	service.roles[follower] = &followerRole{}
	service.roles[candidate] = &candidateRole{voteRequestors: vr}
	service.roles[leader] = &leaderRole{replicators: er,
		matchIndices: []int64{math.MaxInt64, 0}}
	return service
}

func createMockRaftCluster() ([]*raftService, [][]*localRaftClient) {
	services := make([]*raftService, 0, testClusterSize)
	for i := 0; i < testClusterSize; i++ {
		// Create a server state instance
		dao := datasources.MakeTestServerStateDao()
		log := &raftLog{}
		log.e = make([]*service.LogEntry, 0)
		s := makeServerState(dao, log, int64(i))

		// Create the raft service
		service := &raftService{state: s, roles: make(map[int]serverRole)}
		service.roles[follower] = &followerRole{}
		services = append(services, service)
	}

	// For each service initialize the candidate and leader roles
	css := make([][]*localRaftClient, 0)
	for i, service := range services {
		vr := make([]abstractVoteRequestor, 0, testClusterSize)
		er := make([]abstractEntryReplicator, 0, testClusterSize)

		cs := make([]*localRaftClient, 0)
		for j := 0; j < testClusterSize; j++ {
			if i != j {
				// Create local client
				c := &localRaftClient{serverID: int64(i), s: services[j],
					requestVote: true}
				cs = append(cs, c)

				// Create vote requestor
				vr = append(vr, makeVoteRequestor(c))

				// Create entry replicator
				er = append(er, MakeEntryReplicator(int64(j), c, service))
			}
		}
		css = append(css, cs)

		// Initialize match indices
		matchIndices := make([]int64, testClusterSize)
		matchIndices[i] = math.MaxInt64

		// Initialize leader and candidate roles
		service.roles[candidate] = &candidateRole{voteRequestors: vr}
		service.roles[leader] = &leaderRole{replicators: er,
			matchIndices: []int64{0, 0, 0}}
	}

	return services, css
}

func TestLeaderAppendEntry(t *testing.T) {
	// Create service
	s := createMockRaftService(nil, nil)
	s.state.updateTerm(s.state.currentTerm() + 1)
	s.state.role = leader

	// Store initial term
	initialTerm := s.state.currentTerm()

	// Server receives append entry request from previous leader
	newTerm, success := s.AppendEntry(nil, initialTerm-1,
		testServiceIntegrationOtherServerID, 0, 0, 0)

	if newTerm != initialTerm {
		t.Fatalf("invalid term returned by AppendEntry: "+
			"expected: %d, actual: %d", initialTerm, newTerm)
	}

	if success {
		t.Fatalf("call to AppendEntry was expected to fail")
	}

	// Server receives append entry request from new leader
	newTerm, success = s.AppendEntry(nil, initialTerm+1,
		testServiceIntegrationOtherServerID, 0, 0, 0)

	if newTerm != initialTerm+1 {
		t.Fatalf("invalid term returned by AppendEntry: "+
			"expected: %d, actual: %d", initialTerm+1, newTerm)
	}

	if !success {
		t.Fatalf("call to AppendEntry was expected to succeed")
	}

}

func TestLeaderElection(t *testing.T) {
	// Create service
	vr := []abstractVoteRequestor{&mockVoteRequestor{maxCount: 1},
		&mockVoteRequestor{maxCount: 2}}
	s := createMockRaftService(vr, nil)

	// Store initial term
	initialTerm := s.state.currentTerm()

	// Force new election. Election is expected to fail
	d := time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != candidate {
		t.Fatalf("server expected to transition to candidate")
	}

	if s.state.leaderID != invalidServerID {
		t.Fatalf("unexpected leader ID: expected: %d, actual: %d",
			invalidServerID, s.state.leaderID)
	}

	if s.state.currentTerm() != (initialTerm + 1) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}

	// Force another election. Election is expected to succeed
	d = time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != leader {
		t.Fatalf("server expected to transition to leader")
	}

	if s.state.leaderID != testServiceIntegrationLeaderID {
		t.Fatalf("unexpected leader ID: expected: %d, actual: %d",
			testServiceIntegrationLeaderID, s.state.leaderID)
	}

	if s.state.currentTerm() != (initialTerm + 2) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}
}

func TestRaftCluster(t *testing.T) {
	// Create Raft cluster
	services, clients := createMockRaftCluster()
	time.Sleep(time.Millisecond)

	// Start a new election
	leaderID := 0
	services[leaderID].StartElection(time.Duration(0))

	time.Sleep(time.Millisecond)
	for serverID := 0; serverID < testClusterSize; serverID++ {
		// Server with ID equal to leaderID should be the new leader
		if serverID == leaderID {
			if services[serverID].state.role != leader {
				t.Fatalf(fmt.Sprintf("server %d expected to become leader",
					serverID))
			}
		} else {
			if services[serverID].state.role != follower {
				t.Fatalf(fmt.Sprintf("server %d expected to become follower",
					serverID))
			}
		}

		// All servers should have voted for leaderID
		_, votedFor := services[serverID].state.votedFor()
		if votedFor != int64(leaderID) {
			t.Fatalf("vote casted for wrong server: "+
				"expected: %d, actual: %d", leaderID, votedFor)
		}

		// Current term should be 1
		currentTerm := services[serverID].state.currentTerm()
		if currentTerm != 1 {
			t.Fatalf("invalid term: expected %d, actual %d", 1, currentTerm)
		}
	}

	// Try starting a new election where timer was reset just in time
	newLeaderID := 1
	services[newLeaderID].StartElection(time.Minute)

	time.Sleep(time.Millisecond)
	for serverID := 0; serverID < testClusterSize; serverID++ {
		// Leader should have not changed
		if serverID == leaderID {
			if services[serverID].state.role != leader {
				t.Fatalf(fmt.Sprintf("server %d expected to become leader",
					serverID))
			}
		} else {
			if services[serverID].state.role != follower {
				t.Fatalf(fmt.Sprintf("server %d expected to become follower",
					serverID))
			}
		}

		// Current term should still be 1
		currentTerm := services[serverID].state.currentTerm()
		if currentTerm != 1 {
			t.Fatalf("invalid term: expected %d, actual %d", 1, currentTerm)
		}
	}

	// This time, election will succeed
	services[newLeaderID].StartElection(time.Duration(0))

	time.Sleep(time.Millisecond)
	for serverID := 0; serverID < testClusterSize; serverID++ {
		// Leader should have changed to newLeaderID
		if serverID == newLeaderID {
			if services[serverID].state.role != leader {
				t.Fatalf("server %d expected to become leader", serverID)
			}
		} else {
			if services[serverID].state.role != follower {
				t.Fatalf("server %d expected to become follower", serverID)
			}
		}

		// All servers should have voted for newLeaderID
		_, votedFor := services[serverID].state.votedFor()
		if votedFor != int64(newLeaderID) {
			t.Fatalf("vote casted for wrong server: "+
				"expected: %d, actual: %d", newLeaderID, votedFor)
		}

		// Current term should now be 2
		currentTerm := services[serverID].state.currentTerm()
		if currentTerm != 2 {
			t.Fatalf("invalid term: expected %d, actual %d", 2, currentTerm)
		}
	}

	// Consider now a situation where one server does not vote
	clients[leaderID][1].requestVote = false
	services[leaderID].StartElection(time.Duration(0))

	time.Sleep(time.Millisecond)
	for serverID := 0; serverID < testClusterSize; serverID++ {
		// Leader should have changed to leaderID
		if serverID == leaderID {
			if services[serverID].state.role != leader {
				t.Fatalf("server %d expected to become leader", serverID)
			}
		} else {
			if services[serverID].state.role != follower {
				t.Fatalf("server %d expected to become follower", serverID)
			}
		}

		// All servers but 1 should have voted for new leader
		_, votedFor := services[serverID].state.votedFor()
		if serverID == 2 {
			if votedFor != invalidServerID {
				t.Fatalf("server %d should have not casted a vote", serverID)
			}
		} else if votedFor != int64(leaderID) {
			t.Fatalf("vote casted for wrong server: "+
				"expected: %d, actual: %d", leaderID, votedFor)
		}

		// Current term should now be 3
		currentTerm := services[serverID].state.currentTerm()
		if currentTerm != 3 {
			t.Fatalf("invalid term: expected %d, actual %d", 3, currentTerm)
		}
	}

	// Teardown cluster
	for _, service := range services {
		role := service.roles[leader].(*leaderRole)
		for _, er := range role.replicators {
			er.(*entryReplicator).stop()
		}
	}
}
