FROM ghcr.io/betagorobot/betago-golang:master as builder

WORKDIR /app
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd cmd/larkrobot && go build -ldflags="-w -s" -o /app/larkrobot

# --- (!!! 在这里添加新步骤 !!!) ---
# 把 jieba 字典从缓存卷复制到镜像层中
RUN --mount=type=cache,target=/go/pkg/mod \
    mkdir -p /tmp/jieba_dict && \
    cp -r /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba/dict /tmp/jieba_dict


# -----------------
# Runner 阶段
# -----------------
FROM kevinmatt/libvips as runner

# --- (!!! 在这里修改 COPY 路径 !!!) ---
# 从 builder 层的 /tmp/jieba_dict 复制
COPY --from=builder /tmp/jieba_dict/dict /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba/dict
# --- (其他 COPY 保持不变) ---
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo
COPY --from=builder /app/larkrobot /larkrobot
COPY --from=builder /app/fonts /data/fonts
COPY --from=builder /app/assets/images /data/images

WORKDIR /

CMD ["./larkrobot"]