FROM golang:1.24.2-alpine as builder

ENV GOPROXY https://goproxy.io,direct
ENV GO111MODULE="auto"
WORKDIR /data/


RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add -U tzdata && \
    apk add upx && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN apk add vips-dev g++