#!/usr/bin/env bash

rm ../ssl/*
rm ../../grabber/ssl/*
rm ../../dashboard-api/ssl/*

# shellcheck disable=SC2046
export $(sed 's/[[:blank:]]//g; /^#/d' ../../.env.secrets | xargs)

sed -i -- "s/alt.dns/$DOMAIN/" script.sh && rm script.sh--

docker build . -t gen-ssl

docker container create --name gen-ssl gen-ssl
docker container cp gen-ssl:/ssl ../

docker container rm -f gen-ssl
docker rmi gen-ssl

sed -i -- "s/$DOMAIN/alt.dns/" script.sh && rm script.sh--

chmod -R 777 ../ssl/

cp ../ssl/ca-cert.pem ../../grabber/ssl/
cp ../ssl/client-cert.pem ../../grabber/ssl/
cp ../ssl/client-key.pem ../../grabber/ssl/

cp ../ssl/ca-cert.pem ../../dashboard-api/ssl/
cp ../ssl/client-cert.pem ../../dashboard-api/ssl/
cp ../ssl/client-key.pem ../../dashboard-api/ssl/