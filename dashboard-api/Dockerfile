FROM golang:1.18.2-buster as builder
WORKDIR /app
COPY dashboard-api/go.* ./
RUN go mod download
COPY dashboard-api/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o dashboard-api .

FROM alpine:latest
#COPY /ssl /ssl
#COPY /dist /dist
COPY --from=builder /app/dashboard-api .
CMD ["./dashboard-api"]