FROM ghcr.io/betagorobot/betago-golang:master as builder

COPY betago.zip /data/

WORKDIR /data/

RUN apk add vips-dev g++

RUN unzip betago.zip

RUN cd cmd/larkrobot && go build -ldflags="-w -s" -o larkrobot-source ./*.go && \
    upx -9 -o larkrobot larkrobot-source


# FROM ghcr.io/betagorobot/betago-chrome:master as runner
FROM kevinmatt/libvips as runner

COPY --from=builder /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba/dict /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder  /data/cmd/larkrobot/larkrobot /larkrobot

COPY --from=builder ./data/fonts /data/fonts

COPY --from=builder ./data/assets/images /data/images

COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./larkrobot"]
