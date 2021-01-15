package domain

import (
	"fmt"
	"time"
)

const (
	leaderErrMultFmt = "multiple leaders in the same term detected"
)

type leaderRole struct{}

func (l *leaderRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
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
	s.role = follower
	s.updateTerm(serverTerm)
	s.leaderID = serverID
	s.lastModified = time.Now()
	return true
}

func (l *leaderRole) requestVote(serverTerm int64, serverID int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server term greater than local term, cast vote and become follower
	s.role = follower
	s.updateTermVotedFor(serverTerm, serverID)
	s.leaderID = invalidLeaderID
	s.lastModified = time.Now()
	return serverTerm, true
}

func (l *leaderRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "leader"))
}
