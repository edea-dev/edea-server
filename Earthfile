VERSION 0.6
FROM earthly/dind:alpine
WORKDIR /build
RUN apk add postgresql-client go make bash yarn ncurses git

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    WORKDIR /build
    COPY . /build
    RUN make deps
    RUN make build
    RUN go build -o build/edea-server ./cmd/edea-server
    SAVE ARTIFACT build/edea-server /edea-server AS LOCAL edea-server
    SAVE ARTIFACT frontend/template /frontend/template
    SAVE ARTIFACT static /static AS LOCAL static

docker:
    COPY +build/edea-server .
    COPY +build/edea-server .
    COPY +build/frontend/template ./frontend/template
    COPY +build/static ./static
    ENTRYPOINT ["/build/edea-server"]
    SAVE IMAGE --push edea-server:latest

integration-test:
    FROM +build
    COPY docker-compose.yml ./
    COPY frontend/test ./
    COPY integration-test.sh ./
    WITH DOCKER --load=edea-server:latest=+docker \
                --compose docker-compose.yml
        RUN ./integration-test.sh
    END

all:
  BUILD +build
  BUILD +docker
  BUILD +integration-test
