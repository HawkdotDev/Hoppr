//go:build windows

package storage

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	lockfileExclusiveLock = 0x00000002
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
	r1, _, err := procLockFileEx.Call(
		fl.file.Fd(),
		0, // 0 = Shared lock
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func (fl *fileLock) Lock() error {
	var ol syscall.Overlapped
	r1, _, err := procLockFileEx.Call(
		fl.file.Fd(),
		uintptr(lockfileExclusiveLock),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func (fl *fileLock) Unlock() error {
	var ol syscall.Overlapped
	r1, _, err := procUnlockFileEx.Call(
		fl.file.Fd(),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	_ = fl.file.Close()
	if r1 == 0 {
		return err
	}
	return nil
}
