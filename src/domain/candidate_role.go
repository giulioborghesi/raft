package domain

import (
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

type candidateRole struct {
	voteRequestorMaker func() voteRequestor
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
		maxTerm = utils.MaxI64(maxTerm, result.serverTerm)
		if result.success {
			count += 1
		}
	}

	// A remote server has a term greater than local term, transition to
	// follower
	if maxTerm > s.currentTerm() {
		s.role = follower
		s.updateTerm(maxTerm)
		s.lastModified = time.Now()
		return
	}

	// Election is won if count of votes received is more than half of servers
	// in the cluster
	if count > ((1 + len(results)) / 2) {
		s.role = leader
		s.lastModified = time.Now()
	}
}
