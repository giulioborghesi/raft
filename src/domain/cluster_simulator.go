package domain

import (
	"math"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/datasources"
)

// makeRaftClusterSimulator creates a Raft cluster simulator to be used for
// unit testing purposes
func makeRaftClusterSimulator(currentTerm []int64, votedFor []int64,
	entries [][]int64) raftClusterSimulator {
	// Create a Raft service for each server in the cluster
	clusterSize := len(currentTerm)
	services := make([]AbstractRaftService, 0, clusterSize)
	for i := 0; i < clusterSize; i++ {
		// Create the server state
		dao := datasources.MakeInMemoryServerStateDao(currentTerm[i],
			votedFor[i])
		log := &raftLog{e: nil}
		s := makeServerState(dao, log, int64(i))

		// Create the Raft service
		service := &raftService{state: s, roles: make(map[int]serverRole)}
		service.roles[follower] = &followerRole{}
		services = append(services, service)
	}

	// Initialize the candidate and leader roles for each service
	css := make([][]clients.AbstractRaftClient, 0)
	for i, service := range services {
		vr := make([]abstractVoteRequestor, 0, clusterSize)
		er := make([]abstractEntryReplicator, 0, clusterSize)

		cs := make([]clients.AbstractRaftClient, 0)
		for j := 0; j < clusterSize; j++ {
			if i == j {
				continue
			}

			// Create mock client
			c := &mockRaftClient{serverID: int64(i), s: services[j],
				requestVote: true}

			// Create vote requestor and entry replicator
			cs = append(cs, c)
			vr = append(vr, makeVoteRequestor(c))
			er = append(er, makeEntryReplicator(int64(j), c, service))
		}
		css = append(css, cs)

		// Initialize match indices
		matchIndices := make([]int64, clusterSize)
		matchIndices[i] = math.MaxInt64

		// Initializae candidate and leader roles
		service.roles[candidate] = &candidateRole{voteRequestors: vr}
		service.roles[leader] = &leaderRole{replicators: er,
			matchIndices: matchIndices}
	}

	return raftClusterSimulator{services: services}
}

// raftClusterSimulator implements a simulator for a Raft cluster. The
// simulator is useful for testing the behavior of a Raft cluster on a local
// machine, without having to deploy an entire cluster
type raftClusterSimulator struct {
	services []AbstractRaftService
}

func (s *raftClusterSimulator) setConnections(serverID int64, status []bool) {

}

func (s *raftClusterSimulator) restoreConnections(serverID int64) {

}
