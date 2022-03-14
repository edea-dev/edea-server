FROM golang:1.17-alpine3.15
WORKDIR /build

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    COPY embed.go .
    COPY static ./static
    COPY pkg ./pkg
    COPY internal ./internal
    COPY cmd ./cmd
    RUN go build -o build/edead ./cmd/edead
    SAVE ARTIFACT build/edead /edead AS LOCAL edead

docker:
    COPY +build/edead .
    ENTRYPOINT ["/build/edead"]
    SAVE IMAGE edead:latest
