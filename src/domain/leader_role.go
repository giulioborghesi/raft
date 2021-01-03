package domain

import "time"

type leaderRole struct{}

func (f *leaderRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm == serverTerm {
		panic("cannot have more than one leader per term")
	}

	// appendEntry call came from a previous leader
	if currentTerm > serverTerm {
		return currentTerm, false
	}

	// appendEntry call came from new leader
	s.role = follower
	s.updateTerm(serverTerm)
	s.lastModified = time.Now()
	return serverTerm, true
}

func (f *leaderRole) requestVote(serverTerm int64, serverID int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()

	// Candidate term is greater than server term, grant vote
	if currentTerm < serverTerm {
		s.role = follower
		s.updateTerm(serverTerm)
		s.lastModified = time.Now()
		return serverTerm, true
	}

	// Candidate term is not greater than server term, deny vote
	return currentTerm, false
}

func (f *leaderRole) startElection(currentTerm int64,
	localServerID int64) bool {
	return false
}
