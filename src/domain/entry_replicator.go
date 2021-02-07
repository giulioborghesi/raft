package domain

import (
	"fmt"
	"sync"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	entrySenderInconsistencyErrA = "invalid prev entry index"
	entrySenderInconsistencyErrB = "next index not greater than match index"
)

var (
	entryAppenderTimeout time.Duration = 50 * time.Millisecond
)

// abstractEntrySender provides an API to send entries to a remote server and
// to query the value of the index of the next entry in the remote server log
type abstractEntrySender interface {
	// sendEntries sends the input entries to the remote log. It returns the
	// remote server term as well as a boolean value indicating whether the
	// operation was successful or not
	sendEntries([]*service.LogEntry, int64, int64, int64, int64) (int64, bool)

	// nextEntryIndex returns the last known index of the next entry in the
	// remote server log. This value may not be up-to-date
	nextEntryIndex() int64
}

// entrySender implements the abstractEntrySender interface. Concrete instances
// of this type keep track of the match index and of the next index in the
// remote server log
type entrySender struct {
	client     clients.AbstractRaftClient
	appendTerm int64
	matchIndex int64
	nextIndex  int64
}

// makeEntrySender creates an instance of entrySender that communicates with
// another server by mean of an abstract Raft client
func makeEntrySender(appendTerm int64,
	client clients.AbstractRaftClient) *entrySender {
	return &entrySender{client: client, appendTerm: appendTerm,
		matchIndex: invalidLogEntryIndex}
}

func (s *entrySender) nextEntryIndex() int64 {
	return s.nextIndex
}

func (s *entrySender) sendEntries(entries []*service.LogEntry,
	appendTerm int64, prevEntryTerm int64, prevEntryIndex int64,
	commitIndex int64) (int64, bool) {
	// Reset entry sender state if a new term has started
	if appendTerm > s.appendTerm {
		s.appendTerm = appendTerm
		s.matchIndex = invalidLogEntryIndex
		s.nextIndex = prevEntryIndex + 1
	}

	if s.nextIndex != prevEntryIndex+1 {
		panic(entrySenderInconsistencyErrA)
	}

	// Send entries to remote server
	remoteServerTerm, success := s.client.AppendEntry(entries, appendTerm,
		prevEntryTerm, prevEntryIndex, commitIndex)

	// Remote server term greater than local term, return new term
	if remoteServerTerm > appendTerm {
		return remoteServerTerm, false
	}

	// Update state based on success of append entry call and return
	if success {
		s.matchIndex = prevEntryIndex + int64(len(entries))
		s.nextIndex = s.matchIndex + 1
	} else {
		s.nextIndex -= 1
	}

	// Ensure next index value is always found
	if s.nextIndex <= s.matchIndex {
		panic(entrySenderInconsistencyErrB)
	}
	return appendTerm, success
}

// abstractEntryReplicator defines an interface for replicating local log
// entries on a remote server
type abstractEntryReplicator interface {
	// appendEntry replicates the log entry with the specified index and log
	// entry term in the remote server. The replication may be asynchronous
	appendEntry(int64, int64, int64)
}

// entryReplicator implements the abstractEntryReplicator interfaces.
// Replication is conducted asynchronously to improve performance
type entryReplicator struct {
	active, wait           bool
	nextIndex, commitIndex int64
	appendTerm, leaderTerm int64
	remoteServerID         int64

	done    chan bool
	sender  abstractEntrySender
	service AbstractRaftService
	*sync.Cond
}

// makeEntryReplicator creates an instance of entryReplicator. The object will
// be initialized with an invalid term ID to ensure that log entries starts to
// be replicated only after receiving a signal from the current leader
func newEntryReplicator(remoteServerID int64, c clients.AbstractRaftClient,
	rs AbstractRaftService) *entryReplicator {
	// Create entry replicator
	es := makeEntrySender(invalidTerm, c)
	r := &entryReplicator{active: false, wait: true, sender: es, service: rs,
		remoteServerID: remoteServerID}

	// Initialize entry replicator state
	r.nextIndex = 0
	r.commitIndex = r.nextIndex
	r.appendTerm = invalidTerm
	r.leaderTerm = r.appendTerm

	// Initialize entry replicator condition variable and return
	m := sync.Mutex{}
	r.Cond = sync.NewCond(&m)
	return r
}

