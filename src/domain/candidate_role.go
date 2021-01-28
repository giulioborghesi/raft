package domain

import (
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// candidateRole implements the serverRole interface for a candidate server
type candidateRole struct {
	voteRequestors []abstractVoteRequestor
}

func (c *candidateRole) appendEntry(_ []*service.LogEntry, _, _, _, _, _ int64,
	s *serverState) (int64, bool) {
	panic(fmt.Sprintf(roleErrCallFmt, "appendEntry", "candidate"))
}

func (c *candidateRole) appendNewEntry(_ *service.LogEntry, _ int64,
	s *serverState) (string, int64, error) {
	return "", s.leaderID, fmt.Errorf(wrongRoleErrFmt, "candidate")
}

func (c *candidateRole) entryStatus(_ string, _ int64,
	s *serverState) (logEntryStatus, int64, error) {
	return invalid, s.leaderID, fmt.Errorf(wrongRoleErrFmt, "candidate")
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
		s.updateServerState(leader, maxTerm, s.serverID, s.serverID)
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
	_ *serverState) bool {
	return false
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

func (c *candidateRole) sendHeartbeat(time.Duration, int64, *serverState) {
	return
}

func (c *candidateRole) startElection(
	s *serverState) []chan requestVoteResult {
	candidateTerm := s.currentTerm()
	candidateID := s.serverID

	results := make([]chan requestVoteResult, 0)
	for i := 0; i < len(c.voteRequestors); i++ {
		// Fetch vote requestor and prepare channel to save result
		result := make(chan requestVoteResult)
		voteRequestor := c.voteRequestors[i]

		// Send vote request asynchronously
		go func() {
			result <- voteRequestor.requestVote(candidateTerm, candidateID)
		}()

		// Append result to output list
		results = append(results, result)
	}

	// Return slice of results for each remote server
	return results
}
