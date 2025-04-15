//go:build windows

package winmutex

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	"github.com/gentlemanautomaton/winobj/api/synchapi"
	"github.com/gentlemanautomaton/winobj/internal/lockedthread"
)

// Mutex provides access to a single named or unnamed system mutex on
// Windows.
type Mutex struct {
	name string

	mutex  sync.Mutex
	thread *lockedthread.Thread
	handle syscall.Handle
	locked bool
}

// New returns a system mutex with the given name. If name is empty, it
// returns an unnamed mutex. If name is not empty and a mutex with the given
// name does not already exist, it is created.
//
// If the name is prefixed with "Global\", the mutex will be created or
// opened in the global namespace.
//
// If the name is prefixed with "Session\", the mutex will be created or
// opened in the session namespace.
//
// If the call is successful, it returns a non-nil Mutex. An operating system
// thread will be allocated for the duration of its existince. This is
// necessary to retain thread affinity for the underlying system handle.
//
// It is the caller's responsibility to close the mutex that is returned,
// which will close the underlying system handle and allow the allocated
// operating system thread to be reused to the gorouting thread pool. Closing
// the mutex will automatically unlock the mutex if it is locked at the time
// it is closed.
//
// If the mutex name is invalid, or if the calling process does not have
// sufficient permissions to create or access a named mutex, it returns
// an error and the mutex is not created or opened.
func New(name string) (*Mutex, error) {
	// Mutexes are bound to a specific operating system threads in Windows.
	// Prepare an OS thread that will be dedicated to holding the mutex.
	//
	// This is a sad waste of an operating system thread allocation, but
	// there's not much that can be done about it.
	thread := lockedthread.New()

	// Attempt to create or open the mutex via the OS thread.
	var (
		handle syscall.Handle
		err    error
	)
	thread.Run(func() {
		handle, _, err = synchapi.CreateMutex(name, false, nil)
	})

	// If mutex creation failed, close the thread and return the error.
	if err != nil {
		thread.Close()
		return nil, fmt.Errorf("winmutex: failed to create %s: %w", mutexDescription(name), err)
	}

	// Return the mutex that wraps the thread and system handle.
	return &Mutex{
		name:   name,
		thread: thread,
		handle: handle,
		locked: false,
	}, nil
}

// Name returns the name of the mutex.
//
// If the mutex is unnamed, it returns an empty string.
func (m *Mutex) Name() string {
	return m.name
}

// Lock locks the underlying system mutex represented by m. If the lock is
// already in use, the calling goroutine blocks until the mutex is available.
func (m *Mutex) Lock() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.thread == nil {
		panic("winmutex: Mutex.Lock() called on a mutex that has been closed")
	}

	var err error
	m.thread.Run(func() {
		_, err = syscall.WaitForSingleObject(m.handle, syscall.INFINITE)
	})
	if err != nil {
		panic(mutexWaitError(m.name, err))
	}

	m.locked = true
}

// TryLock tries to lock the underlying system mutex represented by m and
// reports whether it succeeded.
func (m *Mutex) TryLock() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.thread == nil {
		panic("winmutex: Mutex.TryLock() called on a mutex that has been closed")
	}

	var (
		event uint32
		err   error
	)
	m.thread.Run(func() {
		event, err = syscall.WaitForSingleObject(m.handle, 0)
	})
	if err != nil {
		panic(mutexWaitError(m.name, err))
	}

	if event == synchapi.WaitTimeout {
		return false
	}

	m.locked = true

	return true
}

// Unlock unlocks the underlying system mutex represented by m. It is a
// run-time error if m is not locked on entry to Unlock.
func (m *Mutex) Unlock() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.locked {
		panic("winmutex: Mutex.Unlock() called on a mutex that is not locked")
	}

	var (
		released bool
		err      error
	)
	m.thread.Run(func() {
		released, err = synchapi.ReleaseMutex(m.handle)
	})
	if err != nil {
		panic(fmt.Errorf("winmutex: Mutex.Unlock(): %w", err))
	}
	if !released {
		panic("winmutex: Mutex.Unlock() called on a mutex that was not locked, but was expected to be")
	}

	m.locked = false

	return
}

// Close releases the underlying system mutex handle and releases its
// operating system thread back into the goroutine thread pool.
//
// If the mutex is locked, it will be unlocked before being closed.
func (m *Mutex) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err1, err2, err3 error
	if m.thread != nil {
		if m.handle != 0 {
			m.thread.Run(func() {
				if m.locked {
					_, err1 = synchapi.ReleaseMutex(m.handle)
				}
				err2 = syscall.CloseHandle(m.handle)
			})
			m.handle = 0
			m.locked = false
		}
		err3 = m.thread.Close()
		m.thread = nil
	}

	return errors.Join(err1, err2, err3)
}

func mutexWaitError(name string, err error) error {
	return fmt.Errorf("winmutex: failed to wait for %s: %w", mutexDescription(name), err)
}

func mutexDescription(name string) string {
	if name == "" {
		return "an unnamed windows mutex"
	}
	return fmt.Sprintf("the windows mutex named \"%s\"", name)
}
