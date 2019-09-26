FROM golang:1.13 AS build-env
WORKDIR /go/src/github.com/munisystem/rosculus
COPY . .
RUN make update-deps && make

FROM alpine:latest
RUN apk --update add ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /app/
COPY --from=build-env /go/src/github.com/munisystem/rosculus/bin/rosculus .
ENTRYPOINT ["./rosculus"]
