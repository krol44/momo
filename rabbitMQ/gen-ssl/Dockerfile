FROM alpine:3.17.1

RUN apk add openssl
RUN mkdir "/ssl"

COPY script.sh .
RUN ./script.sh