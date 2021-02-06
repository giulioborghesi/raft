package domain

import (
	"fmt"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	followerErrIdFmt   = "invalid leader id: expected: %d, actual: %d"
	followerErrTermFmt = "invalid term number: expected: %d, actual: %d"
)

// checkLeader checks whether the leader ID is acceptable or not. The leader ID
// is acceptable if it is unknown or equal to the ID of the server that sent
// the append entry RPC
func checkLeader(currentTerm int64, serverTerm int64, leaderID int64,
	actualLeaderID int64) bool {
	if currentTerm == serverTerm && leaderID != invalidServerID {
		return leaderID == actualLeaderID
	}
	return true
}

// followerRole implements the serverRole interface for a follower server
type followerRole struct{}

func (f *followerRole) appendEntry(entries []*service.LogEntry, serverTerm,
	serverID, prevLogTerm, prevLogIndex int64, commitIndex int64,
	s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm != serverTerm {
		panic(fmt.Sprintf(followerErrTermFmt, currentTerm, serverTerm))
	}

	// Server ID
	if serverID != s.leaderID {
		panic(fmt.Sprintf(followerErrIdFmt, s.leaderID, serverID))
	}

	// Try appending log entries to log
	success := s.log.appendEntries(entries, prevLogTerm, prevLogIndex)
	if success {
		s.targetCommitIndex = commitIndex
	}
	return currentTerm, success
}

func (f *followerRole) appendNewEntry(_ string, _ int64,
	s *serverState) (string, int64, error) {
	return emptyString, s.leaderID,
		fmt.Errorf(notAvailableErrFmt, "appendNewEntry", roleName(follower))
}

func (f *followerRole) entryStatus(_ string, _ int64,
	s *serverState) (logEntryStatus, int64, error) {
	return invalid, s.leaderID,
		fmt.Errorf(notAvailableErrFmt, "entryStatus", roleName(follower))
}

func (f *followerRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) bool {
	panic(fmt.Sprintf(notAvailableErrFmt, "finalizeElection",
		roleName(follower)))
}

func (f *followerRole) makeCandidate(to time.Duration, s *serverState) bool {
	// Cannot become a candidate if time since last change is less than timeout
	d := time.Since(s.lastModified)
	if d < to {
		return false
	}

	// Advance term and update server state
	newTerm := s.currentTerm() + 1
	s.updateServerState(candidate, newTerm, s.serverID, invalidServerID)
	return true
}

func (f *followerRole) prepareAppend(serverTerm int64, serverID int64,
	s *serverState) bool {
	// Get current term
	currentTerm, votedFor := s.votedFor()
	if currentTerm > serverTerm {
		return false
	}

	// If term has not changed, leader ID must be equal to remote server ID or
	// be unknown: if not, multiple leaders exist in the same term
	if ok := checkLeader(currentTerm, serverTerm, s.leaderID, serverID); !ok {
		panic(fmt.Sprintf(multipleLeadersErrFmt, currentTerm))
	}

	// If term changed, set votedFor to invalid server ID
	if currentTerm < serverTerm {
		votedFor = invalidServerID
	}

	// Update server state and return
	s.updateServerState(follower, serverTerm, votedFor, serverID)
	return true
}

func (f *followerRole) processAppendEntryEvent(_, _, _ int64,
	_ *serverState) bool {
	return false
}

func (f *followerRole) requestVote(serverTerm int64, serverID int64,
	lastEntryTerm int64, lastEntryIndex int64, s *serverState) (int64, bool) {
	// Get current term and ID of server that has received this server's vote
	// in the current term
	currentTerm, votedFor := s.votedFor()
	if currentTerm > serverTerm {
		return currentTerm, false
	} else if serverTerm == currentTerm && votedFor != invalidServerID {
		return currentTerm, votedFor == serverID
	}

	// Grant vote only if remote server log is not current
	current := isRemoteLogCurrent(s.log, lastEntryTerm, lastEntryIndex)
	votedFor = invalidServerID
	if current {
		votedFor = serverID
	}

	// Update server state and return
	leaderID := s.leaderID
	if serverTerm != currentTerm {
		leaderID = invalidServerID
	}
	s.updateServerState(follower, serverTerm, votedFor, leaderID)
	return serverTerm, current
}

func (f *followerRole) sendHeartbeat(time.Duration, int64, *serverState) {
	return
}

func (f *followerRole) startElection(
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(notAvailableErrFmt, "startElection",
		roleName(follower)))
}
