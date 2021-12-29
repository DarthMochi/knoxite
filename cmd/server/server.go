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
		Short: "Knoxite is a data storage & backup tool",
		Long: "Knoxite is a secure and flexible data storage and backup tool\n" +
			"Complete documentation is available at https://github.com/knoxite/knoxite",
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
