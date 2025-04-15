//go:build windows

// Package winmutex provides access to system mutexes on Windows.
//
// The package is designed to follow idiomatic Go programming conventions
// and to hide the peculiarities of mutex handling on Windows.
//
// The primary use of this package is to create and evaluate named mutexes
// that are accessible by multiple processes. This is probably not the right
// package to use if you have any other use case.
package winmutex
