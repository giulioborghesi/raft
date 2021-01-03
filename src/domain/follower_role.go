package domain

import "time"

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
	localServerID int64) bool {
	return false
}
