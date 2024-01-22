#!/usr/bin/sh

go build -gcflags=all="-N -l"
gdb ota_server

