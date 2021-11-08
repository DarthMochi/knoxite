package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	Name     string
	AuthCode string
}

var setupCmd = &cobra.Command{
	Use: "setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("couldn't connect to database")
		}
		db.AutoMigrate(&Client{})
		db.Create(&Client{Name: "Drachenclient", AuthCode: "1510"})
		return nil
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
