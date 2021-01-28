package datasources

import (
	"fmt"
	"testing"
	"time"
)

const (
	errorMsg = "minimum number of retries not reached yet"
)

// mockServerStateDao is a mock DAO that succeeds only when the number of
// retries exceeds a minimum number of retries
type mockServerStateDao struct {
	retries, minRetries int
}

func (d *mockServerStateDao) CurrentTerm() int64 {
	return 0
}

func (d *mockServerStateDao) UpdateTerm(val int64) error {
	return d.UpdateTermVotedFor(val, -1)
}

func (d *mockServerStateDao) UpdateTermVotedFor(int64, int64) error {
	if d.retries < d.minRetries {
		d.retries++
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

func (d *mockServerStateDao) UpdateVotedFor(int64) error {
	if d.retries < d.minRetries {
		d.retries++
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

func (d *mockServerStateDao) VotedFor() (int64, int64) {
	return 0, 0
}

// createMockServerStateDao creates a mock server state DAO object
func createMockServerStateDao(minRetries int) *mockServerStateDao {
	return &mockServerStateDao{minRetries: minRetries}
}

func TestServerStateDaoWithRetryFail(t *testing.T) {
	// Create DAO object
	minRetries := maxRetries + 1
	dao := ServerStateDaoWithRetry{
		ServerStateDao: createMockServerStateDao(minRetries)}

	// Record time before function call
	timeStart := time.Now()

	// Call UpdateTermVotedFor, which should fail
	err := dao.UpdateTermVotedFor(0, 0)
	if err == nil {
		t.Fatal("expected UpdateTermVotedFor to fail")
	}

	if err.Error() != errorMsg {
		t.Fatal("invalid error message returned by UpdateTermVotedFor: "+
			"expected: %s, actual: %s", err.Error(), errorMsg)
	}

	// Compute time elapsed since function call
	timeElapsed := time.Since(timeStart)

	// Time elapsed exceed maximum allowed by exponential backoff algorithm
	maxTimeElapsed := time.Duration(time.Millisecond * 1 << (maxRetries + 1))
	if timeElapsed > maxTimeElapsed {
		t.Fatalf("excessive elapsed time: "+
			"expected less than: %v, actual: %v", maxTimeElapsed, timeElapsed)
	}
}
