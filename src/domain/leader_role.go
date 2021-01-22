package domain

import (
	"container/list"
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	leaderErrMultFmt = "multiple leaders in the same term detected"
)

// computeNewCommitIndex computes an estimate of the commit index from the
// match indices of the remote servers
func computeNewCommitIndex(matchIndices []int64) int64 {
	// Make a copy of the match indices
	indices := []int64{}
	copy(indices, matchIndices)

	// Update match index of local server and sort the resulting slice
	utils.SortInt64List(indices)

	// Return index corresponding to half minus one of the servers
	k := (len(indices) - 1) / 2
	return indices[k]
}

// leaderRole implements the serverRole interface for a leader server
type leaderRole struct {
	appenders     []abstractEntryReplicator
	appendResults *list.List
	matchIndices  []int64
}

func (l *leaderRole) appendEntry(_, _, _, _ int64,
	s *serverState) (int64, bool) {
	panic(fmt.Sprintf(roleErrCallFmt, "appendEntry", "leader"))
}

func (l *leaderRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) {
	panic(fmt.Sprintf(roleErrCallFmt, "finalizeElection", "leader"))
}

func (l *leaderRole) makeCandidate(_ time.Duration, s *serverState) bool {
	// Leaders cannot transition to candidates
	return false
}

func (l *leaderRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm := s.currentTerm()

	if currentTerm == serverTerm {
		// Multiple leaders in the same term are not allowed
		panic(fmt.Sprintf(leaderErrMultFmt))
	} else if currentTerm > serverTerm {
		// Request received from a former leader, reject it
		return false
	}

	// Request received from new leader, transition to follower
	s.updateServerState(follower, serverTerm, invalidServerID, serverID)
	return true
}

func (l *leaderRole) processAppendEntryEvent(appendTerm int64,
	matchIndex int64, serverID int64, s *serverState) (int64, bool) {
	// Append entry succeeded and server term did not change
	if appendTerm == s.currentTerm() && matchIndex != invalidLogID {
		l.matchIndices[serverID] = matchIndex
		return computeNewCommitIndex(l.matchIndices), true
	}

	// Switch role to follower if a new term was detected
	if appendTerm > s.currentTerm() {
		s.updateServerState(follower, appendTerm, invalidServerID,
			invalidServerID)
	}
	return invalidLogID, false
}

func (l *leaderRole) requestVote(serverTerm int64, serverID int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server term greater than local term, cast vote and become follower
	s.updateServerState(follower, serverTerm, serverID, invalidServerID)
	return serverTerm, true
}

func (l *leaderRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "leader"))
}
