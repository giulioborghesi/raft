package domain

import (
	"time"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// MakeServerState creates a serverState instance
func MakeServerState(dao datasources.ServerStateDao,
	log abstractRaftLog, serverID int64) *serverState {
	s := new(serverState)
	s.dao = dao
	s.log = log
	s.lastModified = time.Now()
	s.role = follower
	s.targetCommitIndex = invalidLogEntryIndex
	s.leaderID = invalidServerID
	s.serverID = serverID
	return s
}

// serverState encapsulates the state of a Raft server
type serverState struct {
	dao               datasources.ServerStateDao
	log               abstractRaftLog
	lastModified      time.Time
	role              int
	targetCommitIndex int64
	serverID          int64
	leaderID          int64
}

// currentTerm returns the current term ID
func (s *serverState) currentTerm() int64 {
	return s.dao.CurrentTerm()
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

// updateCommitIndex computes the new commit index based on the match indices
// of the remote servers
func (s *serverState) updateTargetCommitIndex(matchIndices []int64) {
	// Make a deep copy of the match indices slice
	indices := make([]int64, len(matchIndices))
	copy(indices, matchIndices)

	// Update match index of local server and sort the resulting slice
	utils.SortInt64List(indices)

	// Return index corresponding to half minus one of the servers
	k := (len(indices) - 1) / 2
	if s.log.entryTerm(indices[k]) == s.currentTerm() {
		s.targetCommitIndex = utils.MaxInt64(s.targetCommitIndex, indices[k])
	}
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
