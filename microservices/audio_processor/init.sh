#!/bin/bash

# Crear el directorio de claves si no existe
mkdir -p /etc/mongodb/pki

# Generar el archivo rs_keyfile
openssl rand -base64 756 > /etc/mongodb/pki/keyfile

# Establecer permisos para el archivo
chmod 0400 /etc/mongodb/pki/keyfile
chown 999:999 /etc/mongodb/pki/keyfile

# Iniciar MongoDB con los parámetros dados
exec docker-entrypoint.sh "$@"
