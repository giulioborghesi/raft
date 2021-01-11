package server_state_dao

import "time"

const (
	maxRetries = 10
)

type ServerStateDaoWithRetry struct {
	ServerStateDao
}

func (d *ServerStateDaoWithRetry) UpdateTermVotedFor(newTerm int64,
	newVotedFor int64) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		// Try to run operation
		err = d.ServerStateDao.UpdateTermVotedFor(newTerm, newVotedFor)
		if err == nil {
			break
		}

		// Operation failed, retry after exponential backoff
		fact := (1 << i)
		time.Sleep(time.Millisecond * time.Duration(fact))
	}
	return err
}

func (d *ServerStateDaoWithRetry) UpdateVotedFor(newVotedFor int64) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		// Try to run operation
		err = d.ServerStateDao.UpdateVotedFor(newVotedFor)
		if err == nil {
			break
		}

		// Operation failed, retry after exponential backoff
		fact := (1 << i)
		time.Sleep(time.Millisecond * time.Duration(fact))
	}
	return err
}
