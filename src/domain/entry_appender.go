package domain

import (
	"sync"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
)

var (
	entryAppenderTimeout time.Duration = 50 * time.Millisecond
)

type abstractEntryAppender interface {
	appendEntry(int64, int64)
	start()
	stop()
}

type entryAppender struct {
	active                                     bool
	matchIndex, nextIndex, logNextIndex        int64
	appendTerm, lastAppendTerm, lastLeaderTerm int64

	done    chan bool
	client  clients.AbstractRaftClient
	service AbstractRaftService
	sync.Cond
}

func (a *entryAppender) appendEntry(appendTerm int64,
	logNextIndex int64) {
	a.L.Lock()
	defer a.L.Unlock()

	a.appendTerm = appendTerm
	a.logNextIndex = logNextIndex
	a.Signal()
}

func (a *entryAppender) processEntries() {
	for a.active {
		a.L.Lock()

		// Must wait when term has expired or there is no entry to append
		for a.appendTerm < a.lastLeaderTerm || (a.appendTerm ==
			a.lastAppendTerm && (a.matchIndex+1) == a.logNextIndex) {
			a.Wait()
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
		a.sendEntries(a.matchIndex, prevEntryTerm, entries)
	}
}

// updateMatchIndex finds the match index. This method performs non-trivial
// work only when the current term exceeds the last known term
func (a *entryAppender) updateMatchIndex(appendTerm int64,
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
		if ok := a.sendEntries(prevEntryTerm, prevEntryIndex, nil); !ok {
			return false
		}

		// Update loop condition
		prevEntryIndex = a.nextIndex - 1
		done = a.nextIndex == a.matchIndex+1
	}
	return true
}

func (a *entryAppender) resetState(appendTerm int64, nextIndex int64) {
	a.lastAppendTerm = appendTerm
	a.lastLeaderTerm = a.lastAppendTerm
	a.matchIndex = 0
	a.nextIndex = nextIndex
}

func (a *entryAppender) sendEntries(prevEntryTerm int64,
	prevEntryIndex int64, entries []byte) bool {
	// Send entries to remote server
	remoteTerm, success := a.client.AppendEntry(a.lastAppendTerm,
		prevEntryTerm, prevEntryIndex)

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

func (a *entryAppender) start() {
	a.L.Lock()
	defer a.L.Unlock()

	// Nothing to do if appender is already active
	if a.active {
		return
	}

	// Create channel
	done := make(chan bool)

	// Launch process entries asynchronously
	a.active = true
	a.done = done
	go func() {
		a.processEntries()
		done <- true
	}()
}

func (a *entryAppender) stop() {
	a.L.Lock()
	defer a.L.Unlock()

	// Nothing to do if appender is already inactive
	if a.active == false {
		return
	}

	// Set appender to inactive and wait for processEntries to return
	a.active = false
	<-a.done
}

func (a *entryAppender) updateLeaderTerm(newLeaderTerm int64) {
	a.lastLeaderTerm = newLeaderTerm
	a.service.makeFollower(newLeaderTerm)
}
