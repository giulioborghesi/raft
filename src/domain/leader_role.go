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

// computeCommitIndex computes the maximum index for which at least
// half plus one of the servers has that log entry
func computeCommitIndex(matchIndices []int64, s *serverState) int64 {
	// Make a copy of the match indices
	indices := []int64{}
	copy(indices, matchIndices)

	// Update match index of local server and sort the resulting slice
	utils.SortInt64List(indices)

	// Return index corresponding to half minus one of the servers
	k := (len(indices) - 1) / 2
	return indices[k]
}

type leaderRole struct {
	appenders     []abstractEntryReplicator
	appendResults *list.List
	matchIndices  []int64
}

func (l *leaderRole) appendEntry(serverTerm int64,
	serverID int64, s *serverState) (int64, bool) {
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

func (l *leaderRole) makeFollower(serverTerm int64, serverID int64,
	leaderID int64, s *serverState) {
	// Pending append entries request must return false
	for l.appendResults.Len() > 0 {
		front := l.appendResults.Front()
		front.Value.(appendEntryResult).success <- false
		l.appendResults.Remove(front)
	}

	// Update server state
	s.role = follower
	s.updateTermVotedFor(serverTerm, serverID)
	s.leaderID = leaderID
	s.lastModified = time.Now()
}

func (l *leaderRole) notifyAppendEntrySuccess(serverTerm int64,
	matchIndex int64, remoteServerID int64, s *serverState) {
	if s.currentTerm() != serverTerm {
		return
	}

	// Update match index and compute new commit index
	l.matchIndices[remoteServerID] = matchIndex
	newCommitIndex := computeCommitIndex(l.matchIndices, s)

	// TODO: advance local commit index
	s.commitIndex = newCommitIndex

	// Notify waiting clients
	for l.appendResults.Len() > 0 {
		entry := l.appendResults.Front()
		result := entry.Value.(appendEntryResult)
		if result.index > newCommitIndex {
			break
		}
		result.success <- true
		l.appendResults.Remove(entry)
	}
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
	l.makeFollower(serverTerm, invalidServerID, serverID, s)
	return true
}

func (l *leaderRole) requestVote(serverTerm int64, serverID int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server term greater than local term, cast vote and become follower
	l.makeFollower(serverTerm, serverID, invalidLeaderID, s)
	return serverTerm, true
}

func (l *leaderRole) startElection(servers []string,
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "leader"))
}
