//go:build windows

package winmutex

import (
	"runtime"
	"syscall"

	"github.com/gentlemanautomaton/winobj/api/synchapi"
)

// Exists returns true if a mutex with the given name exists.
func Exists(name string) (bool, error) {
	// Lock the operating system thread for the duration of this call, as
	// the API calls require thread affinity.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Attempt to open an existing mutex with the given name.
	handle, err := synchapi.OpenMutex(name)
	if err != nil {
		if err, ok := (err).(syscall.Errno); ok {
			if err == syscall.ERROR_FILE_NOT_FOUND {
				return false, nil
			}
		}
		return false, err
	}

	// If we succeeded in opening the handle, be sure to close it.
	defer syscall.CloseHandle(handle)

	return true, nil
}
