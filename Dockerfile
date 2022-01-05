FROM alpine as builder

COPY * /data/
ENV GOPROXY https://goproxy.io,direct
ENV GO111MODULE="auto"
WORKDIR /data/


RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add go && \
    apk add -U tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    CGO_ENABLED=0 go build -ldflags="-w -s" -o betaGo *.go 


FROM scratch as runner

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add -U tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    
ARG BOTAPI

ENV BOTAPI=${BOTAPI} ROBOT_NAME=${ROBOT_NAME} ROBOT_ID=${ROBOT_ID} TEST_CHAN_ID=${TEST_CHAN_ID}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder  /data/betaGo /betaGo
COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./betaGo"]
