#!/bin/bash
docker build . -t momo-grabber-image
docker rm -f momo-grabber
docker run -d -e TZ="Europe/Moscow" \
  -e HOSTNAME=$(hostname) \
  -e AMQP_URL="" \
  -e CA_CERT="/ssl/ca-cert.pem" \
  -e CLIENT_CERT="/ssl/client-cert.pem" \
  -e CLIENT_KEY="/ssl/client-key.pem" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --restart=always --log-opt max-size=5m --name=momo-grabber momo-grabber-image

rm -r /tmp/momo-service/
rm -r /tmp/momo-service.zip