package domain

import "time"

type AbstractRaftService interface {
	// AppendEntry is the abstract method called by the AppendEntry RPC call
	AppendEntry(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// entryInfo returns the current leader term and the term of the log entry
	// with the specified index
	entryInfo(int64) (int64, int64)

	// entries returns the log entries starting from the specified index, as
	// well as the
	entries(int64) ([]byte, int64, int64)

	// lastModified returns the time of last server modification
	lastModified() time.Time

	// makeFollower changes the server role to follower if the term argument
	// exceeds the current term
	makeFollower(int64)

	// RequestVote is the abstract method called by the RequestVote RPC call
	RequestVote(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// StartElection initiates an election based on the latest timeout
	StartElection(time.Duration)
}
