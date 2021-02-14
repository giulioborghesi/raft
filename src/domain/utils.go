package domain

import (
	"fmt"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

// EntryStatusToString converts an entry status value to a string literal
func EntryStatusToString(s LogEntryStatus) string {
	switch s {
	case appended:
		return "appended"
	case lost:
		return "lost"
	case invalid:
		return "invalid"
	case unknown:
		return "unknown"
	default:
		panic(fmt.Sprintf(invalidEntryStatus, s))
	}
}

// entryTermsToLogEntry converts a slice of entry terms into a slice of empty
// log entries
func entryTermsToLogEntry(entryTerms []int64) []*service.LogEntry {
	entries := make([]*service.LogEntry, 0, len(entryTerms))
	for _, entryTerm := range entryTerms {
		entries = append(entries, &service.LogEntry{EntryTerm: entryTerm})
	}
	return entries
}

// roleName returns a human-readable string describing the server role
func roleName(role int) string {
	switch role {
	case candidate:
		return "candidate"
	case follower:
		return "follower"
	case leader:
		return "leader"
	default:
		panic("unknown role")
	}
}
