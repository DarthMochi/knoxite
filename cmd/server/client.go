//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/knoxite/knoxite/cmd/server/utils"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	Name      string
	AuthCode  string
	Quota     uint64
	UsedSpace uint64
}

func (a *App) NewClient(name string, quotaString string) (*url.URL, error) {
	u := &url.URL{}
	quota, err := strconv.ParseUint(quotaString, 10, 64)
	if err != nil {
		WarningLogger.Println(err)
		return u, err
	}

	availableSpace, err := a.AvailableSpaceMinusQuota()
	if err != nil {
		WarningLogger.Println(err)
		return u, err
	}

	if quota > availableSpace {
		WarningLogger.Println(errNoSpace)
		return u, errNoSpace
	}

	authcode := utils.GenerateToken(32)

	client := &Client{
		Name:     name,
		Quota:    quota,
		AuthCode: authcode,
	}

	if strings.Contains(client.Name, "..") {
		WarningLogger.Println(errInvalidURL)
		return u, errInvalidURL
	}

	a.DB.Create(client)

	storagePath := filepath.Join(cfg.StoragesPath, client.Name)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		WarningLogger.Println(errNoPath)
		a.DB.Delete(client)
		return u, errNoPath
	}

	u, err = utils.ParseClientURL(client.ID)
	if err != nil {
		WarningLogger.Println(errInvalidURL)
		os.RemoveAll(storagePath)
		a.DB.Delete(client)
		return u, errInvalidBody
	}

	return u, nil
}

func (a *App) UpdateClient(clientId string, name string, quotaString string) error {
	var client Client
	a.DB.First(&client, clientId)

	oldName := client.Name

	quota, err := strconv.ParseUint(quotaString, 10, 64)
	if err != nil {
		WarningLogger.Println(errInvalidBody)
		return errInvalidBody
	}

	availableSpace, err := a.AvailableSpacePlusQuota(client.Quota)
	if err != nil {
		WarningLogger.Println(errNoSpace)
		return errNoSpace
	}

	if quota > availableSpace || quota < client.UsedSpace {
		WarningLogger.Println(errNoSpace)
		return errNoSpace
	}

	client.Name = name
	client.Quota = quota

	if strings.Contains(client.Name, "..") {
		WarningLogger.Println(errInvalidURL)
		return errInvalidURL
	}

	storagePath := filepath.Join("/", cfg.StoragesPath, client.Name)
	if !utils.Exist(storagePath) {
		err = os.Rename(filepath.Join("/", cfg.StoragesPath, oldName), filepath.Join("/", cfg.StoragesPath, client.Name))
		if err != nil {
			WarningLogger.Println(err)
			return err
		}
	}
	a.DB.Model(&client).Where("id = ?", clientId).Updates(&client)
	return nil
}
