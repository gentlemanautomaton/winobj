//go:build windows

package synchapi

// Special return values for syscall.WaitForSingleObject().
const (
	WaitAbandoned = 0x00000080 // WAIT_ABANDONED
	WaitTimeout   = 0x00000102 // WAIT_TIMEOUT
)
