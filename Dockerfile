FROM golang:1.17-alpine3.14 as base

FROM base as dev
WORKDIR /build
ADD . /build
RUN go build ./cmd/edead

CMD ["./edead"]

FROM docker.io/alpine:3.14 AS prod
WORKDIR /app
COPY --from=dev /build/edead .
CMD [ "./edead" ]
