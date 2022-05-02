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
    RUN cd frontend; yarn install
    RUN cd frontend; ./build-fe.sh
    RUN go build -o build/edea-server ./cmd/edea-server
    SAVE ARTIFACT build/edea-server /edea-server AS LOCAL edea-server
    SAVE ARTIFACT frontend/template /frontend/template
    SAVE ARTIFACT static /static AS LOCAL static

docker:
    COPY +build/edea-server .
    COPY +build/frontend/template ./frontend/template
    COPY +build/static ./static
    EXPOSE 3000
    ENTRYPOINT ["/build/edea-server"]
    SAVE IMAGE --push edea-server:latest

tester:
    FROM mcr.microsoft.com/playwright:v1.21.0-focal
    WORKDIR /app
    COPY frontend/test .
    COPY integration-test.sh .
    ENTRYPOINT ["/app/integration-test.sh"]
    SAVE IMAGE tester:latest

integration-test:
    FROM +build
    COPY docker-compose.yml ./
    WITH DOCKER --load=edea-server:latest=+docker \
                --load=tester:latest=+tester \
                --compose docker-compose.yml \
                --service db \
                --service search
        RUN while ! pg_isready --host=localhost --port=5432 --dbname=edea --username=edea; do sleep 1; done ;\
            docker run  -e "DB_DSN=host=edea-db user=edea password=edea dbname=edea port=5432 sslmode=disable" \
                        -e "REPO_CACHE_BASE=/tmp/repo" \
                        -e "SEARCH_HOST=http://edea-meilisearch:7700" \
                        -e "SEARCH_INDEX=edea" \
                        -e "SEARCH_API_KEY=meiliedea" \
                        --network edea-net \
                        -d \
                        edea-server:latest; \
            docker run  -e "TEST_HOST=http://edea-server:3000" \
                        --network edea-net tester:latest
    END

all:
  BUILD +build
  BUILD +integration-test
