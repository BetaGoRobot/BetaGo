FROM golang:1.18.3-alpine as builder

COPY . /data/
ENV GOPROXY https://goproxy.io,direct
ENV GO111MODULE="auto"
WORKDIR /data/


RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add -U tzdata && \
    apk add upx &&\
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s" -o betaGo-source *.go &&\
    upx -9 -o betaGo betaGo-source
    

# FROM alpine as runner
FROM scratch as runner

ARG BOTAPI ROBOT_NAME ROBOT_NAME ROBOT_ID TEST_CHAN_ID NETEASE_PHONE NETEASE_PASSWORD

ENV BOTAPI=${BOTAPI} ROBOT_NAME=${ROBOT_NAME} ROBOT_ID=${ROBOT_ID} TEST_CHAN_ID=${TEST_CHAN_ID} NETEASE_PHONE=${NETEASE_PHONE} NETEASE_PASSWORD=${NETEASE_PASSWORD}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder  /data/betaGo /betaGo

COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./betaGo"]
