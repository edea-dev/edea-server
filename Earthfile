VERSION 0.6
FROM earthly/dind:alpine
WORKDIR /build
RUN apk add postgresql-client go make bash yarn ncurses git

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

numpy:
    FROM docker.io/python:3.10-alpine

    ENV NUMPY_VERSION=1.22.3

    WORKDIR /build
    RUN apk add --update musl-dev linux-headers g++ git
    RUN curl -sSL https://github.com/numpy/numpy/releases/download/v${NUMPY_VERSION}/numpy-${NUMPY_VERSION}.tar.gz -o numpy.tar.gz
    RUN tar xf numpy.tar.gz
    RUN cd numpy-${NUMPY_VERSION}; pip wheel -w dist .
    SAVE ARTIFACT numpy-${NUMPY_VERSION}/dist/numpy-${NUMPY_VERSION}-cp310-cp310-linux_x86_64.whl

edea-tool:
    FROM docker.io/python:3.10-alpine

    ENV NUMPY_VERSION=1.22.3
    ENV EDEA_VERSION=0.1.0

    WORKDIR /build

    COPY +numpy/numpy-${NUMPY_VERSION}-cp310-cp310-linux_x86_64.whl .

    RUN apk add --update git curl    
    RUN curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python -
    RUN git clone https://gitlab.com/edea-dev/edea.git
    RUN cd edea; ~/.poetry/bin/poetry run pip install /build/numpy-${NUMPY_VERSION}-cp310-cp310-linux_x86_64.whl
    RUN cd edea; ~/.poetry/bin/poetry install --no-dev
    RUN cd edea; ~/.poetry/bin/poetry build

    SAVE ARTIFACT edea/dist/edea-${EDEA_VERSION}-py3-none-any.whl

build:
    FROM +deps
    WORKDIR /build
    COPY . /build
    RUN cd frontend; yarn install
    RUN cd frontend; ./build-fe.sh
    RUN go build -o build/edea-server ./cmd/edea-server
    SAVE ARTIFACT build/edea-server /edea-server
    SAVE ARTIFACT frontend/template /frontend/template
    SAVE ARTIFACT static /static

docker:
    FROM docker.io/python:3.10-alpine

    ENV NUMPY_VERSION=1.22.3
    ENV EDEA_VERSION=0.1.0

    COPY +build/edea-server .
    COPY +build/frontend/template ./frontend/template
    COPY +build/static ./static
    COPY +edea-tool/edea-${EDEA_VERSION}-py3-none-any.whl .
    COPY +numpy/numpy-${NUMPY_VERSION}-cp310-cp310-linux_x86_64.whl .

    RUN pip install numpy-${NUMPY_VERSION}-cp310-cp310-linux_x86_64.whl
    RUN pip install edea-${EDEA_VERSION}-py3-none-any.whl
    RUN rm *.whl
    EXPOSE 80 3000
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
    COPY ci.env ./ci.env
    WITH DOCKER --load=edea-server:latest=+docker \
                --load=tester:latest=+tester \
                --compose docker-compose.yml \
                --service db \
                --service search
        RUN while ! pg_isready --host=localhost --port=5432 --dbname=edea --username=edea; do sleep 1; done ;\
            docker run --env-file ci.env --network build_default --name edea-server -d edea-server:latest; \
            docker run --env-file ci.env --network build_default tester:latest; \
            docker logs edea-server
    END

all:
    BUILD +build
    BUILD +integration-test
