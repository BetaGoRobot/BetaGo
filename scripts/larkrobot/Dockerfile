FROM ghcr.io/betagorobot/betago-golang:master as builder

COPY betago.zip /data/

WORKDIR /data/

RUN unzip betago.zip

RUN cd cmd/larkrobot && CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s" -o betaGo-source ./*.go && \
    upx -9 -o betaGo betaGo-source


# FROM ghcr.io/betagorobot/betago-chrome:master as runner
FROM debian as runner

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder  /data/cmd/larkrobot/betaGo /betaGo

COPY --from=builder ./data/fonts /data/fonts

COPY --from=builder ./data/images /data/images

COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./betaGo"]
