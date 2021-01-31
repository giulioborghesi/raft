package domain

import (
	"sync"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

var (
	entryAppenderTimeout time.Duration = 50 * time.Millisecond
)

// abstractEntryReplicator defines an interface for replicating local log
// entries in a remote server
type abstractEntryReplicator interface {
	// appendEntry replicates the log entry with the specified index and log
	// entry term in the remote server. The replication may be asynchronous
	appendEntry(int64, int64, int64)
}

// entryReplicator implements the abstractEntryReplicator interfaces.
// Replication is conducted asynchronously to improve performance
type entryReplicator struct {
	active                                     bool
	matchIndex, nextIndex, logNextIndex        int64
	appendTerm, lastAppendTerm, lastLeaderTerm int64
	commitIndex                                int64
	remoteServerID                             int64

	done    chan bool
	client  clients.AbstractRaftClient
	service AbstractRaftService
	*sync.Cond
}

// MakeEntryReplicator creates an instance of entryAppender. appendTerm will be
// initialized with an invalid term ID to ensure that log entries starts to be
// replicated only after receiving a signal from the current leader
func MakeEntryReplicator(remoteServerID int64, c clients.AbstractRaftClient,
	s AbstractRaftService) *entryReplicator {
	// Create entry appender and initialize appendTerm to invalid
	a := &entryReplicator{active: false, client: c, service: s,
		remoteServerID: remoteServerID}
	a.appendTerm = invalidTermID
	a.matchIndex = invalidLogID

	// Initialize condition variable
	m := sync.Mutex{}
	a.Cond = sync.NewCond(&m)

	// Activate entry appender and return
	a.start()
	return a
}

func (a *entryReplicator) appendEntry(appendTerm int64,
	logNextIndex int64, commitIndex int64) {
	a.L.Lock()
	defer a.L.Unlock()

	a.appendTerm = appendTerm
	a.logNextIndex = logNextIndex
	a.commitIndex = commitIndex
	a.Signal()
}

// processEntries replicates log entries on the remote server asynchronously.
func (a *entryReplicator) processEntries() {
	for a.active {
		a.L.Lock()

		// Must wait when term has expired or there is no entry to append
		for a.appendTerm < a.lastLeaderTerm || (a.appendTerm ==
			a.lastAppendTerm && (a.matchIndex+1) == a.logNextIndex) {
			a.Wait()

			// If server is no longer active, stop
			if a.active == false {
				return
			}

			// Handle heartbeats
			if a.appendTerm >= a.lastLeaderTerm {
				break
			}
		}

		// Store append term and log index for later use
		appendTerm := a.appendTerm
		logNextIndex := a.logNextIndex
		a.L.Unlock()

		// Update the match index
		if success := a.updateMatchIndex(appendTerm, logNextIndex); !success {
			continue
		}

		// Fetch the log entries to send to the remote server
		entries, leaderTerm, prevEntryTerm := a.service.entries(a.nextIndex)
		if leaderTerm > a.lastLeaderTerm {
			a.lastLeaderTerm = leaderTerm
			continue
		}

		// Send the log entries to the remote server
		if ok := a.sendEntries(entries, prevEntryTerm, a.matchIndex); ok {
			a.service.processAppendEntryEvent(appendTerm, a.matchIndex,
				a.remoteServerID)
		}
	}
}

// updateMatchIndex finds the match index. This method performs non-trivial
// work only when the current term exceeds the last known term
func (a *entryReplicator) updateMatchIndex(appendTerm int64,
	logNextIndex int64) bool {
	// Append term did not change from last append, nothing to do
	if appendTerm == a.lastAppendTerm {
		return true
	}

	// A new term has started, reset appender state
	a.resetState(appendTerm, logNextIndex)

	// Update match index
	prevEntryIndex := a.nextIndex - 1
	done := a.nextIndex == a.matchIndex+1
	for !done {
		leaderTerm, prevEntryTerm := a.service.entryInfo(prevEntryIndex)
		if leaderTerm > appendTerm {
			a.updateLeaderTerm(leaderTerm)
			return false
		}

		// Send empty entry to remote server
		if ok := a.sendEntries(nil, prevEntryTerm, prevEntryIndex); !ok {
			return false
		}

		// Update loop condition
		prevEntryIndex = a.nextIndex - 1
		done = a.nextIndex == a.matchIndex+1
	}
	return true
}

func (a *entryReplicator) resetState(appendTerm int64, nextIndex int64) {
	a.lastAppendTerm = appendTerm
	a.lastLeaderTerm = a.lastAppendTerm
	a.matchIndex = invalidLogID
	a.nextIndex = nextIndex
}

func (a *entryReplicator) sendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	// Send entries to remote server
	remoteTerm, success :=
		a.client.AppendEntry(entries, a.lastAppendTerm, prevEntryTerm,
			prevEntryIndex, a.commitIndex)

	// Remote server term greater than local term, return
	if remoteTerm > a.lastAppendTerm {
		a.updateLeaderTerm(remoteTerm)
		return false
	}

	// Update entry appender state and return
	if success {
		a.matchIndex = prevEntryIndex + int64(len(entries))
		a.nextIndex = a.matchIndex + 1
	} else {
		a.nextIndex -= 1
	}
	return true
}

func (a *entryReplicator) start() {
	a.L.Lock()
	defer a.L.Unlock()

	// Nothing to do if appender is already active
	if a.active {
		return
	}

	// Launch process entries asynchronously
	a.active = true
	a.done = make(chan bool)
	go func() {
		a.processEntries()
		a.done <- true
	}()
}

func (a *entryReplicator) stop() {
	a.L.Lock()

	// Nothing to do if appender is already inactive
	if a.active == false {
		return
	}

	// Set appender to inactive and wait for processEntries to return
	a.active = false
	a.Signal()

	// Must release lock before synchronizing
	a.L.Unlock()
	<-a.done
}

func (a *entryReplicator) updateLeaderTerm(newLeaderTerm int64) {
	a.lastLeaderTerm = newLeaderTerm
	a.service.processAppendEntryEvent(newLeaderTerm, invalidLogID,
		invalidServerID)
}
