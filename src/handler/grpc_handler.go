package service

import (
	"context"

	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

type RaftRPCHandler struct {
	service.UnimplementedRaftServer
	raftService domain.AbstractRaftService
}

func (h *RaftRPCHandler) AppendEntry(ctx context.Context,
	request *service.AppendEntryRequest) (*service.AppendEntryReply, error) {
	return nil, nil
}

func (h *RaftRPCHandler) RequestVote(ctx context.Context,
	request *service.RequestVoteRequest) (*service.RequestVoteReply, error) {
	// Forward call arguments to RaftService
	localServerTerm, success := h.raftService.RequestVote(request.ServerTerm,
		request.ServerID)

	// Prepare reply and return
	return &service.RequestVoteReply{Success: success,
		ServerTerm: localServerTerm}, nil
}
