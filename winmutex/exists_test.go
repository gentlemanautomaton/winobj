//go:build windows

package winmutex_test

import (
	"testing"

	"github.com/gentlemanautomaton/winobj/winmutex"
)

func TestExistsShouldExist(t *testing.T) {
	name := testMutexName("ShouldExist")

	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	defer mutex.Close()

	exists, err := winmutex.Exists(name)
	if err != nil {
		t.Fatalf("Failed to check for existing mutex: %v", err)
	}

	if !exists {
		t.Fatalf("The winmutex.Exists() call returned false when it should have returned true")
	}
}

func TestExistsShouldNotExist(t *testing.T) {
	name := testMutexName("ShouldNotExist")

	exists, err := winmutex.Exists(name)
	if err != nil {
		t.Fatalf("Failed to check for existing mutex: %v", err)
	}

	if exists {
		t.Fatalf("The winmutex.Exists() call returned true when it should have returned false")
	}
}
