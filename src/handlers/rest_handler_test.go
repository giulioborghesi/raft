package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/domain"
	"github.com/gorilla/mux"
)

// Create an instance of a router for a RAFT server
func makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.StrictSlash(true)

	server := &RaftRestServer{s: &restRaftService{}}
	router.HandleFunc("/command/", server.ApplyCommandAsync).Methods("POST")
	router.HandleFunc("/command/{id:[0-9]+}", server.CommandStatus).Methods("GET")
	return router
}

func TestApplyCommandAsync(t *testing.T) {
	// Create router
	router := makeRouter()

	// Prepare request
	var jsonStr = []byte(`{"payload":"lorem ipsum"}`)
	req, err := http.NewRequest("POST", "/command/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Serve request
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Validate response code
	if rr.Code != http.StatusOK {
		t.Fatalf(unexpectedResponseStatusErrFmt, http.StatusOK, rr.Code)
	}
}

func TestCommandStatusAsync(t *testing.T) {
	// Create router
	router := makeRouter()

	// Prepare request
	url := fmt.Sprintf("/command/%s", magicEntryID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Serve request
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Validate response code
	if rr.Code != http.StatusOK {
		t.Fatalf(unexpectedResponseStatusErrFmt, http.StatusOK, rr.Code)
	}

	// Validate expected result
	decoder := json.NewDecoder(rr.Body)
	serverResponse := &commandStatusResponse{}
	err = decoder.Decode(serverResponse)
	if err != nil {
		t.Fatalf("%v", err)
	}

	entryStatus := serverResponse.EntryStatus
	expectedEntryStatus := domain.EntryStatusToString(magicEntryStatus)

	if entryStatus != expectedEntryStatus {
		t.Fatalf(unexpectedEntryStatusErrFmt, expectedEntryStatus, entryStatus)
	}
}
