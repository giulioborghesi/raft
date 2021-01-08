package domain

import "testing"

const (
	startingTerm     = 15
	startingVotedFor = 0
)

// mockedServerStateDao is a server state DAO mack
type mockedServerStateDao struct {
	currentTerm int64
	votedFor    int64
}

func (s *mockedServerStateDao) CurrentTerm() int64 {
	return s.currentTerm
}

func (s *mockedServerStateDao) UpdateTerm(serverTerm int64) error {
	s.currentTerm = serverTerm
	s.votedFor = -1
	return nil
}

func (s *mockedServerStateDao) UpdateTermVotedFor(serverTerm int64,
	serverID int64) error {
	s.currentTerm = serverTerm
	s.votedFor = serverID
	return nil
}

func (s *mockedServerStateDao) UpdateVotedFor(serverID int64) error {
	s.votedFor = serverID
	return nil
}

func (s *mockedServerStateDao) VotedFor() (int64, int64) {
	return s.currentTerm, s.votedFor
}

func TestAppendEntry(t *testing.T) {
	// Initialize mocked server state dao
	dao := new(mockedServerStateDao)
	dao.currentTerm = startingTerm
	dao.votedFor = startingVotedFor

	// Initialize server state
	s := new(serverState)
	s.dao = dao

	// Instantiate follower
	f := new(followerRole)

	// Server term is less than current term, append entry should fail
	var serverTerm int64 = 10
	var serverID int64 = 1
	term, ok := f.appendEntry(serverTerm, serverID, s)
	if term != startingTerm {
		t.Errorf("wrong term following appendEntry: expected %d, actual %d",
			startingTerm, term)
	}

	if ok {
		t.Errorf("appendEntry succeded although it was expected to fail")
	}

	// Server term is greater than current term, append entry will succeed
	serverTerm = 18
	term, ok = f.appendEntry(serverTerm, serverID, s)
	if term != serverTerm {
		t.Errorf("wrong term following appendEntry: expected %d, actual %d",
			serverTerm, term)
	}

	if !ok {
		t.Errorf("appendEntry failed although it was expected to succeed")
	}
}

func TestMakeCandidate(t *testing.T) {
	// Initialize mocked server state dao
	dao := new(mockedServerStateDao)
	dao.currentTerm = startingTerm
	dao.votedFor = startingVotedFor

	// Initialize server state
	s := new(serverState)
	s.dao = dao
	s.serverID = 2

	// Instantiate follower
	f := new(followerRole)

	// Change server role to candidate
	f.makeCandidate(s)
	term, votedFor := s.votedFor()
	if term != startingTerm+1 {
		t.Errorf("wrong term following makeCandidate: expected %d, actual %d",
			startingTerm+1, term)
	}

	if votedFor != 2 {
		t.Errorf("vote should have been casted for local server")
	}

	if s.role != candidate {
		t.Errorf("role should be candidate following makeCandidate")
	}
}

func TestRequestVote(t *testing.T) {
	// Initialize mocked server state dao
	dao := new(mockedServerStateDao)
	dao.currentTerm = startingTerm
	dao.votedFor = startingVotedFor

	// Initialize server state
	s := new(serverState)
	s.dao = dao

	// Instantiate follower
	f := new(followerRole)

	// Server term is less than current term, request vote will fail
	var serverTerm int64 = startingTerm - 2
	var serverID int64 = 1
	term, ok := f.requestVote(serverTerm, serverID, s)
	if term != startingTerm {
		t.Errorf("wrong term following requestVote: expected %d, actual %d",
			startingTerm, term)
	}

	if ok {
		t.Errorf("requestVote succeded although it was expected to fail")
	}

	// Server term is equal to current term but vote was already casted
	serverTerm = startingTerm
	term, ok = f.requestVote(serverTerm, serverID, s)
	if term != startingTerm {
		t.Errorf("wrong term following requestVote: expected %d, actual %d",
			startingTerm, term)
	}

	if ok {
		t.Errorf("requestVote succeded although it was expected to fail")
	}

	// Term is now updated and votedFor is reset to nil
	serverTerm = startingTerm + 5
	s.updateTerm(serverTerm)
	term, votedFor := s.votedFor()
	if term != serverTerm {
		t.Errorf("wrong term following updateTerm: expected %d, actual %d",
			serverTerm, term)
	}

	if votedFor != -1 {
		t.Errorf("votedFor was not reset following updateTerm")
	}

	// requestVote should now succeed
	term, ok = f.requestVote(serverTerm, serverID, s)
	if term != serverTerm {
		t.Errorf("wrong term returned by requestVote: expected %d, actual %d",
			serverTerm, term)
	}

	if !ok {
		t.Errorf("requestVote failed although it was expected to succeed")
	}

	// If server term is greater than current term, requestVote should succeed
	serverTerm += 2
	term, ok = f.requestVote(serverTerm, serverID, s)
	if term != serverTerm {
		t.Errorf("wrong term returned by requestVote: expected %d, actual %d",
			serverTerm, term)
	}

	if !ok {
		t.Errorf("requestVote failed although it was expected to succeed")
	}
}
