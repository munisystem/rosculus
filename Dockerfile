FROM golang:1.8 AS build-env
WORKDIR /go/src/app
COPY . .
RUN make update-deps && make

FROM alpine:latest
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
WORKDIR /app/
COPY --from=build-env /go/src/app/bin/rosculus /app/
ENTRYPOINT ["./rosculus"]
