package domain

import (
	"math/rand"
	"time"
)

// AbstractTimeoutGenerator defines the interface of a generic timeout
// generator
type AbstractTimeoutGenerator interface {
	RingAlarm()
}

// MakeTimeoutGenerator creates and initializes an instance of a timeout
// generator
func MakeTimeoutGenerator(lh func() time.Time, th func(time.Duration),
	toMin time.Duration, toMax time.Duration) *timeoutGenerator {
	return &timeoutGenerator{active: true, tLastHandler: lh,
		tOutHandler: th, toMin: toMin, toMax: toMax}
}

// timeoutGenerator defines an object that handles a timeout
type timeoutGenerator struct {
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

// randomDuration returns a random duration in the [minDuration, maxDuration)
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

// ringAlarm starts the timeout handler and calls a timeout handler function
// whenever the timeout has expired. A new timeout is generated only after
// the current timeout has expired. Expiration occurs when the time elapsed
// since the last change made to the server exceeds the timeout
func (h *timeoutGenerator) RingAlarm() {
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
