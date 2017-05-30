FROM golang:1.8
WORKDIR /go/src/app
COPY . .
RUN make update-deps && make
ENTRYPOINT ["bin/rstack"]
