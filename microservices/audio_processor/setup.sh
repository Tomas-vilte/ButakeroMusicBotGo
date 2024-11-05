openssl rand -base64 756 > ${PWD}/rs_keyfile
chmod 0400 ${PWD}/rs_keyfile
chown 999:999 ${PWD}/rs_keyfile