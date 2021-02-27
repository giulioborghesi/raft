package app

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implmenetation/src/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
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
	return &raftApp{
		rpcServer:  handlers.RaftRPCService{raftService: service},
		restServer: handlers.RaftRestServer{s: service}}
}

type raftApp struct {
	rpcServer  handlers.RaftRPCService
	restServer handlers.RaftRestServer
}

func (a *raftApp) ListenAndServe(restPort, rpcPort string) error {
	// Create REST listener
	restLis, err := net.Listen("tcp", restPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create RPC listener
	rpcLis, err := net.Listen("tpc", rpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cErr := make(chan error, 2)

	// Start REST server
	go func() {
		r := mux.NewRouter()
		r.StrictSlash(true)
		r.HandleFunc("/command/", a.restServer.ApplyCommandAsync).Methods("POST")
		r.HandleFunc("/command/{id:[0-9]+}", a.restServer.CommandStatus).Methods("GET")
		cErr <- http.Serve(restLis, r)
	}()

	// Start RPC server
	go func() {
		s := grpc.NewServer()
		service.RegisterRaftServer(s, a.rpcServer)
		cErr <- s.Serve(rpcLis)
	}()

	// Fail service should an error occur
	for err := range cErr {
		panic(err)
	}

	return nil
}
