package domain

import (
	"fmt"
	"time"
)

const (
	followerErrIdFmt   = "invalid leader id: expected: %d, actual: %d"
	followerErrTermFmt = "invalid term number: expected: %d, actual: %d"
)

// followerRole implements the serverRole interface for a follower server
type followerRole struct{}

func (f *followerRole) appendEntry(entries []*logEntry, serverTerm, serverID,
	prevLogTerm, prevLogIndex int64, commitIndex int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm != serverTerm {
		panic(fmt.Sprintf(followerErrTermFmt, currentTerm, serverTerm))
	}

	// Server ID
	if serverID != s.leaderID {
		panic(fmt.Sprintf(followerErrIdFmt, s.leaderID, serverID))
	}

	// Try appending log entries to log
	return currentTerm, s.log.appendEntries(entries, prevLogTerm, prevLogIndex)
}

func (f *followerRole) appendNewEntry(_ *logEntry, _ int64,
	s *serverState) (string, int64, error) {
	return "", s.leaderID, fmt.Errorf(wrongRoleErrFmt, "follower")
}

func (f *followerRole) entryStatus(_ string, _ int64,
	s *serverState) (logEntryStatus, int64, error) {
	return invalid, s.leaderID, fmt.Errorf(wrongRoleErrFmt, "follower")
}

func (f *followerRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) {
	panic(fmt.Sprintf(roleErrCallFmt, "finalizeElection", "follower"))
}

func (f *followerRole) makeCandidate(to time.Duration, s *serverState) bool {
	// Cannot become a candidate if time since last change is less than timeout
	d := time.Since(s.lastModified)
	if d < to {
		return false
	}

	// Advance term and update server state
	newTerm := s.currentTerm() + 1
	s.updateServerState(candidate, newTerm, s.serverID, invalidServerID)
	return true
}

func (f *followerRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm > serverTerm {
		return false
	}

	// If term changed, set votedFor to invalid server ID
	_, votedFor := s.votedFor()
	if currentTerm < serverTerm {
		votedFor = invalidServerID
	}

	// Update server state and return
	s.updateServerState(follower, serverTerm, votedFor, serverID)
	return true
}

func (f *followerRole) processAppendEntryEvent(_, _, _ int64,
	_ *serverState) bool {
	return false
}

func (f *followerRole) requestVote(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term and ID of server that has received this server's vote
	// in the current term
	currentTerm, votedFor := s.votedFor()
	if (currentTerm > serverTerm) ||
		(serverTerm == currentTerm && votedFor != invalidServerID) {
		return currentTerm, false
	}

	// Grant vote and update server state
	leaderID := s.leaderID
	if serverTerm != currentTerm {
		leaderID = invalidServerID
	}
	s.updateServerState(follower, serverTerm, serverID, leaderID)
	return serverTerm, true
}

func (f *followerRole) sendHeartbeat(time.Duration, int64, *serverState) {
	return
}

func (f *followerRole) startElection(
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "follower"))
}
