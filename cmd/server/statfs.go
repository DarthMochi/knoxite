//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

type StatFS interface {
	GetAvailableStorageSpace() (uint64, error)
}

type StatOS struct {
	StatFS
}
