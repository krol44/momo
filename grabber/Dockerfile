FROM golang:1.18.2-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o momo-grabber .

FROM alpine:latest
COPY /ssl /ssl
COPY --from=builder /app/momo-grabber .
CMD ["./momo-grabber"]