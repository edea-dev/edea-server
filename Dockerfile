FROM golang:1.17-alpine3.14 as base
RUN apk add --update make bash yarn ncurses

FROM base as dev
WORKDIR /build
ADD . /build
RUN make

CMD ["./edead"]

FROM docker.io/alpine:3.14 AS prod
WORKDIR /app
COPY --from=dev /build/edead .
CMD [ "./edead" ]
