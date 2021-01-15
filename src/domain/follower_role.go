package domain

import (
	"fmt"
	"time"
)

const (
	followerErrIdFmt   = "invalid leader id: expected: %d, actual: %d"
	followerErrTermFmt = "invalid term number: expected: %d, actual: %d"
)

type followerRole struct{}

func (f *followerRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Prior to calling appendEntry, both current term and leader ID have
	// been updated
	currentTerm := s.currentTerm()
	if currentTerm != serverTerm {
		panic(fmt.Sprintf(followerErrTermFmt, currentTerm, serverTerm))
	}

	if serverID != s.leaderID {
		panic(fmt.Sprintf(followerErrIdFmt, s.leaderID, serverID))
	}

	// Return true
	return currentTerm, true
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

	// Get current term
	currentTerm := s.currentTerm()

	// Change role to candidate, update term, voted for and last modified
	s.role = candidate
	s.updateTermVotedFor(currentTerm+1, s.serverID)
	s.leaderID = invalidLeaderID
	s.lastModified = time.Now()
	return true
}

func (f *followerRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm > serverTerm {
		return false
	}

	// Update term and leader ID
	if currentTerm < serverTerm {
		s.updateTerm(serverTerm)
		s.leaderID = serverID
	} else if s.leaderID != serverID {
		panic(fmt.Sprintf(followerErrIdFmt, s.leaderID, serverID))
	}
	return true
}

func (f *followerRole) requestVote(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term and ID of server that has received this server's vote
	// in the current term
	currentTerm, votedFor := s.votedFor()
	if (currentTerm > serverTerm) ||
		(serverTerm == currentTerm && votedFor != invalidLeaderID) {
		return currentTerm, false
	}

	// Grant vote and update term if needed
	if serverTerm == currentTerm {
		s.updateVotedFor(serverID)
	} else {
		currentTerm = serverTerm
		s.updateTermVotedFor(currentTerm, serverID)
		s.leaderID = invalidLeaderID
	}
	s.lastModified = time.Now()
	return currentTerm, true
}

func (f *followerRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "follower"))
}
