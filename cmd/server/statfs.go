package main

type StatFS interface {
	GetAvailableStorageSpace() (uint64, error)
}

type StatOS struct {
	StatFS
}
