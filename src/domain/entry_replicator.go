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

type abstractEntrySender interface {
	sendEntries([]*service.LogEntry, int64, int64, int64) (int64, int64)
}

type entrySender struct {
	client     clients.AbstractRaftClient
	appendTerm int64
	matchIndex int64
	nextIndex  int64
}

func (s *entrySender) sendEntries(entries []*service.LogEntry,
	appendTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, int64, int64) {
	// Reset entry sender state if a new term has started
	if appendTerm > s.appendTerm {
		s.appendTerm = appendTerm
		s.matchIndex = invalidLogID
		s.nextIndex = prevEntryIndex + 1
	}

	// Send entries to remote server
	remoteServerTerm, success := s.client.AppendEntry(entries, appendTerm,
		prevEntryTerm, prevEntryIndex, commitIndex)

	// Remote server term greater than local term, return new term
	if remoteServerTerm > appendTerm {
		return remoteServerTerm, invalidLogID, invalidLogID
	}

	// Update state based on success of append entry call and return
	if success {
		s.matchIndex = prevEntryIndex + int64(len(entries))
		s.nextIndex = s.matchIndex + 1
	} else {
		s.nextIndex -= 1
	}
	return appendTerm, s.matchIndex, s.nextIndex
}

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
	active                              bool
	matchIndex, nextIndex, logNextIndex int64
	appendTerm, leaderTerm              int64
	commitIndex                         int64
	remoteServerID                      int64

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
	r.commitIndex = r.logNextIndex
	r.appendTerm = invalidTermID
	r.leaderTerm = r.appendTerm

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

func (r *entryReplicator) appendEntry(appendTerm int64,
	logNextIndex int64, commitIndex int64) {
	r.L.Lock()
	defer r.L.Unlock()

	r.appendTerm = appendTerm
	r.logNextIndex = logNextIndex
	r.commitIndex = commitIndex
	r.Signal()
}

// nextLogEntryIndex returns the append term and the starting index of the next
// log entries to send to the remote server
func (r *entryReplicator) nextLogEntryIndex() (int64, int64, bool) {
	r.L.Lock()
	defer r.L.Unlock()

	// Must wait when term has expired or there is no entry to append
	for r.appendTerm < r.leaderTerm || (r.appendTerm ==
		r.leaderTerm && (r.matchIndex+1) == r.logNextIndex) {
		r.Wait()

		// If server is no longer active, stop
		if r.active == false {
			return invalidTermID, invalidLogID, false
		}

		// Handle heartbeats
		if r.appendTerm >= r.leaderTerm {
			break
		}
	}

	return r.appendTerm, r.logNextIndex, true
}

// processEntries replicates log entries on the remote server asynchronously.
func (r *entryReplicator) processEntries() {
	for r.active {
		// Fetch append term and starting index of next log entries to send
		appendTerm, logNextIndex, active := r.nextLogEntryIndex()
		if !active {
			break
		}

		// Update the match index if needed
		if success := r.updateMatchIndex(appendTerm, logNextIndex); !success {
			continue
		}

		// Fetch the log entries to send to the remote server
		entries, leaderTerm, prevEntryTerm := r.service.entries(r.nextIndex)
		if leaderTerm > r.leaderTerm {
			r.leaderTerm = leaderTerm
			continue
		}

		// Send the log entries to the remote server
		if ok := r.sendEntries(entries, prevEntryTerm, r.matchIndex); ok {
			r.service.processAppendEntryEvent(appendTerm, r.matchIndex,
				r.remoteServerID)
		}
	}
}

func (a *entryReplicator) resetState(appendTerm int64, nextIndex int64) {
	a.leaderTerm = appendTerm
	a.matchIndex = invalidLogID
	a.nextIndex = nextIndex
}

func (a *entryReplicator) sendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	// Send entries to remote server
	remoteTerm, success :=
		a.client.AppendEntry(entries, a.leaderTerm, prevEntryTerm,
			prevEntryIndex, a.commitIndex)

	// Remote server term greater than local term, return
	if remoteTerm > a.leaderTerm {
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
	a.leaderTerm = newLeaderTerm
	a.service.processAppendEntryEvent(newLeaderTerm, invalidLogID,
		invalidServerID)
}

// updateMatchIndex finds the match index. This method performs non-trivial
// work only when the current term exceeds the last known term
func (a *entryReplicator) updateMatchIndex(appendTerm int64,
	logNextIndex int64) bool {
	// Append term did not change from last append, nothing to do
	if appendTerm == a.leaderTerm {
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
