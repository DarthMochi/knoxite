#!/bin/bash
#
# knoxite
#     Copyright (c) 2016-2022, Christian Muehlhaeuser <muesli@gmail.com>
#     Copyright (c) 2021-2022, Raschaad Yassine <Raschaad@gmx.de>
#
#   For license see LICENSE
#

# Prerequisites:
#     - htpasswd (from apache-utils/apache-tools)
#     - jq (lightweight and flexible command-line JSON pocessor)

ADMIN_USERNAME=abc
ADMIN_PASSWORD=123
ADMIN_PORT=8080
export STORAGES=/tmp/repositories
export SERVER_CONFIG=knoxite-server.conf
export DATABASE_NAME=test.db
PASSWORD_HASH=$(htpasswd -bnBC 14 "" $ADMIN_PASSWORD | tr -d ':\n' | sed 's/$2y/$2a/')
TEST_CLIENT=testuser

# build knoxite server
go build -o knoxite-server ./cmd/server

# setup knoxite server
./knoxite-server setup -d $DATABASE_NAME -u $ADMIN_USERNAME -p $ADMIN_PASSWORD -P $ADMIN_PORT -s $STORAGES -C $SERVER_CONFIG

# serve knoxite server in background
./knoxite-server serve -C $SERVER_CONFIG &

# wait for knoxite server to boot up
sleep 2

# encode username and hashed password to base64 string
USER_AUTH=$(echo -n "$ADMIN_USERNAME:$PASSWORD_HASH" | base64 | tr -d "\n")

# create test client
curl -H "Authorization: Basic $USER_AUTH" -H "Content-Type: application/x-www-form-urlencoded" -X POST "http://localhost:$ADMIN_PORT/clients" -d "name=$TEST_CLIENT&quota=100000000"

# retrieve client info of testuser
JSON=$(curl -H "Authorization: Basic $USER_AUTH" http://localhost:$ADMIN_PORT/clients/1 -s)

# retrieve client auth_code
AUTH_CODE=$(echo $JSON | jq -r '.AuthCode')

# build knoxite url
export KNOXITE_HTTP_URL=http://$AUTH_CODE@localhost:$ADMIN_PORT

