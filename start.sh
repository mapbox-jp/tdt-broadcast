#!/bin/bash
/usr/sbin/sshd -D &
go build -o app ./main.go && ./app
