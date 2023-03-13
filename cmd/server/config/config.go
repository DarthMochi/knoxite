//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package config

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"

	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/pelletier/go-toml"
)

const appName = "knoxite-server"

var cfgFileName = "knoxite-server.conf"

type ServerConfig struct {
	AdminUserName string `toml:"admin_user_name" comment:"User name to authorize api calls"`
	AdminPassword string `toml:"admin_password" comment:"Password to authorize admin user"`
	AdminUIPort   string `toml:"admin_ui_port" comment:"Port to run the server api on"`
	StoragesPath  string `toml:"repositories_path" comment:"Path to store the repositories of clients"`
	DBFileName    string `toml:"db_file_name" comment:"Database file name"`
	UseHostname   bool   `toml:"use_hostname" comment:"Use hostname as dns"`
	UseHTTPS      bool   `toml:"use_https" comment:"Use https"`
	AdminHostname string `toml:"admin_hostname" comment:"Hostname of the server"`
}

func DefaultPath() string {
	return utils.DefaultPath(appName, cfgFileName)
}

func (sc *ServerConfig) Save(u string) error {
	path, err := utils.PathToUrl(u)
	if err != nil {
		return err
	}

	sc.AdminPassword, err = utils.HashPassword(sc.AdminPassword)
	if err != nil {
		return err
	}

	sc.StoragesPath, err = filepath.Abs(sc.StoragesPath)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(*sc); err != nil {
		return err
	}

	cfgDir := filepath.Dir(path.Path)
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path.Path, buf.Bytes(), 0600)
}

func (sc *ServerConfig) Load(u *url.URL) error {
	content, err := os.ReadFile(u.Path)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(content, sc)
	if err != nil {
		return err
	}
	return nil
}
