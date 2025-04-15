//go:build windows

package synchapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel = windows.NewLazySystemDLL("kernel32.dll")

	procCreateMutex   = modkernel.NewProc("CreateMutexW")
	procCreateMutexEx = modkernel.NewProc("CreateMutexExW")
	procOpenMutex     = modkernel.NewProc("OpenMutexW")
	procReleaseMutex  = modkernel.NewProc("ReleaseMutex")
)

// CreateMutex attempts to create a Windows mutex with the given name and
// attributes. If name is empty, it will created an unnamed mutex.
//
// When creating a named mutex, if a mutex with the given name already exists,
// openedExisting will be true and a handle for the existing mutex will be
// returned.
//
// If initial ownership is requested and the mutex does not already exist,
// it will be created in locked (signaled) state. If the named mutex exists
// already, a handle to the existing mutex is returned but it will not be
// be locked.
//
// When successful, a handle to the mutex is returned. The handle is bound
// to the calling thread. This means that the caller should call
// runtime.LockOSThread() before calling this function. If any calls to a wait
// function or to ReleaseMutex are made with this handle, they must be made
// from the same thread.
//
// https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-createmutexw
func CreateMutex(name string, initialOwner bool, attrs *syscall.SecurityAttributes) (h syscall.Handle, openedExisting bool, err error) {
	if len(name)+1 >= syscall.MAX_PATH {
		return 0, false, fmt.Errorf("create mutex: name length exceeds the %d character limit specified by MAX_PATH: %s", syscall.MAX_PATH, name)
	}

	var utf16Name *uint16
	if name != "" {
		var err error
		utf16Name, err = syscall.UTF16PtrFromString(name)
		if err != nil {
			return 0, false, err
		}
	}

	var bInitialOwner uintptr
	if initialOwner {
		bInitialOwner = 1
	}

	r0, _, e := syscall.SyscallN(
		procCreateMutex.Addr(),
		uintptr(unsafe.Pointer(attrs)),
		bInitialOwner,
		uintptr(unsafe.Pointer(utf16Name)))

	switch e {
	case syscall.ERROR_ALREADY_EXISTS:
		return syscall.Handle(r0), true, nil
	case 0:
		return syscall.Handle(r0), false, nil
	default:
		return syscall.Handle(r0), false, e
	}
}

// CreateMutexEx attempts to create a Windows mutex with the given name and
// attributes. If name is empty, it will created an unnamed mutex.
//
// When creating a named mutex, if a mutex with the given name already exists,
// openedExisting will be true and a handle for the existing mutex will be
// returned.
//
// When successful, a handle to the mutex is returned. The handle is bound
// to the calling thread. This means that the caller should call
// runtime.LockOSThread() before calling this function. If any calls to a wait
// function or to ReleaseMutex are made with this handle, they must be made
// from the same thread.
//
// TODO: Add support for flags and desired access settings.
//
// https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-createmutexexw
func CreateMutexEx(name string, attrs *syscall.SecurityAttributes) (h syscall.Handle, openedExisting bool, err error) {
	if len(name)+1 >= syscall.MAX_PATH {
		return 0, false, fmt.Errorf("create mutex: name length exceeds the %d character limit specified by MAX_PATH: %s", syscall.MAX_PATH, name)
	}

	var utf16Name *uint16
	if name != "" {
		var err error
		utf16Name, err = syscall.UTF16PtrFromString(name)
		if err != nil {
			return 0, false, err
		}
	}

	r0, _, e := syscall.SyscallN(
		procCreateMutexEx.Addr(),
		uintptr(unsafe.Pointer(attrs)),
		uintptr(unsafe.Pointer(utf16Name)),
		0,
		0)

	switch e {
	case syscall.ERROR_ALREADY_EXISTS:
		return syscall.Handle(r0), true, nil
	case 0:
		return syscall.Handle(r0), false, nil
	default:
		return syscall.Handle(r0), false, e
	}
}

// OpenMutex attempts to open an existing Windows mutex with the given name
// and attributes. If the named mutex does not already exist, it returns
// a non-nil error.
//
// When successful, a handle to the mutex is returned. The handle is bound
// to the calling thread. This means that the caller should call
// runtime.LockOSThread() before calling this function. If any calls to a wait
// function or to ReleaseMutex are made with this handle, they must be made
// from the same thread.
//
// TODO: Accept the desired access rights as a parameter.
//
// https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-openmutexw
func OpenMutex(name string) (syscall.Handle, error) {
	// Always use SYNCHRONIZE access rights when opening mutexes for now.
	//
	// See this document for possible access rights:
	// https://learn.microsoft.com/en-us/windows/win32/sync/synchronization-object-security-and-access-rights
	const synchronize = 0x00100000

	if len(name)+1 >= syscall.MAX_PATH {
		return 0, fmt.Errorf("open mutex: name length exceeds the %d character limit specified by MAX_PATH: %s", syscall.MAX_PATH, name)
	}

	var utf16Name *uint16
	if name != "" {
		var err error
		utf16Name, err = syscall.UTF16PtrFromString(name)
		if err != nil {
			return 0, err
		}
	}

	r0, _, e := syscall.SyscallN(
		procOpenMutex.Addr(),
		synchronize, // dwDesiredAccess
		0,           // bInheritHandle
		uintptr(unsafe.Pointer(utf16Name)))

	if r0 == 0 && e == 0 {
		e = syscall.EINVAL
	}

	var err error
	if e != 0 {
		err = e
	}

	return syscall.Handle(r0), err
}

// ReleaseMutex attempts to release the Windows mutex with the given handle.
//
// https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-releasemutex
func ReleaseMutex(h syscall.Handle) (released bool, err error) {
	r0, _, e := syscall.SyscallN(procReleaseMutex.Addr(), uintptr(h))

	if r0 != 0 {
		released = true
	}
	if e != 0 {
		err = e
	}
	return
}
