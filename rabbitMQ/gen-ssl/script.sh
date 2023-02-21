#!/usr/bin/env sh

set -e

openssl req -newkey rsa:2048 -new -nodes -text -out /ssl/ca.csr -keyout /ssl/ca-key.pem -subj "/CN=CA"
openssl x509 -days 10000 -req -in /ssl/ca.csr -text -extfile /etc/ssl/openssl.cnf -extensions v3_ca -signkey /ssl/ca-key.pem -out /ssl/ca-cert.pem

openssl req -new -nodes -text -sha256 -key /ssl/ca-key.pem -subj "/CN=*" > /ssl/server.csr -keyout /ssl/server-key.pem
openssl x509 -days 10000 -req -extfile <(printf "subjectAltName=DNS:*, DNS:host.docker.internal") -in /ssl/server.csr -CA /ssl/ca-cert.pem -CAkey /ssl/ca-key.pem -CAcreateserial -out /ssl/server-cert.pem

openssl req -new -nodes -text -sha256 -key /ssl/ca-key.pem -subj "/CN=*" > /ssl/client.csr -keyout /ssl/client-key.pem
openssl x509 -days 10000 -req -extfile <(printf "subjectAltName=DNS:*,DNS:host.docker.internal") -in /ssl/client.csr -CA /ssl/ca-cert.pem -CAkey /ssl/ca-key.pem -CAcreateserial -out /ssl/client-cert.pem