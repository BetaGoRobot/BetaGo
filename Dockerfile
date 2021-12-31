FROM alpine as builder

COPY * /data/
ENV BOTAPI "1/MTA2Mjk=/Cw1HItHKZ9Q/wS7IIHnYUw=="

WORKDIR /data/

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add golang && \
    go build -o betaGo *.go 

FROM alpine as runner

COPY /data/ --from=builder 
WORKDIR /data/

CMD ["./betaGo"]