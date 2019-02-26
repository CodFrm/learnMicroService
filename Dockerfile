FROM golang as builder

COPY . /build/

WORKDIR /build

ENV HTTP_PROXY=http://10.0.75.1:1080/ \
    HTTPS_PROXY=http://10.0.75.1:1080/ \
    CGO_ENABLED=0 \
    GOOS=linux

RUN go get -d -v . && \
    go build -a -installsuffix cgo -o app . && ls

FROM alpine

COPY --from=builder /build/app /run

WORKDIR /run

RUN ls && chmod +x /run/app

CMD [ "./app" ]