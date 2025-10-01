#!/bin/bash

echo "Create server.key"
openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384r1 -out server.key

echo "Create server.csr"
openssl req -new -x509 -sha256 -key server.key -out server.crt -batch -days 365