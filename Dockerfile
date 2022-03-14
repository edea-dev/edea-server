FROM golang:1.17-alpine3.15 as base
RUN apk add --update make bash yarn ncurses

FROM base as dev
WORKDIR /build
ADD . /build
RUN make build

CMD ["./edead"]

FROM docker.io/alpine:3.15 AS prod
WORKDIR /app
COPY --from=dev /build/edead .
CMD [ "./edead" ]
