FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go build -v -o app .

FROM alpine:3 AS deployment

WORKDIR /run/app

COPY --from=builder \
    /build/templates \
    ./templates

COPY --from=builder \
    /build/js \
    ./js

COPY --from=builder \
    /build/css \
    ./css

COPY --from=builder \
    /build/app \
    ./

RUN apk add tzdata && \
    ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN chmod 755 ./app

EXPOSE 5666

CMD ["./app"]