#!/bin/bash

mkdir -p /etc/mongodb/pki

openssl rand -base64 756 > /etc/mongodb/pki/keyfile

chmod 0400 /etc/mongodb/pki/keyfile
chown 999:999 /etc/mongodb/pki/keyfile

exec docker-entrypoint.sh "$@"
