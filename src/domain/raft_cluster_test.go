package domain

import (
	"testing"
	"time"
)

func TestRaftClusterElection(t *testing.T) {
	// Create Raft cluster simulator
	s := makeRaftClusterSimulator([]int64{0, 0, 0},
		[]int64{-1, -1, -1}, [][]int64{{}, {}, {}})

	// Server 0 starts an election and gets elected
	leaderID := int64(0)
	s.startElection(leaderID, time.Duration(0))
	time.Sleep(testSleepTime)

	validateResults([]int64{0, 0, 0}, s.castedVotes(), t)
	validateResults([]int64{leader, follower, follower}, s.roles(), t)
	validateResults([]int64{1, 1, 1}, s.terms(), t)

	// Server 1 starts an election but does not get elected
	leaderID = int64(1)
	s.startElection(leaderID, time.Minute)
	time.Sleep(testSleepTime)

	validateResults([]int64{0, 0, 0}, s.castedVotes(), t)
	validateResults([]int64{leader, follower, follower}, s.roles(), t)
	validateResults([]int64{1, 1, 1}, s.terms(), t)

	// Server 1 starts an election and gets elected
	s.startElection(leaderID, time.Duration(0))
	time.Sleep(testSleepTime)

	validateResults([]int64{1, 1, 1}, s.castedVotes(), t)
	validateResults([]int64{follower, leader, follower}, s.roles(), t)
	validateResults([]int64{2, 2, 2}, s.terms(), t)

	// One connection fails. Server 2 starts an election gets elected
	leaderID = int64(2)
	s.setConnections(leaderID, []bool{true, false})
	s.startElection(leaderID, time.Duration(0))
	time.Sleep(testSleepTime)

	validateResults([]int64{2, 1, 2}, s.castedVotes(), t)
	validateResults([]int64{follower, leader, leader}, s.roles(), t)
	validateResults([]int64{3, 2, 3}, s.terms(), t)

	// Restore connections with all servers
	s.restoreConnections(leaderID)
	time.Sleep(testSleepTime)

	validateResults([]int64{2, -1, 2}, s.castedVotes(), t)
	validateResults([]int64{follower, follower, leader}, s.roles(), t)
	validateResults([]int64{3, 3, 3}, s.terms(), t)

	// Teardown cluster
	s.tearDown()
}

func TestRaftClusterElectionAndAppend(t *testing.T) {
	// Create Raft cluster simulator
	s := makeRaftClusterSimulator([]int64{1, 4, 5},
		[]int64{-1, -1, -1}, [][]int64{{1, 1, 1}, {1, 4, 4}, {1, 5, 5, 5}})

	// Server 2 starts an election and gets elected
	leaderID := int64(2)
	s.startElection(leaderID, time.Duration(0))
	time.Sleep(testSleepTime)

	validateResults([]int64{leaderID, leaderID, leaderID}, s.castedVotes(), t)
	validateResults([]int64{follower, follower, leader}, s.roles(), t)
	validateResults([]int64{6, 6, 6}, s.terms(), t)

	for _, log := range s.logs() {
		validateResults(log, []int64{1, 5, 5, 5}, t)
	}

	// A new entry is appended to the leader's log
	s.applyCommandAsync(leaderID)
	time.Sleep(testSleepTime)

	validateResults([]int64{leaderID, leaderID, leaderID}, s.castedVotes(), t)
	validateResults([]int64{follower, follower, leader}, s.roles(), t)
	validateResults([]int64{6, 6, 6}, s.terms(), t)

	for _, log := range s.logs() {
		validateResults(log, []int64{1, 5, 5, 5, 6}, t)
	}

	// Teardown cluster
	s.tearDown()
}

