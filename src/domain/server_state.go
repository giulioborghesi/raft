package domain

import (
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
)

const (
	follower = iota
	leader
	candidate
)

var (
	// Only one instance of serverState per server should exist
	state *serverState
)

func init() {
	state = new(serverState)
	state.lastModified = time.Now()
	state.role = follower
}

// getServerState returns a pointer to the unique instance of serverState
func getServerState() *serverState {
	return state
}

// serverState represents the state of a Raft server
type serverState struct {
	dao          server_state_dao.ServerStateDao
	lastModified time.Time
	role         int
	serverID     int64
}

// currentTerm returns the current term ID
func (s *serverState) currentTerm() int64 {
	return s.dao.CurrentTerm()
}

func (s *serverState) updateTerm(serverTerm int64) {
	if err := s.dao.UpdateTerm(serverTerm); err != nil {
		panic(err)
	}
}

// updateTermVotedFor updates both the current term ID and the ID of the server
// that has received this server's vote in the current term
func (s *serverState) updateTermVotedFor(serverTerm int64, serverID int64) {
	if err := s.dao.UpdateTermVotedFor(serverTerm, serverID); err != nil {
		panic(err)
	}
}

// updateVotedFor updates the ID of the server that has received this server's
// vote in the current term
func (s *serverState) updateVotedFor(serverID int64) {
	if err := s.dao.UpdateVotedFor(serverID); err != nil {
		panic(err)
	}
}

// votedFor returns the current term ID and the ID of the server that received
// this server's vote during this term
func (s *serverState) votedFor() (int64, int64) {
	return s.dao.VotedFor()
}
