package domain

import (
	"fmt"
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

// makeEntryReplicator creates an instance of entryReplicator. The object will
// be initialized with an invalid term ID to ensure that log entries starts to
// be replicated only after receiving a signal from the current leader
func makeEntryReplicator(remoteServerID int64, c clients.AbstractRaftClient,
	s AbstractRaftService) *entryReplicator {
	// Create entry replicator
	r := &entryReplicator{active: false, client: c, service: s,
		remoteServerID: remoteServerID}

	// Initialize entry replicator state
	r.matchIndex = invalidLogID
	r.nextIndex = r.matchIndex + 1
	r.logNextIndex = r.nextIndex
	r.appendTerm = invalidTermID
	r.lastAppendTerm = invalidTermID
	r.lastLeaderTerm = invalidTermID
	r.commitIndex = invalidLogID

	// Initialize entry replicator condition variable and return
	m := sync.Mutex{}
	r.Cond = sync.NewCond(&m)
	return r
}

// NewEntryReplicator will create and start a new instance of entry replicator
func NewEntryReplicator(remoteServerID int64,
	c clients.AbstractRaftClient, s AbstractRaftService) *entryReplicator {
	// Create entry replicator
	r := makeEntryReplicator(remoteServerID, c, s)

	// Activate entry replicator and return
	r.start()
	return r
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

// nextLogEntryIndex returns the append term and the starting index of the next
// log entries to send to the remote server
func (e *entryReplicator) nextLogEntryIndex() (int64, int64, bool) {
	e.L.Lock()
	defer e.L.Unlock()

	// Must wait when term has expired or there is no entry to append
	for e.appendTerm < e.lastLeaderTerm || (e.appendTerm ==
		e.lastAppendTerm && (e.matchIndex+1) == e.logNextIndex) {
		e.Wait()

		// If server is no longer active, stop
		if e.active == false {
			return invalidTermID, invalidLogID, false
		}

		// Handle heartbeats
		if e.appendTerm >= e.lastLeaderTerm {
			break
		}
	}

	return e.appendTerm, e.logNextIndex, true
}

// processEntries replicates log entries on the remote server asynchronously.
func (a *entryReplicator) processEntries() {
	for a.active {
		// Fetch append term and starting index of next log entries to send
		appendTerm, logNextIndex, active := a.nextLogEntryIndex()
		if !active {
			break
		}

		// Update the match index
		if success := a.updateMatchIndex(appendTerm, logNextIndex); !success {
			continue
		}

		// Fetch log entries to send to the remote server
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

func (a *entryReplicator) stop() error {
	a.L.Lock()

	// Nothing to do if appender is already inactive
	if a.active == false {
		return nil
	}

	// Set appender to inactive and wait for processEntries to return
	a.active = false
	a.Signal()

	// Must release lock before synchronizing
	a.L.Unlock()
	select {
	case _ = <-a.done:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("entry replicator couldn't be stopped")
	}
}

func (a *entryReplicator) updateLeaderTerm(newLeaderTerm int64) {
	a.lastLeaderTerm = newLeaderTerm
	a.service.processAppendEntryEvent(newLeaderTerm, invalidLogID,
		invalidServerID)
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
