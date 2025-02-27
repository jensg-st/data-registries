FROM golang:1.23.3 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY main.go main.go
RUN go mod download

RUN CGO_ENABLED=0 go build -o server main.go

FROM alpine
COPY --chown=0:0 --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]