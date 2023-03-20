//go:build (!darwin || !dragonfly || !freebsd || !linux || !netbsd || !openbsd || !solaris) && windows
// +build !darwin !dragonfly !freebsd !linux !netbsd !openbsd !solaris
// +build windows

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (statOS *StatOS) GetAvailableStorageSpace() (uint64, error) {
	h := syscall.NewLazyDLL("kernel32.dll")
	c := h.NewProc("GetDiskFreeSpaceExW")
	var freeBytes uint64

	cValue, _, err := c.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(cfg.StoragesPath))), uintptr(unsafe.Pointer(&freeBytes)))
	if cValue != 1 && err != nil {
		return 0, err
	}

	return uint64(freeBytes), nil
}
