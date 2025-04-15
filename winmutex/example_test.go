//go:build windows

package winmutex_test

import (
	"fmt"
	"os"

	"github.com/gentlemanautomaton/winobj/winmutex"
)

func Example() {
	const name = `Global\NamedMutexExample`

	mutex, err := winmutex.New(name)
	if err != nil {
		fmt.Printf("Failed to open the %s system mutex: %v\n", name, err)
		os.Exit(1)
	}
	defer mutex.Close()

	mutex.Lock()
	defer mutex.Unlock()

	fmt.Printf("Successfully acquired the %s system mutex.\n", mutex.Name())

	// Output: Successfully acquired the Global\NamedMutexExample system mutex.
}

func ExampleExists() {
	const msiMutex = `Global\_MSIExecute`

	running, err := winmutex.Exists(msiMutex)
	if err != nil {
		fmt.Printf("Failed to open the %s system mutex: %v\n", msiMutex, err)
		os.Exit(1)
	}

	if running {
		fmt.Println("The Windows Installer is currently at work.")
	} else {
		fmt.Println("The Windows Installer is currently dormant.")
	}

	// Output: The Windows Installer is currently dormant.
}
