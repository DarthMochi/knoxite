//go:build (darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris) && !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris
// +build !windows

package main

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (statOS *StatOS) GetAvailableStorageSpace() (uint64, error) {
	var stat unix.Statfs_t

	wd := filepath.Join(cfg.StoragesPath)
	if err := unix.Statfs(wd, &stat); err != nil {
		return 0, err
	}

	return stat.Bavail * uint64(stat.Bsize), nil
}
