FROM golang:1.16 AS build-env
WORKDIR /go/src/app
COPY . .
RUN make

FROM alpine:latest
RUN apk --update add ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /app/
COPY --from=build-env /go/src/app/bin/rosculus /app/
ENTRYPOINT ["./rosculus"]
