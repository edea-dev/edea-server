FROM golang:1.18-alpine3.15 as base
RUN apk add --update make bash yarn ncurses git

FROM base as dev
WORKDIR /build
ADD . /build
RUN make deps
RUN make build
EXPOSE 3000/tcp
CMD ["./edead"]

FROM docker.io/alpine:3.15 AS prod
WORKDIR /app
COPY --from=dev /build/edead .
EXPOSE 3000/tcp
CMD [ "./edead" ]
