package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/giulioborghesi/raft-implementation/src/utils"
	"github.com/gorilla/mux"
)

// extractPayloadFromJSON extracts the command payload from a REST request
func extractPayloadFromJSON(req *http.Request) (string, error) {
	decoder := json.NewDecoder(req.Body)
	clientRequest := &applyCommandRequest{}
	err := decoder.Decode(clientRequest)
	if err != nil {
		return "", err
	}
	return clientRequest.Payload, nil
}

type applyCommandRequest struct {
	Payload string
}

// applyCommandResponse is an helper structure used to construct the response
// to an ApplyCommandAsync request
type applyCommandResponse struct {
	EntryID  string `json:"entry_id"`
	LeaderID int64  `json:"leader_id"`
}

// commandStatusResponse is an helper structure used to construct the response
// to a CommandStatusRequest request
type commandStatusResponse struct {
	EntryStatus string `json:"entry_status"`
	LeaderID    int64  `json:"leader_id"`
}

// RaftRestServer is a REST server used to respond to REST requests received
// from clients of the Raft cluster
type RaftRestServer struct {
	S domain.AbstractRaftService
}

func (h *RaftRestServer) ApplyCommandAsync(w http.ResponseWriter,
	req *http.Request) {
	// Extract payload from request and apply command to Raft service
	payload, err := extractPayloadFromJSON(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entryID, leaderID, _ := h.S.ApplyCommandAsync(payload)

	// Send response to client
	response := &applyCommandResponse{EntryID: entryID, LeaderID: leaderID}
	utils.RenderJSON(w, response)
}

func (h *RaftRestServer) CommandStatus(w http.ResponseWriter,
	req *http.Request) {
	// Extract payload from request and apply command
	entryID := mux.Vars(req)["id"]
	entryStatus, leaderID, _ := h.S.CommandStatus(entryID)

	// Send response to client
	response := &commandStatusResponse{
		EntryStatus: domain.EntryStatusToString(entryStatus),
		LeaderID:    leaderID}
	utils.RenderJSON(w, response)
}