// NewEntryReplicator will create and start a new instance of entry replicator
func makeEntryReplicator(remoteServerID int64, c clients.AbstractRaftClient,
	rs AbstractRaftService) *entryReplicator {
	// Create entry replicator
	r := newEntryReplicator(remoteServerID, c, rs)

	// Activate entry replicator and return
	r.start()
	return r
}

func (r *entryReplicator) appendEntry(appendTerm int64,
	logNextIndex int64, commitIndex int64) {
	r.L.Lock()
	defer r.L.Unlock()

	// Update state
	r.appendTerm = appendTerm
	r.nextIndex = logNextIndex
	r.commitIndex = commitIndex
	r.wait = false

	// Send signal
	r.Signal()
}

// nextLogEntryIndex returns the append term and the starting index of the next
// log entries to send to the remote server
func (r *entryReplicator) nextEntryIndex() (int64, int64, bool) {
	r.L.Lock()
	defer r.L.Unlock()

	// Must wait when term has expired or there is no entry to append
	for r.active && ((r.appendTerm < r.leaderTerm) || ((r.appendTerm ==
		r.leaderTerm) && (r.wait == true))) {
		r.Wait()
	}
	r.wait = true
	return r.appendTerm, r.nextIndex, r.active
}

// processEntries replicates log entries on the remote server asynchronously.
func (r *entryReplicator) processEntries() {
	for r.active {
		// Fetch append term and starting index of next log entries to send
		appendTerm, nextIndex, active := r.nextEntryIndex()
		if !active {
			continue
		}

		// Update the match index if needed
		remoteServerTerm, nextRemoteIndex :=
			r.updateMatchIndex(appendTerm, nextIndex)
		if remoteServerTerm > appendTerm {
			r.updateLeaderTerm(remoteServerTerm)
			continue
		}

		// Fetch the log entries to send to the remote server
		entries, localTerm, prevEntryTerm := r.service.entries(nextRemoteIndex)
		if localTerm > appendTerm {
			r.leaderTerm = localTerm
			continue
		}

		// Send the log entries to the remote server
		remoteServerTerm, success := r.sender.sendEntries(entries,
			appendTerm, prevEntryTerm, nextRemoteIndex-1, r.commitIndex)
		if remoteServerTerm != appendTerm {
			r.updateLeaderTerm(remoteServerTerm)
			continue
		}

		// AppendEntry RPC was successful, notify Raft service
		if success {
			newMatchIndex := r.sender.nextEntryIndex() - 1
			r.service.processAppendEntryEvent(appendTerm, newMatchIndex,
				r.remoteServerID)
		}
	}
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
	a.service.processAppendEntryEvent(newLeaderTerm, invalidLogEntryIndex,
		invalidServerID)
}

// updateMatchIndex finds the match index. This method performs non-trivial
// work only when the current term exceeds the last known term
func (r *entryReplicator) updateMatchIndex(appendTerm int64,
	nextIndex int64) (int64, int64) {
	// Append term did not change from last append, nothing to do
	if appendTerm == r.leaderTerm {
		return appendTerm, r.sender.nextEntryIndex()
	}

	// A new term has started, reset appender state
	r.leaderTerm = appendTerm

	// Initialize loop variables
	done := false
	prevEntryIndex := nextIndex - 1

	// Loop until match index has been updated
	for !done {
		leaderTerm, prevEntryTerm := r.service.entryInfo(prevEntryIndex)
		if leaderTerm > appendTerm {
			return leaderTerm, invalidLogEntryIndex
		}

		// Send empty entry to remote server
		var remoteServerTerm int64
		remoteServerTerm, done = r.sender.sendEntries(nil,
			appendTerm, prevEntryTerm, prevEntryIndex, r.commitIndex)

		// Return if new term is detected
		if remoteServerTerm > appendTerm {
			return remoteServerTerm, invalidLogEntryIndex
		}

		// Update previous entry index
		prevEntryIndex -= 1
	}
	return appendTerm, r.sender.nextEntryIndex()
}
