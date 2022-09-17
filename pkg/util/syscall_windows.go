//go:build windows
// +build windows

package util

import (
	"fmt"
	"os"
)

func SyscallAccessWrite(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", dir)
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		return fmt.Errorf("write permission bit is not set on directory %s for user", dir)
	}

	return nil
}

func SyscallAccessRead(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", dir)
	}

	return nil
}
