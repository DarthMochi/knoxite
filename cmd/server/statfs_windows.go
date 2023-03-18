//go:build (!darwin || !dragonfly || !freebsd || !linux || !netbsd || !openbsd || !solaris) && windows
// +build !darwin !dragonfly !freebsd !linux !netbsd !openbsd !solaris
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

func (statOS *StatOS) GetAvailableStorageSpace() (uint64, error) {
	h := syscall.NewLazyDLL("kernel32.dll")
	c := h.NewProc("GetPhysicallyInstalledSystemMemory")
	var freeBytes int64

	_, _, err := c.Call(uintptr(unsafe.Pointer(&freeBytes)))
	if err != nil {
		return 0, err
	}

	return uint64(freeBytes), nil
}
