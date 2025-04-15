package lockedthread

import (
	"runtime"
	"sync"
)

type command struct {
	fn   func()
	done chan<- struct{}
}

// Thread facilitates execution of functions on an operating system thread
// that is locked, ensuring that all functions execute on the same thread.
type Thread struct {
	mutex sync.Mutex
	cmds  chan<- command
	done  <-chan struct{}
}

// New returns a new Thread that allows commands to be executed on a locked
// operating system thread.
//
// It is the caller's responsibility to close the thread when finished with
// it.
func New() *Thread {
	// Prepare command and completion channels.
	cmds := make(chan command)
	done := make(chan struct{})

	// Launch a goroutine that will be locked to a system thread and will be
	// responsible for executing commands.
	go func(cmds <-chan command) {
		// Close the thread's done channel when the command channel is closed.
		defer close(done)

		// Lock the operating system thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Execute each command received via the command channel, until the
		// channel is closed.
		for cmd := range cmds {
			func(cmd command) {
				defer close(cmd.done) // Signal completion of the command
				cmd.fn()
			}(cmd)
		}
	}(cmds)

	// Return a thread object that is capable of sendind commands to the
	// locked OS thread.
	return &Thread{
		cmds: cmds,
		done: done,
	}
}

// Run executes the given function on the locked operating system thread.
func (t *Thread) Run(f func()) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Panic if the thread has already been closed.
	if t.cmds == nil {
		panic("lockedthread: Thread.Run() was called on a thread that has been closed")
	}

	// Prepare a channel that will be closed upon completion of the command.
	done := make(chan struct{})

	// Send the command to the OS thread.
	t.cmds <- command{fn: f, done: done}

	// Wait for the command to be completed.
	<-done
}

// Close stops the associated OS thread and releases any system resources.
func (t *Thread) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check whether the thread has already been closed.
	if t.cmds == nil {
		return nil
	}

	// Close the command channel, which will cause the thread to exit
	close(t.cmds)

	// Wait for the thread to exit.
	<-t.done

	// Mark the thread as closed.
	t.cmds = nil
	t.done = nil

	return nil
}
