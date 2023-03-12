# Knoxite Server Setup

## Prerequisites
- golang >= v1.19
- npm/yarn

## Setup

To setup the server, execute the following command:

```
$ knoxite-server setup -d <db_file_name> -u <admin_user_name>  -p <admin_password> -P <admin_ui_port> -s <repositories_path> -C <config_file_name>
```

With the help flag, you can investigate the server configuration further:

```
$ knoxite-server setup --help
Usage:
  knoxite-server setup [flags]

Flags:
  -C, --configurl string     Path to configuration file (default "/home/<username>/.config/knoxite-server/knoxite-server.conf")
  -d, --dbfilename string    File name for sqlite database
  -h, --help                 help for setup
  -p, --password string      Admin password
  -P, --port string          Port for server api (default "42024")
  -s, --storagepath string   Path to client storages
  -n, --usehostname          Use hostname and dns, default=true (default true)
  -S, --usehttps             Use https, default=true (default true)
  -u, --username string      Admin username
```

This command will create a sqlite database file, to store information about the clients, like their quota and their authentication code.

It will create a login with the specified username and password, and specifies at what path on your server the repositories of your clients get stored.

The setup command will generate certificates as well, to make a https connection to your server possible.

It is possible to assign port 80 or 443 to the `admin_ui_port`, but this needs to be allowed by your firewall or iptables on your server.

When the setup command is finished, you can start the server with this command:

```
$ knoxite-server serve -C <config_file_path>
```

You can visit now `http(s)://<hostname>:<port>`, login with the specified username and password and create clients within the admin user interface.

