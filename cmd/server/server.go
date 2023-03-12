//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/knoxite/knoxite/cmd/server/config"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/spf13/cobra"
)

var (
	configURL string
	cfg       = &config.ServerConfig{}

	RootCmd = &cobra.Command{
		Use:   "server",
		Short: "knoxite server is a http(s) backend for knoxite",
		Long: "knoxite server is a http(s) backend for knoxite\n" +
			"Complete setup information is available at https://github.com/knoxite/knoxite/cmd/server/README.md",
		SilenceErrors:     true,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}
)

func main() {
	shutdown.OnSignal(0, os.Interrupt, syscall.SIGTERM)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
