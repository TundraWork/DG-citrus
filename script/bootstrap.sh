#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
BIN_FILENAME=citrus_server
echo "$CURDIR/bin/${BIN_FILENAME}"
exec $CURDIR/bin/${BIN_FILENAME}