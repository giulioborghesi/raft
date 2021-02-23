package app

import (
	"fmt"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/giulioborghesi/raft-implmenetation/src/handlers"
)

func MakeRaftApp(stateFilePath string, logFilePath string,
	serverID int64, serversInfo []*clients.ServerInfo) *raftApp {
	// Create server state dao
	serverStateDao := datasources.MakePersistentServerStateDao(
		fmt.Sprintf("%s_A.log", stateFilePath),
		fmt.Sprintf("%s_B.log", stateFilePath),
	)

	// Create log object
	logDao := datasources.MakePersistentLogDao(logFilePath)
	raftLog := domain.MakeRaftLog(logDao)

	// Create server state and Raft service
	serverState := domain.MakeServerState(serverStateDao, raftLog, serverID)
	service := domain.MakeRaftService(serverState, serversInfo)

	// Create and return Raft application
	return &raftApp{restServer: &handlers.RaftRestServer{s: service}}
}

type raftApp struct {
	restServer handlers.RaftRestServer
}
