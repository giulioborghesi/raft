package domain

import (
	"testing"
	"time"
)

const (
	toMin       = 150 * time.Millisecond
	toMax       = 250 * time.Millisecond
	testRetries = 100
)

var (
	currTime time.Time
	count    = 0
)

func lastModified() time.Time {
	return currTime
}

func timeOutHandler(time.Duration) {
	count++
}

func TestRandomDuration(t *testing.T) {
	for i := 0; i < testRetries; i++ {
		// Generate random duration and ensure it falls inside [toMin, toMax)
		d := randomDuration(toMin, toMax)
		if d < toMin || d >= toMax {
			t.Fatalf("random duration outside of range [%v, %v): %v",
				toMin, toMax, d)
		}
	}
}

func TestSleepTime(t *testing.T) {
	// Initialize time and sleep for 175 millisecond
	lastModified := time.Now()
	time.Sleep(time.Duration(175 * time.Millisecond))

	// Case 1: elapsed time exceeds than timeout
	toShort := time.Duration(100 * time.Millisecond)
	sleepFor, toNew, expired := sleepTime(lastModified, toShort, toMin, toMax)

	if !expired {
		t.Fatalf("timeout was expected to expire")
	}

	if sleepFor != toNew {
		t.Fatalf("sleep time not equal to timeout: %v != %v", sleepFor, toNew)
	}

	if toNew < toMin || toNew >= toMax {
		t.Fatalf("new timeout outside of [%v, %v) range: %v", toMin, toMax,
			toNew)
	}

	// Case 2: elapsed time less than timeout
	toLong := time.Duration(200 * time.Millisecond)
	maxSleepFor := toLong - time.Since(lastModified)
	sleepFor, toNew, expired = sleepTime(lastModified, toLong, toMin, toMax)

	if expired {
		t.Fatalf("timeout was not expected to expire")
	}

	if sleepFor == toNew {
		t.Fatalf("sleep time should not be equal to timeout")
	}

	if sleepFor > maxSleepFor {
		t.Fatalf("sleep time exceeds residual time: %v,  %v", sleepFor,
			maxSleepFor)
	}

}

func TestTimeoutGenerator(t *testing.T) {
	// Create timeout generator
	handler := timeoutGenerator{active: true, tLastHandler: lastModified,
		tOutHandler: timeOutHandler, toMin: toMin, toMax: toMax}

	// Initialize current time
	currTime = time.Now()

	// Call ring alarm in a separate threat
	done := make(chan bool)
	go func() {
		handler.RingAlarm()
		done <- true
	}()

	// Make function sleep so that a timeout will occur
	time.Sleep(time.Millisecond * 100)

	// Stop handler and wait for ringAlarm to return
	handler.active = false
	<-done

	// Sleep to ensure timeout handler has been called
	time.Sleep(time.Millisecond * 100)

	// Number of times that timeout handler has been called should be one
	if count != 1 {
		t.Fatalf("wrong number of calls to timeout handler: "+
			"expected: %d, actual: %d", 1, count)
	}

}
