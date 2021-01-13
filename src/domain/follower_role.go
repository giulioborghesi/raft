package domain

import (
	"fmt"
	"time"
)

const (
	followerErrFmt = "%s should not be called when a server is a follower"
)

type followerRole struct{}

func (f *followerRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm > serverTerm {
		return currentTerm, false
	}

	// Update current term if needed and return
	if currentTerm < serverTerm {
		currentTerm = serverTerm
		s.updateTerm(currentTerm)
	}
	s.lastModified = time.Now()
	return currentTerm, true
}

func (f *followerRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) {
	panic(fmt.Sprintf(followerErrFmt, "finalizeElection"))
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
	s.lastModified = time.Now()
	return true
}

func (f *followerRole) makeFollower(serverTerm int64, s *serverState) bool {
	return true
}

func (f *followerRole) requestVote(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term and ID of server that has received this server's vote
	// in the current term
	currentTerm, votedFor := s.votedFor()
	if (currentTerm > serverTerm) ||
		(serverTerm == currentTerm && votedFor != -1) {
		return currentTerm, false
	}

	// Grant vote and update term if needed
	if serverTerm == currentTerm {
		s.updateVotedFor(serverID)
	} else {
		currentTerm = serverTerm
		s.updateTermVotedFor(currentTerm, serverID)
	}
	s.lastModified = time.Now()
	return currentTerm, true
}

func (f *followerRole) startElection(currentTerm int64,
	localServerID int64) []chan requestVoteResult {
	panic(fmt.Sprintf(followerErrFmt, "startElection"))
}
