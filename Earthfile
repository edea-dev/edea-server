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

tester:
    FROM mcr.microsoft.com/playwright:v1.21.0-focal
    COPY frontend/test .
    COPY integration-test.sh .
    ENTRYPOINT ["./integration-test.sh"]
    SAVE IMAGE tester:latest

integration-test:
    FROM +build
    COPY docker-compose.yml ./
    ENV DB_DSN "host=edea-db user=edea password=edea dbname=edea port=5432 sslmode=disable"
    ENV REPO_CACHE_BASE /tmp/repo
    ENV SEARCH_HOST http://edea-meilisearch:7700
    ENV SEARCH_INDEX edea
    ENV SEARCH_API_KEY meiliedea
    WITH DOCKER --load=edea-server:latest=+docker \
                --load=tester:latest=+tester \
                --compose docker-compose.yml \
                --service db \
                --service search
        RUN while ! pg_isready --host=localhost --port=5432 --dbname=edea --username=edea; do sleep 1; done ;\
            docker run edea-server:latest -d; \
            docker run -e  tester:latest
    END

all:
  BUILD +build
  BUILD +integration-test
