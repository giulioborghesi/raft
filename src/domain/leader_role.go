package domain

import (
	"container/list"
	"fmt"
	"time"
)

const (
	leaderErrMultFmt = "multiple leaders in the same term detected"
)

// leaderRole implements the serverRole interface for a leader server
type leaderRole struct {
	appenders     []abstractEntryReplicator
	appendResults *list.List
	matchIndices  []int64
}

func (l *leaderRole) appendEntry(_ []*logEntry, _, _, _, _, _ int64,
	s *serverState) (int64, bool) {
	panic(fmt.Sprintf(roleErrCallFmt, "appendEntry", "leader"))
}

func (l *leaderRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) {
	panic(fmt.Sprintf(roleErrCallFmt, "finalizeElection", "leader"))
}

func (l *leaderRole) makeCandidate(_ time.Duration, s *serverState) bool {
	// Leaders cannot transition to candidates
	return false
}

func (l *leaderRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()

	if currentTerm == serverTerm {
		// Multiple leaders in the same term are not allowed
		panic(fmt.Sprintf(leaderErrMultFmt))
	} else if currentTerm > serverTerm {
		// Request received from a former leader, reject it
		return false
	}

	// Request received from new leader, transition to follower
	s.updateServerState(follower, serverTerm, invalidServerID, serverID)
	return true
}

func (l *leaderRole) processAppendEntryEvent(appendTerm int64,
	matchIndex int64, serverID int64, s *serverState) bool {
	// Append entry succeeded and server term did not change
	if appendTerm == s.currentTerm() && matchIndex != invalidLogID {
		l.matchIndices[serverID] = matchIndex
		s.updateCommitIndex(l.matchIndices)
		return true
	}

	// Switch role to follower if a new term was detected
	if appendTerm > s.currentTerm() {
		s.updateServerState(follower, appendTerm, invalidServerID,
			invalidServerID)
	}
	return false
}

func (l *leaderRole) requestVote(serverTerm int64, serverID int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server term greater than local term, cast vote and become follower
	s.updateServerState(follower, serverTerm, serverID, invalidServerID)
	return serverTerm, true
}

func (l *leaderRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "leader"))
}
