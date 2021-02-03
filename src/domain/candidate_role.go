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
	results []requestVoteResult, s *serverState) bool {
	// Election must have started in current term
	maxTerm := s.currentTerm()
	if electionTerm != maxTerm {
		return false
	}

	// Start counting valid votes from one because a candidate votes for itself
	count := 1
	for _, result := range results {
		// Skip invalid votes
		if result.err != nil {
			continue
		}

		// Server term must not be smaller than local term
		if result.serverTerm < electionTerm {
			panic(invalidRemoteTermErrFmt)
		}

		// Count votes received and determine max term seen
		maxTerm = utils.MaxInt64(maxTerm, result.serverTerm)
		if result.success {
			count += 1
		}
	}

	// A remote server has a term greater than local term, transition to
	// follower
	if maxTerm > electionTerm {
		s.updateServerState(follower, maxTerm, invalidServerID, invalidServerID)
		return false
	}

	// Election is won if count of votes received is more than half of servers
	// in the cluster
	if count >= (len(results)/2 + 1) {
		s.updateServerState(leader, maxTerm, s.serverID, s.serverID)
		return true
	}

	// Server was not elected, return false
	return false
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

	// Voted for should be reset only if remote term greater than local term
	votedFor := s.serverID
	if serverTerm > currentTerm {
		votedFor = invalidServerID
	}

	// Update server state and return
	s.updateServerState(follower, serverTerm, votedFor, serverID)
	return true
}

func (c *candidateRole) processAppendEntryEvent(_, _, _ int64,
	_ *serverState) bool {
	return false
}

func (c *candidateRole) requestVote(serverTerm int64, serverID int64,
	lastEntryTerm int64, lastEntryIndex int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server will grant its vote only if remote log is current
	current := isRemoteLogCurrent(s.log, lastEntryTerm, lastEntryIndex)
	var votedFor int64 = invalidServerID
	if current {
		votedFor = serverID
	}
	s.updateServerState(follower, serverTerm, votedFor, invalidServerID)
	return serverTerm, current
}

func (c *candidateRole) sendHeartbeat(time.Duration, int64, *serverState) {
	return
}

func (c *candidateRole) startElection(
	s *serverState) []chan requestVoteResult {
	// Prepare request arguments
	candidateTerm := s.currentTerm()
	candidateID := s.serverID
	lastEntryIndex := s.log.nextIndex() - 1
	lastEntryTerm := s.log.entryTerm(lastEntryIndex)

	results := make([]chan requestVoteResult, 0)
	for i := 0; i < len(c.voteRequestors); i++ {
		// Fetch vote requestor and prepare channel to save result
		result := make(chan requestVoteResult)
		voteRequestor := c.voteRequestors[i]

		// Send vote request asynchronously
		go func() {
			result <- voteRequestor.requestVote(candidateTerm, candidateID,
				lastEntryTerm, lastEntryIndex)
		}()

		// Append result to output list
		results = append(results, result)
	}

	// Return slice of results for each remote server
	return results
}
