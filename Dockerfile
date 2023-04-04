FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY *.go ./
RUN go build -o owners ./cmd

FROM alpine

RUN apk add --no-cache git

COPY entrypoint /entrypoint.sh
COPY --from=builder /app/owners /usr/local/bin/

ENTRYPOINT ["/entrypoint.sh"]
