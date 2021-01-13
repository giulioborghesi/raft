package domain

import "time"

type AbstractRaftService interface {
	// AppendEntry is the abstract method called by the AppendEntry RPC call
	AppendEntry(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// lastModified returns the time of last server modification
	lastModified() time.Time

	// RequestVote is the abstract method called by the RequestVote RPC call
	RequestVote(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// StartElection initiates an election based on the latest timeout
	StartElection(time.Duration)
}
