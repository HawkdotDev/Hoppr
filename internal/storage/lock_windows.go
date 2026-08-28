//go:build windows

package storage

import (
	"os"
	"syscall"
)

type fileLock struct {
	file *os.File
}

func newFileLock(lockPath string) (*fileLock, error) {
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &fileLock{file: f}, nil
}

func (fl *fileLock) RLock() error {
	var ol syscall.Overlapped
	// Shared lock on Windows: omit LOCKFILE_EXCLUSIVE_LOCK
	return syscall.LockFileEx(syscall.Handle(fl.file.Fd()), 0, 0, 1, 0, &ol)
}

func (fl *fileLock) Lock() error {
	var ol syscall.Overlapped
	const LOCKFILE_EXCLUSIVE_LOCK = 2
	return syscall.LockFileEx(syscall.Handle(fl.file.Fd()), LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &ol)
}

func (fl *fileLock) Unlock() error {
	var ol syscall.Overlapped
	_ = syscall.UnlockFileEx(syscall.Handle(fl.file.Fd()), 0, 1, 0, &ol)
	return fl.file.Close()
}
