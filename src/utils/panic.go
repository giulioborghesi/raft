package utils

import (
	"fmt"
	"testing"
)

const (
	panicErrorMsg = "%s expected to panic"
)

func AssertPanic(t *testing.T, s string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(fmt.Sprintf(panicErrorMsg, s))
		}
	}()

	f()
}
