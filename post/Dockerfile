FROM golang as builder

COPY ./ /build/

WORKDIR /build

RUN go get github.com/CodFrm/learnMicroService &&\
    go build -ldflags '-w -s'exit

FROM alpine

COPY --from=builder /build /run

WORKDIR /run

RUN chmod +x /run/main

CMD [ "./main" ]