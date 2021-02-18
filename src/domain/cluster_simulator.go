package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/datasources"
)

// makeRaftClusterSimulator creates a Raft cluster simulator to be used for
// unit testing purposes
func makeRaftClusterSimulator(currentTerm []int64, votedFor []int64,
	entryTerms [][]int64) raftClusterSimulator {
	// Determine the cluster size and verify that the input is well formed
	clusterSize := len(currentTerm)
	if len(votedFor) != clusterSize || len(entryTerms) != clusterSize {
		panic(fmt.Sprintf(invalidInputErrFmt, clusterSize))
	}

	// Create a Raft service for each server in the cluster
	services := make([]AbstractRaftService, 0, clusterSize)
	for i := 0; i < clusterSize; i++ {
		// Create the server state
		dao := datasources.MakeInMemoryServerStateDao(currentTerm[i],
			votedFor[i])
		log := &raftLog{AbstractLogDao: datasources.MakeInMemoryLogDao(
			entryTermsToLogEntry(entryTerms[i]))}
		s := MakeServerState(dao, log, int64(i))

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
				active: true}

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
		service.(*raftService).roles[candidate] =
			&candidateRole{voteRequestors: vr}
		service.(*raftService).roles[leader] = &leaderRole{replicators: er,
			matchIndices: matchIndices}
	}

	return raftClusterSimulator{services: services, clients: css}
}

// raftClusterSimulator implements a Raft cluster simulator. The simulator is
// useful for testing the behavior of a Raft cluster on a local machine,
// removing the need for a cloud deployment
type raftClusterSimulator struct {
	clients  [][]clients.AbstractRaftClient
	services []AbstractRaftService
}

// applyCommandAsync applies a command to the cluster asynchronously
func (s *raftClusterSimulator) applyCommandAsync(leaderID int64) (string,
	int64, error) {
	return s.services[leaderID].ApplyCommandAsync(emptyString)
}

// castedVotes returns the votes casted by the servers
func (s *raftClusterSimulator) castedVotes() []int64 {
	votes := make([]int64, 0)
	for _, service := range s.services {
		_, votedFor := service.(*raftService).state.votedFor()
		votes = append(votes, votedFor)
	}
	return votes
}

// logs return the log entries in the log of each server. Only the log entry
// terms are returned: since raftClusterSimulator is expected to be used for
// unit testing, the log entry payload is always assumed to be empty
func (s *raftClusterSimulator) logs() [][]int64 {
	logs := make([][]int64, 0)
	for _, service := range s.services {
		entries, _, _ := service.(*raftService).entries(0)
		log := make([]int64, 0)
		for _, entry := range entries {
			log = append(log, entry.EntryTerm)
		}
		logs = append(logs, log)
	}
	return logs
}

// restoreConnections restores the connections between a server and its peers.
// Connection states are changed only in one direction
func (s *raftClusterSimulator) restoreConnections(serverID int64) {
	for _, client := range s.clients[serverID] {
		client.(*mockRaftClient).active = true
	}
}

// roles returns the servers' roles, encoded as int64 values
func (s *raftClusterSimulator) roles() []int64 {
	roles := []int64{}
	for _, service := range s.services {
		roles = append(roles, int64(service.(*raftService).state.role))
	}
	return roles
}

// sendHeartbeat sends an heartbeat to the followers
func (s *raftClusterSimulator) sendHearthbeat(serverID int64) {
	s.services[serverID].sendHeartbeat(0)
}

// setConnections allows to activate / sever the connections between a server
// and its peers. Connection states are changed only in one direction
func (s *raftClusterSimulator) setConnections(serverID int64,
	clientsStatus []bool) {
	if len(s.clients[serverID]) != len(clientsStatus) {
		panic(fmt.Sprintf(invalidInputErrFmt, len(s.clients)))
	}

	for i, status := range clientsStatus {
		s.clients[serverID][i].(*mockRaftClient).active = status
	}
}

func (s *raftClusterSimulator) setServerRole(serverID int64, role int) {
	s.services[serverID].(*raftService).state.role = role
}

// startElection starts an election on the specified server
func (s *raftClusterSimulator) startElection(serverID int64, d time.Duration) {
	s.services[serverID].StartElection(d)
}

// tearDown tears down the Raft cluster. Using an instance of
// raftClusterSimulator after tearDown has been called will lead to undefined
// behavior and potentially a panic
func (s *raftClusterSimulator) tearDown() {
	// Ensure clients are active
	for i := 0; i < len(s.services); i++ {
		s.restoreConnections(int64(i))
	}
	time.Sleep(10 * time.Millisecond)

	// Disable entry replicators
	for _, service := range s.services {
		role := service.(*raftService).roles[leader].(*leaderRole)
		for _, er := range role.replicators {
			err := er.(*entryReplicator).stop()
			if err != nil {
				panic(err)
			}
		}
	}

	// Reset simulator state
	s.clients = nil
	s.services = nil
}

// terms returns the server's terms
func (s *raftClusterSimulator) terms() []int64 {
	terms := []int64{}
	for _, service := range s.services {
		terms = append(terms, service.(*raftService).state.currentTerm())
	}
	return terms
}
