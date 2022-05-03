FROM golang:1.18-alpine3.15 as base
RUN apk add --update make bash yarn ncurses git

FROM base as dev
WORKDIR /build
ADD . /build
RUN cd frontend; yarn install
RUN cd frontend; ./build-fe.sh
RUN go build -o edea-server ./cmd/edea-server
EXPOSE 3000/tcp
CMD ["./edea-server"]

FROM docker.io/python:3.10-alpine AS prod
WORKDIR /app
RUN apk add --update python3
RUN mkdir -p ./frontend/template
COPY --from=dev /build/edea-server .
COPY --from=dev /build/frontend/template ./frontend/template
COPY --from=dev /build/static ./static
EXPOSE 3000/tcp
CMD [ "./edea-server" ]
