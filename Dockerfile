FROM kevinmatt/betago-alpine-golang:1.18.3 as builder

COPY betago.zip /data/

WORKDIR /data/

RUN unzip betago.zip

RUN CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s" -o betaGo-source ./*.go &&\
    upx -9 -o betaGo betaGo-source
    

# # FROM alpine as runner
FROM alpine as runner

ARG BOTAPI ROBOT_NAME ROBOT_NAME ROBOT_ID TEST_CHAN_ID NETEASE_PHONE NETEASE_PASSWORD COS_SECRET_ID COS_SECRET_KEY COS_BASE_URL COS_BUCKET_REGION_URL

ENV BOTAPI=${BOTAPI} ROBOT_NAME=${ROBOT_NAME} ROBOT_ID=${ROBOT_ID} TEST_CHAN_ID=${TEST_CHAN_ID} NETEASE_PHONE=${NETEASE_PHONE} NETEASE_PASSWORD=${NETEASE_PASSWORD} COS_SECRET_ID=${COS_SECRET_ID} COS_SECRET_KEY=${COS_SECRET_KEY} COS_BASE_URL=${COS_BASE_URL} COS_BUCKET_REGION_URL=${COS_BUCKET_REGION_URL}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder  /data/betaGo /betaGo

COPY --from=builder ./data/fonts /data/fonts

COPY --from=builder ./data/images /data/images

COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./betaGo"]
