package domain

import (
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testLeaderActive   = false
	testLeaderLeaderID = 2
	testLeaderServerID = testLeaderLeaderID
	testLeaderRemoteID = 4
	testLeaderVotedFor = testLeaderLeaderID
)

func TestEncodeDecodeKey(t *testing.T) {
	var entryTerm int64 = 6
	var entryIndex int64 = 15

	// Encode and decode log entry info
	s := encodeEntry(entryTerm, entryIndex)
	actualEntryTerm, actualEntryIndex := decodeEntry(s)

	if actualEntryTerm != entryTerm {
		t.Fatalf(invalidTermErrFmt, entryTerm, actualEntryTerm)
	}

	if actualEntryIndex != entryIndex {
		t.Fatalf(invalidEntryIndexErrFmt, entryTerm, actualEntryIndex)
	}
}

func TestLeaderMethodsThatPanic(t *testing.T) {
	// Initialize server state and leader
	l := new(leaderRole)
	s := makeTestServerState(testStartingTerm, testLeaderVotedFor,
		testLeaderServerID, testLeaderLeaderID, leader, testLeaderActive)

	// Test appendEntry
	appendEntry := func() {
		l.appendEntry(nil, 0, 0, 0, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntry)

	// Test finalizeElection
	finalizeElection := func() {
		l.finalizeElection(0, []requestVoteResult{}, s)
	}
	utils.AssertPanic(t, "finalizeElection", finalizeElection)

	// Test startElection
	startElection := func() {
		l.startElection(s)
	}
	utils.AssertPanic(t, "startElection", startElection)
}

func TestLeaderMakeCandidate(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := makeTestServerState(testStartingTerm, testLeaderVotedFor,
		testLeaderServerID, testLeaderLeaderID, leader, testLeaderActive)

	// makeCandidate always returns false
	if success := l.makeCandidate(time.Millisecond, s); success {
		t.Fatalf(methodExpectedToFailErrFmt, "makeCandidate")
	}

	validateServerState(s, leader, testStartingTerm, testLeaderLeaderID,
		testLeaderVotedFor, t)
}

func TestLeaderPrepareAppend(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := makeTestServerState(testStartingTerm, testLeaderVotedFor,
		testLeaderServerID, testLeaderLeaderID, leader, testLeaderActive)

	// Initialize remote server term
	remoteServerTerm := int64(initialTerm - 1)

	// Server term less than local term, prepareAppend should fail
	if success := l.prepareAppend(remoteServerTerm,
		testLeaderRemoteID, s); success {
		t.Fatalf(methodExpectedToFailErrFmt, "prepareAppend")
	}

	validateServerState(s, leader, testStartingTerm, testLeaderLeaderID,
		testLeaderVotedFor, t)

	// prepareAppend should change server role to follower and return true
	remoteServerTerm = testStartingTerm + 1
	if success := l.prepareAppend(remoteServerTerm,
		testLeaderRemoteID, s); !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "prepareAppend")
	}

	validateServerState(s, follower, remoteServerTerm, invalidServerID,
		testLeaderRemoteID, t)

	// prepareAppend should panic when two leaders in same term are detected
	prepareAppend := func() {
		remoteServerTerm = s.currentTerm()
		l.prepareAppend(remoteServerTerm, testLeaderRemoteID-1, s)
	}
	utils.AssertPanic(t, "prepareAppend", prepareAppend)
}

func TestLeaderRequestVote(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := makeTestServerState(testStartingTerm, testLeaderVotedFor,
		testLeaderServerID, testLeaderLeaderID, leader, testLeaderActive)

	// Initialize remote server term
	remoteServerTerm := int64(initialTerm - 1)

	// Server term less than local term, requestVote should fail
	_, success := l.requestVote(remoteServerTerm,
		testLeaderRemoteID, invalidTerm, invalidLogEntryIndex, s)

	if success {
		t.Fatalf(methodExpectedToFailErrFmt, "requestVote")
	}

	validateServerState(s, leader, testStartingTerm, testLeaderLeaderID,
		testLeaderLeaderID, t)

	// requestVote should change server role to follower and return true
	remoteServerTerm = testStartingTerm + 1
	_, success = l.requestVote(remoteServerTerm,
		testLeaderRemoteID, initialTerm, invalidLogEntryIndex, s)

	if !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "requestVote")
	}

	validateServerState(s, follower, remoteServerTerm, testLeaderRemoteID,
		invalidServerID, t)
}
