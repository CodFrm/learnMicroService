FROM golang as builder

COPY ./ /build/

WORKDIR /build

RUN wget https://github.com/grpc/grpc-go/archive/v1.18.0.tar.gz && \
    tar -zxvf v1.18.0.tar.gz && \
    mkdir -p $GOPATH/src/google.golang.org/grpc && \
    mv grpc-go-1.18.0 $GOPATH/src/google.golang.org/grpc && \
    go get -d -v ./...;exit 0

RUN go build -ldflags '-w -s' && ls -l

FROM alpine

COPY --from=builder /build /run

WORKDIR /run

RUN chmod +x /run/main

CMD [ "./main" ]