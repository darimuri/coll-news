//go:build !windows
// +build !windows

package util

import (
	"golang.org/x/sys/unix"
)

func SyscallAccessWrite(dir string) error {
	return unix.Access(dir, unix.W_OK)
}

func SyscallAccessRead(dir string) error {
	return unix.Access(dir, unix.R_OK)
}
