//go:build windows
// +build windows

package main

import (
	"golang.org/x/sys/windows"
)

func (statOS *StatOS) GetAvailableStorageSpace() (uint64, error) {
	h := windows.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytes int64

	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	_, _, err := c.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(wd))), uintptr(unsafe.Pointer(&freeBytes)), nil, nil)
	if err != nil {
		return 0, err
	}

	return uint64(freeBytes), nil
}
