package domain

import "time"

type leaderRole struct{}

func (f *leaderRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()

	// No two leaders can be elected during the same term
	if currentTerm == serverTerm {
		panic("cannot have more than one leader per term")
	}

	// If calling server term is greater than local server term, role should
	// have been changed to follower: failure to do so points to a bug
	if currentTerm < serverTerm {
		panic("leader has not transitioned to follower before" +
			"calling appendEntry")
	}
	return currentTerm, false
}

func (f *leaderRole) makeCandidate(s *serverState) {}

func (f *leaderRole) makeFollower(serverTerm int64, s *serverState) {
	// Get current term
	currentTerm := s.currentTerm()

	// If server term greater than current term, change role to follower and
	// update term
	if currentTerm < serverTerm {
		s.role = follower
		s.updateTerm(serverTerm)
	}
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
