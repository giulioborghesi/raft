package domain

import (
	"math/rand"
	"time"
)

// timeoutHandler defines an object that handles a timeout
type timeoutHandler struct {
	active       bool
	tLastHandler func() time.Time
	tOutHandler  func(time.Duration)
	to           time.Duration
	toMin        time.Duration
	toMax        time.Duration
}

func init() {
	// Initialize random number generator
	rand.Seed(time.Now().Unix())
}

// randomDuration returns a random duration in the [0, maxDuration-minDuration)
// range
func randomDuration(minDuration, maxDuration time.Duration) time.Duration {
	delta := time.Duration(rand.Intn(int(maxDuration) - int(minDuration)))
	return minDuration + delta
}

// sleepTime returns the sleep interval and the timeout value
func sleepTime(t time.Time, to, toMin, toMax time.Duration) (time.Duration,
	time.Duration, bool) {
	// Timeout must be less than maximum timeout
	if to >= toMax {
		panic("timeout cannot exceed maximum timeout")
	}

	// Compute time elapsed since last update
	d := time.Since(t)

	// Time elapsed exceeds timeout: compute new timeout and start election
	if d >= to {
		newTo := randomDuration(toMin, toMax)
		return newTo, newTo, true
	}

	// Time elapsed smaller than timeout: compute residual sleep time
	return to - d, to, false
}

// ringAlarm starts a new election should the election timeout have expired.
// The election timeout is computed as the sum of a constant contribution (
// average election timeout) and a variable one (election timeout rms). A new
// election is started iff the server is not the cluster leader and no action
// has occurred since the election timeout had been set
func (h *timeoutHandler) ringAlarm() {
	sleepFor, to, expired := sleepTime(time.Now(), h.to, h.toMin, h.toMax)
	for h.active {
		// Sleep for the specified time and update timeout
		time.Sleep(sleepFor)
		h.to = to

		// Compute new sleep time and timeout
		lastModified := h.tLastHandler()
		sleepFor, to, expired = sleepTime(lastModified, h.to, h.toMin, h.toMax)

		// Timeout has expired, invoke timeout handler
		if expired {
			go func(to time.Duration) {
				h.tOutHandler(to)
			}(h.to)
		}
	}
}
