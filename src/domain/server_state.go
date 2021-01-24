package domain

import (
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	follower = iota
	leader
	candidate

	invalidLogID    = -1
	invalidServerID = -1
	invalidTermID   = -1
)

// makeServerState creates a serverState instance
func makeServerState(dao server_state_dao.ServerStateDao,
	log abstractRaftLog, serverID int64) *serverState {
	s := new(serverState)
	s.dao = dao
	s.log = log
	s.lastModified = time.Now()
	s.role = follower
	s.commitIndex = 0
	s.leaderID = invalidServerID
	s.serverID = serverID
	return s
}

// serverState encapsulates the state of a Raft server
type serverState struct {
	dao          server_state_dao.ServerStateDao
	log          abstractRaftLog
	lastModified time.Time
	role         int
	commitIndex  int64
	serverID     int64
	leaderID     int64
}

// currentTerm returns the current term ID
func (s *serverState) currentTerm() int64 {
	return s.dao.CurrentTerm()
}

// updateCommitIndex computes the new commit index based on the match indices
// of the remote servers
func (s *serverState) updateCommitIndex(matchIndices []int64) {
	// Make a copy of the match indices
	indices := []int64{}
	copy(indices, matchIndices)

	// Update match index of local server and sort the resulting slice
	utils.SortInt64List(indices)

	// Return index corresponding to half minus one of the servers
	k := (len(indices) - 1) / 2
	s.commitIndex = utils.MaxInt64(s.commitIndex, indices[k])
}

// updateServerState updates the server state according to the provided
// parameters
func (s *serverState) updateServerState(newRole int, newTerm int64,
	newVotedFor int64, newLeaderID int64) {
	s.role = newRole
	s.updateTermVotedFor(newTerm, newVotedFor)
	s.leaderID = newLeaderID
	s.lastModified = time.Now()
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
