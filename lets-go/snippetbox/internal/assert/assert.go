package assert

import (
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	// This indicates to the Go test runner that this function is a test helper
	// This means that when t.Errorf() is called from inside this function,
	// the Go test runner will report the filename and line number of the code
	// that actually called this Equal() function
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
