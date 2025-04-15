package lockedthread_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gentlemanautomaton/winobj/internal/lockedthread"
	"golang.org/x/sys/windows"
)

func TestThreadMulitpleClose(t *testing.T) {
	thread := lockedthread.New()
	for range 48 {
		if err := thread.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestThreadUseAfterClose(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	thread := lockedthread.New()
	thread.Close()
	thread.Run(func() {})
}

func TestThreadConcurrent(t *testing.T) {
	const (
		threads    = 128 // Number of locked threads to spawn
		concurrent = 128 // Number of concurrent actions on each thread
	)
	for i := range threads {
		t.Run(fmt.Sprintf("Thread %d", i), func(t *testing.T) {
			// Run this sub-test in parallel with others
			t.Parallel()

			// Create a new locked thread.
			thread := lockedthread.New()
			defer thread.Close()

			// Collect the ID of that thread for comparison.
			var lockedThreadID uint32
			thread.Run(func() {
				lockedThreadID = windows.GetCurrentThreadId()
			})

			// Make sure the testing thread ID doesn't match, because that
			// would be weird.
			if id := windows.GetCurrentThreadId(); id == lockedThreadID {
				t.Fatalf("The test thread ID and the thread-locked thread ID are the same.")
			}

			// Spawn concurrent operations via the thread.
			var wg sync.WaitGroup
			wg.Add(concurrent)
			for c := range concurrent {
				c := c
				go thread.Run(func() {
					defer wg.Done()
					if windows.GetCurrentThreadId() != lockedThreadID {
						panic(fmt.Sprintf("goroutine %d: the function ", c))
					}
				})
			}

			// Wait for all of the concurrent test to finish
			wg.Wait()
		})
	}
}
