#!/usr/bin/env sh
set -e

docker build . -t gen-ssl

docker container create --name gen-ssl gen-ssl
docker container cp gen-ssl:/ssl ../

docker container rm -f gen-ssl
docker rmi gen-ssl

cp ../ssl/ca-cert.pem ../../grabber/ssl/ca-cert.pem
cp ../ssl/client-cert.pem ../../grabber/ssl/client-cert.pem
cp ../ssl/client-key.pem ../../grabber/ssl/client-key.pem

cp ../ssl/ca-cert.pem ../../dashboard-api/ssl/ca-cert.pem
cp ../ssl/client-cert.pem ../../dashboard-api/ssl/client-cert.pem
cp ../ssl/client-key.pem ../../dashboard-api/ssl/client-key.pem