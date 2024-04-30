FROM ghcr.io/betagorobot/betago-golang:master as builder

COPY betago.zip /data/

WORKDIR /data/

RUN unzip betago.zip

RUN cd cmd/betago && CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s" -o betaGo-source ./*.go &&\
    upx -9 -o betaGo betaGo-source


FROM ghcr.io/betagorobot/betago-chrome:master as runner
# FROM scratch as runner

ARG BOTAPI \
    ROBOT_NAME \
    ROBOT_NAME \
    ROBOT_ID \
    TEST_CHAN_ID \
    NETEASE_PHONE \
    NETEASE_PASSWORD \
    COS_SECRET_ID \
    COS_SECRET_KEY \
    COS_BASE_URL \
    COS_BUCKET_REGION_URL \
    MAIL_SECRET \
    GPT_TOKEN \
    NEWS_API_KEY \
    GOTIFY_APPLICATION_TOKEN \
    GOTIFY_URL \
    GITHUB_SHA \
    GIT_COMMIT_MESSAGE

ENV BOTAPI=${BOTAPI} \
    ROBOT_NAME=${ROBOT_NAME} \
    ROBOT_ID=${ROBOT_ID} \
    TEST_CHAN_ID=${TEST_CHAN_ID} \
    NETEASE_PHONE=${NETEASE_PHONE} \
    NETEASE_PASSWORD=${NETEASE_PASSWORD} \
    COS_SECRET_ID=${COS_SECRET_ID} \
    COS_SECRET_KEY=${COS_SECRET_KEY} \
    COS_BASE_URL=${COS_BASE_URL} \
    COS_BUCKET_REGION_URL=${COS_BUCKET_REGION_URL} \
    MAIL_SECRET=${MAIL_SECRET} \
    GPT_TOKEN=${GPT_TOKEN} \
    NEWS_API_KEY=${NEWS_API_KEY} \
    GOTIFY_APPLICATION_TOKEN=${GOTIFY_APPLICATION_TOKEN} \
    GOTIFY_URL=${GOTIFY_URL} \
    GITHUB_SHA=${GITHUB_SHA} \
    GIT_COMMIT_MESSAGE=${GIT_COMMIT_MESSAGE}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder  /data/cmd/betago/betaGo /betaGo

COPY --from=builder ./data/fonts /data/fonts

COPY --from=builder ./data/images /data/images

COPY --from=builder  /usr/share/zoneinfo/ /usr/share/zoneinfo

WORKDIR /

CMD ["./betaGo"]
