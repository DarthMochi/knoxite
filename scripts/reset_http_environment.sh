#!/bin/bash
#
# knoxite
#     Copyright (c) 2016-2022, Christian Muehlhaeuser <muesli@gmail.com>
#     Copyright (c) 2021-2022, Raschaad Yassine <Raschaad@gmx.de>
#
#   For license see LICENSE
#

pkill -9 -f "./server serve"
rm -rf $STORAGES
rm -rf $SERVER_CONFIG
rm -rf $DATABASE_NAME