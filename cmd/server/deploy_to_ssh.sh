#!/bin/sh
TARGET=$1
INSTALL_PATH=$2
PASSWD_FILE=$3
sshpass -f $PASSWD_FILE scp -r $(pwd)/../../ $TARGET:$INSTALL_PATH
# sshpass -f $PASSWD_FILE  ssh -o StrictHostKeyChecking=no $TARGET cd $INSTALL_PATH; go build ./cmd/server
exit 0