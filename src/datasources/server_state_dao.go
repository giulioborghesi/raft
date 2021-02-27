package datasources

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// initializeLogStateFromEmptyFile creates a log state file and
// initializes the log state to its default value
func initializeLogStateFromEmptyFile(path string) (*os.File, int64, int64,
	error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, permissions)
	if err != nil {
		panic(err)
	}
	return file, 0, -1, err
}

// initializeLogStateFromFile initializes the log state from an input file
func initializeLogStateFromFile(path string) (*os.File, int64, int64, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return initializeLogStateFromEmptyFile(path)
	}

	file, err := os.OpenFile(path, os.O_RDWR, permissions)
	if err != nil {
		panic(err)
	}

	currentTerm, votedFor, err := readLogStateFromFile(file)
	return file, currentTerm, votedFor, err
}

// readLogStateFromFile reads the log state from file and returns the current
// term, voted for and error message if successful, an error otherwise
func readLogStateFromFile(file *os.File) (int64, int64, error) {
	r := bufio.NewReader(file)
	var currentTerm int64
	if err := binary.Read(r, binary.LittleEndian, &currentTerm); err != nil {
		return 0, -1, err
	}

	var votedFor int64
	if err := binary.Read(r, binary.LittleEndian, &votedFor); err != nil {
		return 0, -1, err
	}

	var checksum [28]byte
	if err := binary.Read(r, binary.LittleEndian, &checksum); err != nil {
		return 0, -1, err
	}

	actualCheckSum := utils.Int64PairCheckSum(currentTerm, votedFor)
	if actualCheckSum != checksum {
		return 0, -1, fmt.Errorf(utils.ChecksumFailedErrFmt,
			hex.EncodeToString(checksum[:]),
			hex.EncodeToString(actualCheckSum[:]))
	}
	return currentTerm, votedFor, nil
}

// writeLogStateToFile writes to log state to file
func writeLogStateToFile(file *os.File, currentTerm int64,
	votedFor int64) error {
	// Set file offset to beginning
	_, err := file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Write server term and voted for to file
	w := bufio.NewWriter(file)
	if err := binary.Write(w, binary.LittleEndian, currentTerm); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, votedFor); err != nil {
		return err
	}

	// Write checksum to file
	checksum := utils.Int64PairCheckSum(currentTerm, votedFor)
	if err := binary.Write(w, binary.LittleEndian, checksum); err != nil {
		return err
	}

	// Commit changes to stable storage and return
	return w.Flush()
}

// ServerStateDao defines the interface of a server state DAO object, which is
// used to access and persist to disk the term number and the ID of the server
// that has received the server's vote in the current term
type ServerStateDao interface {
	// CurrentTerm returns the ID of the current term
	CurrentTerm() int64

	// UpdateTerm updates the current term ID and persists its value to disk
	UpdateTerm(int64) error

	// UpdateTermVotedFor updates the current term ID and the ID of the server
	// that has received this server's vote in the current term
	UpdateTermVotedFor(int64, int64) error

	// UpdateVotedFor updates the ID of the server that has received this
	// server's vote in the current term
	UpdateVotedFor(int64) error

	// VotedFor returns the current term ID and the ID of the server that has
	// received this server's vote in the current term
	VotedFor() (int64, int64)
}

// MakeInMemoryServerStateDao creates an instance of inMemoryServerStateDao for
// use in unit tests outside of this package
func MakeInMemoryServerStateDao(currentTerm,
	votedFor int64) *inMemoryServerStateDao {
	return &inMemoryServerStateDao{currentTerm: currentTerm, votedFor: votedFor}
}

// inMemoryServerStateDao implements an in-memory version of the ServerStateDao
// interface. This implementation is useful for testing code
type inMemoryServerStateDao struct {
	currentTerm int64
	votedFor    int64
}

func (s *inMemoryServerStateDao) CurrentTerm() int64 {
	return s.currentTerm
}

func (s *inMemoryServerStateDao) UpdateTerm(serverTerm int64) error {
	return s.UpdateTermVotedFor(serverTerm, -1)
}

func (s *inMemoryServerStateDao) UpdateTermVotedFor(serverTerm int64,
	serverID int64) error {
	s.currentTerm = serverTerm
	s.votedFor = serverID
	return nil
}

func (s *inMemoryServerStateDao) UpdateVotedFor(serverID int64) error {
	s.votedFor = serverID
	return nil
}

func (s *inMemoryServerStateDao) VotedFor() (int64, int64) {
	return s.currentTerm, s.votedFor
}

func MakePersistentServerStateDao(pathA string, pathB string) ServerStateDao {
	// Initialize log state from file
	fileA, currentTermA, votedForA, errorA := initializeLogStateFromFile(pathA)
	fileB, currentTermB, votedForB, errorB := initializeLogStateFromFile(pathB)

	// At least one file must not be corrupted
	if errorA != nil && errorB != nil {
		panic(fmt.Sprintf(corruptedLogFilesErrFmt, pathA, pathB))
	}

	// Ensure most current log state is used
	var lastA bool = true
	currentTerm, votedFor := currentTermA, votedForA
	if currentTermA < currentTermB ||
		(currentTermA == currentTermB && votedForA == -1) {
		currentTerm, votedFor, lastA = currentTermB, votedForB, false
	}

	// Create and return pointer to persistent state dao object
	d := inMemoryServerStateDao{currentTerm: currentTerm, votedFor: votedFor}
	return &persistentServerStateDao{fileA: fileA, fileB: fileB, lastA: lastA,
		inMemoryServerStateDao: d}
}

// persistentServerStateDao persists the term and voted for information to
// stable storage. It wraps inMemoryServerStateDao to guarantee fast access to
// the current term and voted for, removing the need of going to disk
type persistentServerStateDao struct {
	fileA, fileB *os.File
	lastA        bool
	inMemoryServerStateDao
}

func (s *persistentServerStateDao) UpdateTerm(serverTerm int64) error {
	return s.UpdateTermVotedFor(serverTerm, -1)
}

func (s *persistentServerStateDao) UpdateTermVotedFor(serverTerm int64,
	votedFor int64) error {
	// Get a pointer to the file that was not used during last write cycle
	file := s.fileA
	if s.lastA == true {
		file = s.fileB
	}
	s.lastA = !s.lastA

	// Write log state to file
	if err := writeLogStateToFile(file, serverTerm, votedFor); err != nil {
		return err
	}

	// Modify in-memory server state and return
	s.inMemoryServerStateDao.UpdateTerm(serverTerm)
	return nil
}

func (s *persistentServerStateDao) UpdateVotedFor(votedFor int64) error {
	return s.UpdateTermVotedFor(s.currentTerm, votedFor)
}
