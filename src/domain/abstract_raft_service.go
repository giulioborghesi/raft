package domain

import "time"

type AbstractRaftService interface {
	// AppendEntry implements the AppendEntry RPC call
	AppendEntry(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// lastModified returns the time of last modification
	lastModified() time.Time

	// RequestVote implements the RequestVote RPC call
	RequestVote(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// StartElection attempts to start an election
	StartElection(time.Duration)
}
