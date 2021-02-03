package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	leaderErrMultFmt = "multiple leaders in the same term detected"
)

// encodeEntry creates a unique key for a newly inserted log entry by combining
// the log term with the log index
func encodeEntry(entryTerm int64, entryIndex int64) string {
	return fmt.Sprintf("%d#%d", entryTerm, entryIndex)
}

// decodeEntry decodes a log entry unique key and returns the corresponding log
// term and log entry index
func decodeEntry(key string) (int64, int64) {
	idx := strings.LastIndex(key, "#")

	// Extract entry term
	entryTerm, err := strconv.Atoi(key[:idx])
	if err != nil {
		panic(fmt.Errorf("decodeEntry: %v", err))
	}

	// Extract entry index
	entryIndex, err := strconv.Atoi(key[idx+1:])
	if err != nil {
		panic(fmt.Errorf("decodeEntry: %v", err))
	}

	// Return decoded entry term and index
	return int64(entryTerm), int64(entryIndex)
}

// leaderRole implements the serverRole interface for a leader server
type leaderRole struct {
	replicators  []abstractEntryReplicator
	matchIndices []int64
}

func (l *leaderRole) appendEntry(_ []*service.LogEntry, _, _, _, _, _ int64,
	s *serverState) (int64, bool) {
	panic(fmt.Sprintf(roleErrCallFmt, "appendEntry", "leader"))
}

func (l *leaderRole) appendNewEntry(entry *service.LogEntry, commitIndex int64,
	s *serverState) (string, int64, error) {
	// Append new entry to log
	nextLogIndex := s.log.appendEntry(entry)

	// Replicate log entry and return unique entry key to client
	l.sendEntries(s.currentTerm(), nextLogIndex, commitIndex, s)
	return encodeEntry(s.currentTerm(), nextLogIndex), s.leaderID, nil
}

func (l *leaderRole) entryStatus(key string, commitIndex int64,
	s *serverState) (logEntryStatus, int64, error) {
	// Decode log entry key
	logNextIndex := s.log.nextIndex()
	entryTerm, entryIndex := decodeEntry(key)

	// Entry must be a valid one
	if entryTerm > s.currentTerm() || entryIndex > logNextIndex {
		return invalid, s.leaderID, fmt.Errorf("entry is not valid")
	}

	// Entry is lost if a term mismatch is detected
	actualEntryTerm := s.log.entryTerm(entryIndex)
	if entryTerm != actualEntryTerm {
		return lost, s.leaderID, nil
	}

	// Entry is committed if its index is at most the commit index
	if entryIndex <= commitIndex {
		return committed, s.leaderID, nil
	}
	return appended, s.leaderID, nil
}

func (l *leaderRole) finalizeElection(_ int64, _ []requestVoteResult,
	_ *serverState) bool {
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
	matchIndex int64, serverID int64, s *serverState) bool {
	// Append entry succeeded and server term did not change
	if appendTerm == s.currentTerm() && matchIndex != invalidLogID {
		l.matchIndices[serverID] = matchIndex
		s.updateTargetCommitIndex(l.matchIndices)
		return true
	}

	// Switch role to follower if a new term was detected
	if appendTerm > s.currentTerm() {
		s.updateServerState(follower, appendTerm, invalidServerID,
			invalidServerID)
	}
	return false
}

func (l *leaderRole) requestVote(serverTerm int64, serverID int64,
	lastEntryTerm int64, lastEntryIndex int64, s *serverState) (int64, bool) {
	// Get current term
	currentTerm := s.currentTerm()
	if currentTerm >= serverTerm {
		return currentTerm, false
	}

	// Server will grant its vote only if remote log is current
	current := isRemoteLogCurrent(s.log, lastEntryTerm, lastEntryIndex)
	var votedFor int64 = invalidServerID
	if current {
		votedFor = serverID
	}
	s.updateServerState(follower, serverTerm, votedFor, invalidServerID)
	return serverTerm, current
}

// sendEntries notifies the entry replicator about the new log entries to
// append to the remote server log
func (l *leaderRole) sendEntries(appendTerm int64, logNextIndex int64,
	commitIndex int64, s *serverState) {
	for _, replicator := range l.replicators {
		replicator.appendEntry(appendTerm, logNextIndex, commitIndex)
	}
	s.lastModified = time.Now()
}

func (l *leaderRole) sendHeartbeat(to time.Duration, commitIndex int64,
	s *serverState) {
	d := time.Since(s.lastModified)
	if d >= to {
		l.sendEntries(s.currentTerm(), s.log.nextIndex(), commitIndex, s)
	}
}

func (l *leaderRole) startElection(
	s *serverState) []chan requestVoteResult {
	panic(fmt.Sprintf(roleErrCallFmt, "startElection", "leader"))
}
