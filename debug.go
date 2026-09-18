package main

import (
	"fmt"
	"os"
)

// dumpToFile writes a debug string to the given file path.
func dumpToFile(filename string, data string) {
	// 0644 is the file permission mode (read/write for owner, read-only for others)
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		err2 := fmt.Errorf("Failed to dump debug string to file: %v", err)
		panic(err2)
	}
}
