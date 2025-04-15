//go:build windows

package winmutex_test

import (
	"sync"
	"testing"

	"github.com/gentlemanautomaton/winobj/winmutex"
)

func TestMutexLockBasic(t *testing.T) {
	name := testMutexName("LockBasic")

	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	mutex.Unlock()
	if err := mutex.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMutexTryLockBasic(t *testing.T) {
	name := testMutexName("TryLockBasic")

	mutex1, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	defer mutex1.Close()

	mutex1.Lock()
	defer mutex1.Unlock()

	mutex2, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	defer mutex2.Close()

	if mutex2.TryLock() {
		t.Fatalf("A lock was acquired when it should have been blocked")
	}
}

func TestMutexLockConcurrent(t *testing.T) {
	const threads = 128
	name := testMutexName("LockConcurrent")

	var wg sync.WaitGroup
	wg.Add(threads)
	for i := range threads {
		go func() {
			defer wg.Done()
			mutex, err := winmutex.New(name)
			if err != nil {
				panic(err)
			}
			mutex.Lock()
			mutex.Unlock()
			if err := mutex.Close(); err != nil {
				panic(err)
			}
			t.Logf("%d: success", i)
		}()
	}
	wg.Wait()
}

func TestMutexTryLockConcurrent(t *testing.T) {
	const threads = 128
	name := testMutexName("TryLockConcurrent")

	var wg sync.WaitGroup
	wg.Add(threads)
	for i := range threads {
		go func() {
			defer wg.Done()
			mutex, err := winmutex.New(name)
			if err != nil {
				panic(err)
			}
			if locked := mutex.TryLock(); !locked {
				t.Logf("%d: blocked", i)
				return
			}
			t.Logf("%d: locked", i)
			mutex.Unlock()
			if err := mutex.Close(); err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()
}

func TestMutexMulitpleClose(t *testing.T) {
	name := testMutexName("MulitpleClose")
	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	for range 48 {
		if err := mutex.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMutexReleaseViaClose(t *testing.T) {
	name := testMutexName("ReleaseViaClose")
	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	mutex.Close()
}

func TestMutexLockAfterClose(t *testing.T) {
	name := testMutexName("LockAfterClose")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	mutex.Close()
	mutex.Lock()
}

func TestMutexTryLockAfterClose(t *testing.T) {
	name := testMutexName("TryLockAfterClose")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	mutex.Close()
	mutex.TryLock()
}

func TestMutexUnlockAfterClose(t *testing.T) {
	name := testMutexName("UnlockAfterClose")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	mutex, err := winmutex.New(name)
	if err != nil {
		t.Fatal(err)
	}
	mutex.Close()
	mutex.Unlock()
}

func TestMutexBadName(t *testing.T) {
	name := `\`
	mutex, err := winmutex.New(name)
	if err == nil {
		t.Error("a mutex was successfully created with a bad name")
		mutex.Close()
	}
}

func testMutexName(name string) string {
	return "WinObj-WinMutex-Test-" + name
}
