//go:build !windows

package storage

import (
	"os"
	"syscall"
)

type fileLock struct {
	file *os.File
}

func newFileLock(lockPath string) (*fileLock, error) {
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	return &fileLock{file: f}, nil
}

func (fl *fileLock) RLock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_SH)
}

func (fl *fileLock) Lock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX)
}

func (fl *fileLock) Unlock() error {
	_ = syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	return fl.file.Close()
}
