package handlers

import (
	"github.com/giulioborghesi/raft-implementation/src/domain"
)

const (
	magicEntryID     = "3"
	magicEntryStatus = 4
)

// restRaftService is an implementation of domain.AbstractRaftService to be
// used for testing the application REST handlers
type restRaftService struct {
	domain.UnimplementedRaftService
}

func (s *restRaftService) ApplyCommandAsync(string) (string, int64, error) {
	return "", -1, nil
}

func (s *restRaftService) CommandStatus(entryID string) (domain.LogEntryStatus,
	int64, error) {
	if entryID == magicEntryID {
		return magicEntryStatus, 0, nil
	}
	return -1, -1, nil
}
