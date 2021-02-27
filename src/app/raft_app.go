package app

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/giulioborghesi/raft-implementation/src/handlers"
	"github.com/giulioborghesi/raft-implementation/src/service"
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

	// Create timeout generators
	eTmin := time.Duration(250 * time.Millisecond)
	eTmax := time.Duration(350 * time.Millisecond)
	electionTG := domain.MakeTimeoutGenerator(service.LastModified,
		service.StartElection, eTmin, eTmax)

	hTmin := time.Duration(100 * time.Millisecond)
	hTmax := time.Duration(150 * time.Millisecond)
	heartbeatTG := domain.MakeTimeoutGenerator(service.LastModified,
		service.SendHeartbeat, hTmin, hTmax)

	// Create and return Raft application
	return &raftApp{
		electionTG: electionTG, heartbeatTG: heartbeatTG,
		rpcServer:  handlers.RaftRPCService{S: service},
		restServer: handlers.RaftRestServer{S: service}}
}

type raftApp struct {
	electionTG  domain.AbstractTimeoutGenerator
	heartbeatTG domain.AbstractTimeoutGenerator
	rpcServer   handlers.RaftRPCService
	restServer  handlers.RaftRestServer
}

func (a *raftApp) ListenAndServe(restPort, rpcPort string) error {
	// Create REST listener
	restLis, err := net.Listen("tcp", restPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create RPC listener
	rpcLis, err := net.Listen("tcp", rpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cErr := make(chan error, 2)

	// Start election timeout generator
	go a.electionTG.RingAlarm()

	// Start REST server
	go func() {
		r := mux.NewRouter()
		r.StrictSlash(true)
		r.HandleFunc("/command/", a.restServer.ApplyCommandAsync).Methods("POST")
		r.HandleFunc("/command/{id}", a.restServer.CommandStatus).Methods("GET")
		cErr <- http.Serve(restLis, r)
	}()

	// Start RPC server
	go func() {
		s := grpc.NewServer()
		service.RegisterRaftServer(s, &a.rpcServer)
		cErr <- s.Serve(rpcLis)
	}()

	// Fail service should an error occur
	for err := range cErr {
		panic(err)
	}

	return nil
}
