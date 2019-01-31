FROM golang as builder

COPY ./ /build/

WORKDIR /build

RUN cd post \
    && go build -ldflags '-w -s'

FROM alpine

COPY --from=builder /build/main /run

WORKDIR /run

RUN chmod +x /run/main

CMD [ "./main" ]