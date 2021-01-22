package domain

import (
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	candidateErrFmt = "%s should not be called when a server is a candidate"
)

// candidateRole implements the serverRole interface for a candidate server
type candidateRole struct {
	voteRequestorMaker func() voteRequestor
}

func (c *candidateRole) appendEntry(_, _, _, _ int64,
	s *serverState) (int64, bool) {
	panic(fmt.Sprintf(candidateErrFmt, "appendEntry"))
}

func (c *candidateRole) finalizeElection(electionTerm int64,
	results []requestVoteResult, s *serverState) {
	// Election must have started in current term
	maxTerm := s.currentTerm()
	if electionTerm != maxTerm {
		return
	}

	// Start counting valid votes from one because a candidate votes for itself
	count := 1
	for _, result := range results {
		// Skip invalid votes
		if result.err != nil {
			continue
		}

		// Count votes received and determine max term seen
		maxTerm = utils.MaxInt64(maxTerm, result.serverTerm)
		if result.success {
			count += 1
		}
	}

	// A remote server has a term greater than local term, transition to
	// follower
	if maxTerm > s.currentTerm() {
		s.updateServerState(follower, maxTerm, invalidServerID, invalidServerID)
		return
	}

	// Election is won if count of votes received is more than half of servers
	// in the cluster
	if count > ((1 + len(results)) / 2) {
		s.role = leader
		s.leaderID = s.serverID
		s.lastModified = time.Now()
	}
}

func (c *candidateRole) makeCandidate(_ time.Duration, s *serverState) bool {
	// Get new term and update server state
	newTerm := s.currentTerm() + 1
	s.updateServerState(candidate, newTerm, s.serverID, invalidServerID)
	return true
}

func (c *candidateRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()

	// Candidate remains candidate if server term less than current term
	if currentTerm > serverTerm {
		return false
	}

	// Update server state and return
	s.updateServerState(follower, serverTerm, invalidServerID, serverID)
	return true
}

func (c *candidateRole) processAppendEntryEvent(_, _, _ int64,
	_ *serverState) {
}

func (c *candidateRole) requestVote(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Remote server has a term higher than current one, update server state
	s.updateServerState(follower, serverTerm, serverID, invalidServerID)
	return serverTerm, true
}

func (c *candidateRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	candidateTerm := s.currentTerm()
	candidateID := s.serverID

	// Request a vote from each server asynchronously
	results := make([]chan requestVoteResult, 0, len(servers))
	for _, server := range servers {
		// Create vote requestor
		r := c.voteRequestorMaker()
		result := make(chan requestVoteResult)

		// Request vote from server asynchronously
		localServer := server
		go func() {
			result <- r.requestVote(localServer,
				candidateTerm, candidateID)
		}()
		results = append(results, result)
	}

	// Return slice of results for each remote server
	return results
}
