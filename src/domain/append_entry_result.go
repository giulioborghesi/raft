package domain

type appendEntryResult struct {
	index   int64
	success chan bool
}
