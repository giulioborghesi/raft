package domain

import (
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	candidateErrFmt = "%s should not be called when a server is a candidate"
)

type candidateRole struct {
	voteRequestorMaker func() voteRequestor
}

func (c *candidateRole) appendEntry(serverTerm int64, serverID int64,
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

func (l *candidateRole) makeCandidate(_ time.Duration, s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()

	// Update term, voted for and last modified
	s.updateTermVotedFor(currentTerm+1, s.serverID)
	s.lastModified = time.Now()
	return true
}

func (c *candidateRole) makeFollower(serverTerm int64, s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()

	// Candidate remains candidate if server term less than current term
	if currentTerm > serverTerm {
		return false
	}

	// Change role to follower, update term and last modified
	s.role = follower
	s.updateTerm(serverTerm)
	s.lastModified = time.Now()
	return true
}

func (c *candidateRole) requestVote(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Remote server has a term higher than current one, cast vote and convert
	// to follower
	s.role = follower
	s.updateTermVotedFor(serverTerm, serverID)
	s.lastModified = time.Now()
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
