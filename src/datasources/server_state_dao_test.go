package datasources

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testFileA = "/root/testFileA.bin"
	testFileB = "/root/testFileB.bin"
)

func TestInitializeLogStateFromFile(t *testing.T) {
	// Remove test file A if needed
	if _, err := os.Stat(testFileA); !os.IsNotExist(err) {
		os.Remove(testFileA)
	}

	// Initialize log state file
	file, currentTerm, votedFor, err :=
		initializeLogStateFromFile(testFileA)

	utils.ValidateResult(t, utils.InvalidTermErrFmt, 0, currentTerm)
	utils.ValidateResult(t, utils.InvalidVotedForErrFmt, -1, votedFor)

	if err != nil {
		t.Fatalf(utils.UnexpectedErrFmt, "none", err)
	}

	if _, err := os.Stat(testFileA); os.IsNotExist(err) {
		t.Fatalf(utils.FileNotFoundErrFmt, testFileA)
	}

	// Close log state file
	file.Close()

	// Initialize log state file from existing file
	file, currentTerm, votedFor, err =
		initializeLogStateFromFile(testFileA)

	utils.ValidateResult(t, utils.InvalidTermErrFmt, 0, currentTerm)
	utils.ValidateResult(t, utils.InvalidVotedForErrFmt, -1, votedFor)

	if err != io.EOF {
		t.Fatalf(utils.UnexpectedErrFmt, "EOF", err)
	}

	// Close and delete log state file
	file.Close()
	os.Remove(testFileA)
}

func TestReadWriteLogState(t *testing.T) {
	// Remove test file A if needed
	if _, err := os.Stat(testFileA); !os.IsNotExist(err) {
		os.Remove(testFileA)
	}

	// Open log state file
	file, err := os.OpenFile(testFileA, os.O_RDWR|os.O_CREATE, permissions)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	// Write log state to file
	err = writeLogStateToFile(file, 15, 2)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	// Read log state from file
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	actualTerm, actualVotedFor, err := readLogStateFromFile(file)
	if err != nil {
		t.Fatalf(fmt.Sprintf("%v", err))
	}

	utils.ValidateResult(t, utils.InvalidTermErrFmt, 15, actualTerm)
	utils.ValidateResult(t, utils.InvalidVotedForErrFmt, 2, actualVotedFor)

	// Close and delete log state file
	file.Close()
	os.Remove(testFileA)
}

func TestInMemoryServerStateDao(t *testing.T) {
	s := &inMemoryServerStateDao{}

	// Fetch current term
	currentTerm := s.CurrentTerm()
	utils.ValidateResult(t, utils.InvalidTermErrFmt, 0, currentTerm)

	// Update voted for
	err := s.UpdateVotedFor(5)
	if err != nil {
		t.Fatalf(utils.UnexpectedErrFmt, "none", err)
	}

	// Check current term and voted for
	currentTerm, votedFor := s.VotedFor()
	utils.ValidateResult(t, utils.InvalidTermErrFmt, 0, currentTerm)
	utils.ValidateResult(t, utils.InvalidVotedForErrFmt, 5, votedFor)

	// Update term
	err = s.UpdateTerm(8)
	if err != nil {
		t.Fatalf(utils.UnexpectedErrFmt, "none", err)
	}

	// Check current term and voted for
	currentTerm, votedFor = s.VotedFor()
	utils.ValidateResult(t, utils.InvalidTermErrFmt, 8, currentTerm)
	utils.ValidateResult(t, utils.InvalidVotedForErrFmt, -1, votedFor)
}
