FROM golang:1.19 AS base

WORKDIR MAIN
ENV CGO_ENABLED=0

COPY go.* ./

COPY config config
COPY server.go server.go
COPY server server

RUN go build -tags netgo -ldflags '-w -s -extldflags "-static"' -o /go/bin/main server.go

########
FROM alpine:3.14.4 AS final

WORKDIR MAIN

COPY --from=base /go/bin/main .

ENTRYPOINT ["./main"]
