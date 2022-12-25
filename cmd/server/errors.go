package main

import "fmt"

var (
	errInvalidHeader = fmt.Errorf("invalid header params")
	errInvalidBody   = fmt.Errorf("invalid body params")
	errCantDelete    = fmt.Errorf("can't delete folders")
	errInvalidURL    = fmt.Errorf("invalid url")
	errNoAuth        = fmt.Errorf("no authorization was given")
	errNoSpace       = fmt.Errorf("client storage space used up")
	errNoPath        = fmt.Errorf("path not found")
)
