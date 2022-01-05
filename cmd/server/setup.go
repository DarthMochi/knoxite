package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	Name      string
	AuthCode  string
	Quota     uint64
	UsedSpace uint64
}

var (
	setupCmd = &cobra.Command{
		Use: "setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Save(configURL); err != nil {
				return err
			}

			if err := initDB(cfg.DBFileName); err != nil {
				return err
			}

			if err := initStoragePath(cfg.StoragesPath); err != nil {
				defer os.Remove(cfg.DBFileName)
				return err
			}
			return nil
		},
	}
)

func init() {
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminPassword, "password", "p", "", "Admin password")
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminUserName, "username", "u", "", "Admin username")
	setupCmd.PersistentFlags().StringVarP(&cfg.DBFileName, "dbfilename", "d", "", "File name for sqlite database")
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminUIPort, "port", "P", "42024", "Port for server api")
	setupCmd.PersistentFlags().StringVarP(&cfg.StoragesPath, "storagepath", "s", "", "Path to client storages")
	setupCmd.PersistentFlags().StringVarP(&configURL, "configurl", "C", config.DefaultPath(), "Path to configuration file")

	setupCmd.MarkFlagRequired("password")
	setupCmd.MarkFlagRequired("username")
	setupCmd.MarkFlagRequired("dbfilename")
	setupCmd.MarkFlagRequired("port")
	setupCmd.MarkFlagRequired("repopath")

	RootCmd.AddCommand(setupCmd)
}

func initDB(dbURL string) error {
	db, err := gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("couldn't connect to database")
	}
	db.AutoMigrate(&Client{})
	return nil
}

// Note: the repoURL, since it is a folder, needs to end with a "/", e.g. /tmp/repositories/
func initStoragePath(storageURL string) error {
	path, err := utils.PathToUrl(storageURL)
	if err != nil {
		return err
	}

	cfgDir := filepath.Dir(path.Path)
	if !utils.Exist(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return err
		}
	}

	return nil
}
