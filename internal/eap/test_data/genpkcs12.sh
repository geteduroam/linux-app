#!/bin/sh

mkdir -p pkcs12
openssl req -x509 -newkey rsa:4096 -keyout pkcs12/myKey.pem -out pkcs12/cert.pem -days 365 -nodes

openssl pkcs12 -export -out pkcs12/container.p12 -inkey pkcs12/myKey.pem -in pkcs12/cert.pem

cat pkcs12/container.p12 | base64
