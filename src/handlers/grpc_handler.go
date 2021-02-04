package handlers

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
	r *service.AppendEntryRequest) (*service.AppendEntryReply, error) {
	// Process AppendEntry request
	currentTerm, success := h.raftService.AppendEntry(r.Entries, r.ServerTerm,
		r.ServerID, r.PrevEntryTerm, r.PrevEntryIndex, r.CommitIndex)

	// Send reply to server
	return &service.AppendEntryReply{Success: success,
		CurrentTerm: currentTerm}, nil
}

func (h *RaftRPCHandler) RequestVote(ctx context.Context,
	request *service.RequestVoteRequest) (*service.RequestVoteReply, error) {
	// Forward call arguments to RaftService
	localServerTerm, success := h.raftService.RequestVote(request.ServerTerm,
		request.ServerID, request.LastEntryTerm, request.LastEntryIndex)

	// Prepare reply and return
	return &service.RequestVoteReply{Success: success,
		ServerTerm: localServerTerm}, nil
}