func TestRaftClusterFailedElection(t *testing.T) {
	// Create Raft cluster simulator
	s := makeRaftClusterSimulator([]int64{8, 4, 5},
		[]int64{-1, -1, -1}, [][]int64{{1, 1, 1}, {1, 4, 4}, {1, 5, 5, 5}})

	// Server 2 starts an election but does not gets elected and goes back
	// to being a follower
	leaderID := int64(2)
	s.startElection(leaderID, time.Duration(0))
	time.Sleep(testSleepTime)

	validateResults([]int64{-1, leaderID, -1}, s.castedVotes(), t)
	validateResults([]int64{follower, follower, follower}, s.roles(), t)
	validateResults([]int64{8, 6, 8}, s.terms(), t)

	expectedLogs := [][]int64{{1, 1, 1}, {1, 4, 4}, {1, 5, 5, 5}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Teardown cluster
	s.tearDown()
}

func TestRaftClusterPaperScenario(t *testing.T) {
	// Create Raft cluster simulator
	s := makeRaftClusterSimulator([]int64{2, 2, 2, 2, 2},
		[]int64{0, 0, 0, 0, 0}, [][]int64{{1, 2}, {1, 2}, {1}, {1}, {1}})
	s.setServerRole(0, leader)

	// Server 4 starts an election. Server 4 can communicate with servers 3 and
	// 2, but not with servers 0 and 1. Server 0 is death
	s.setConnections(4, []bool{false, false, true, true})
	s.startElection(4, time.Duration(0))
	time.Sleep(testSleepTime)

	roles := []int64{leader, follower, follower, follower, leader}
	validateResults([]int64{0, 0, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{2, 2, 3, 3, 3}, s.terms(), t)

	expectedLogs := [][]int64{{1, 2}, {1, 2}, {1}, {1}, {1}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 4 loses connections with all servers and also receives an apply
	// command request from a client
	s.setConnections(4, []bool{false, false, false, false})
	_, _, _ = s.applyCommandAsync(4)
	time.Sleep(testSleepTime)

	roles = []int64{leader, follower, follower, follower, leader}
	validateResults([]int64{0, 0, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{2, 2, 3, 3, 3}, s.terms(), t)

	expectedLogs = [][]int64{{1, 2}, {1, 2}, {1}, {1}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 4 dies and server 0 comes back to life and starts an election
	s.setConnections(0, []bool{true, true, false, false})
	s.setServerRole(0, follower)
	s.startElection(0, time.Duration(0))
	time.Sleep(testSleepTime)

	roles = []int64{candidate, follower, follower, follower, leader}
	validateResults([]int64{0, 0, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{3, 3, 3, 3, 3}, s.terms(), t)

	expectedLogs = [][]int64{{1, 2}, {1, 2}, {1}, {1}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 0 was not elected. After timeout, it starts a second election
	s.startElection(0, time.Duration(0))
	time.Sleep(testSleepTime)

	roles = []int64{leader, follower, follower, follower, leader}
	validateResults([]int64{0, 0, 0, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{4, 4, 4, 3, 3}, s.terms(), t)

	expectedLogs = [][]int64{{1, 2}, {1, 2}, {1, 2}, {1}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 0 loses all connections with other servers while receiving an
	// apply command request from a client
	s.setConnections(0, []bool{false, false, false, false})
	_, _, _ = s.applyCommandAsync(0)
	time.Sleep(testSleepTime)

	roles = []int64{leader, follower, follower, follower, leader}
	validateResults([]int64{0, 0, 0, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{4, 4, 4, 3, 3}, s.terms(), t)

	expectedLogs = [][]int64{{1, 2, 4}, {1, 2}, {1, 2}, {1}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 4 now comes back to life
	s.restoreConnections(4)
	time.Sleep(testSleepTime)

	roles = []int64{leader, follower, follower, follower, follower}
	validateResults([]int64{0, 0, 0, 4, -1}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{4, 4, 4, 3, 4}, s.terms(), t)

	expectedLogs = [][]int64{{1, 2, 4}, {1, 2}, {1, 2}, {1, 3}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Server 4 starts an election and gets elected
	s.startElection(4, time.Duration(0))
	time.Sleep(testSleepTime)

	roles = []int64{follower, follower, follower, follower, leader}
	validateResults([]int64{-1, 4, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{5, 5, 5, 5, 5}, s.terms(), t)

	expectedLogs = [][]int64{{1, 3}, {1, 3}, {1, 3}, {1, 3}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Cluster state doesn't change when connections are re-established
	s.restoreConnections(0)

	roles = []int64{follower, follower, follower, follower, leader}
	validateResults([]int64{-1, 4, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{5, 5, 5, 5, 5}, s.terms(), t)

	expectedLogs = [][]int64{{1, 3}, {1, 3}, {1, 3}, {1, 3}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Sending an heartbeat won't change the cluster state
	s.sendHearthbeat(0)

	roles = []int64{follower, follower, follower, follower, leader}
	validateResults([]int64{-1, 4, 4, 4, 4}, s.castedVotes(), t)
	validateResults(roles, s.roles(), t)
	validateResults([]int64{5, 5, 5, 5, 5}, s.terms(), t)

	expectedLogs = [][]int64{{1, 3}, {1, 3}, {1, 3}, {1, 3}, {1, 3}}
	for i, log := range s.logs() {
		validateResults(expectedLogs[i], log, t)
	}

	// Teardown cluster
	s.tearDown()
}
