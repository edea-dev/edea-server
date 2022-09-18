VERSION 0.6
FROM earthly/dind:alpine
WORKDIR /build
RUN apk add postgresql-client go make bash yarn ncurses git

deps:
    FROM golang:1.19-alpine
    RUN apk add --update git
    WORKDIR /build
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

edea-tool:
    FROM alpine:edge

    ENV EDEA_VERSION=0.1.0

    WORKDIR /build

    RUN apk add --update git curl poetry
    RUN git clone https://gitlab.com/edea-dev/edea.git
    RUN cd edea; poetry build

    SAVE ARTIFACT edea/dist/edea-${EDEA_VERSION}-py3-none-any.whl

frontend:
    FROM +deps
    WORKDIR /build
    COPY --dir ./frontend /build
    RUN apk add --update yarn bash git
    RUN cd frontend; yarn install
    RUN cd frontend; ./build-fe.sh
    SAVE ARTIFACT frontend/template /frontend/template
    SAVE ARTIFACT static /static

build:
    FROM +deps
    WORKDIR /build
    COPY --dir ./cmd ./internal /build
    COPY ./embed.go /build
    COPY +frontend/static ./static
    RUN go build -o edea-server ./cmd/edea-server
    SAVE ARTIFACT edea-server /edea-server

# create a base image with the python tools, speeds up incremental builds a lot
docker-base:
    FROM alpine:edge
    WORKDIR /build
    RUN apk -U add py3-numpy py3-pip py3-cairosvg py3-pillow
    RUN apk add mdbook --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing/

    ENV EDEA_VERSION=0.1.0

    COPY +edea-tool/edea-${EDEA_VERSION}-py3-none-any.whl .

    RUN pip install edea-${EDEA_VERSION}-py3-none-any.whl
    RUN rm *.whl

docker:
    FROM +docker-base
    ARG ref

    COPY +build/edea-server .
    COPY +frontend/frontend/template ./frontend/template
    COPY +frontend/static ./static
    
    ENTRYPOINT /build/edea-server
    IF [ "$ref" = "" ]
       SAVE IMAGE --push edea-server:latest
    ELSE
        SAVE IMAGE --push $ref
    END

docker-test:
    FROM +docker
    COPY users.yml .

tester:
    FROM mcr.microsoft.com/playwright:v1.25.1-focal
    ARG ref

    WORKDIR /app
    COPY frontend/test .
    COPY integration-test.sh .
    ENTRYPOINT ["/app/integration-test.sh"]

    IF [ "$ref" = "" ]
       SAVE IMAGE --push tester:latest
    ELSE
        SAVE IMAGE --push $ref
    END

integration-test:
    COPY docker-compose.yml ./
    COPY ci.env ./ci.env
    COPY users.yml ./users.yml
    WITH DOCKER --load=edea-server:latest=+docker-test \
                --load=tester:latest=+tester \
                --compose docker-compose.yml \
                --service db \
                --service search
        RUN while ! pg_isready --host=localhost --port=5432 --dbname=edea --username=edea; do sleep 1; done ;\
            docker run --env-file ci.env --network build_default --name edea-server -d edea-server:latest; \
            docker run --env-file ci.env --network build_default tester:latest || (echo fail > fail; docker logs edea-server); \
            docker stop edea-server;
    END
    IF [ -f fail ]
        RUN echo "Integration tests have failed" \
            && exit 1
    END

all:
    BUILD +build
    BUILD +integration-test
