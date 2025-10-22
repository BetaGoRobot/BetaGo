# -----------------
# Builder 阶段
# -----------------
FROM ghcr.io/betagorobot/betago-golang:master as builder

# 1. 安装系统依赖 (很少变动，放最前面)
RUN apk add vips-dev g++

WORKDIR /app

# 2. 仅复制 go.mod 和 go.sum
# 这是为了让 Docker 的“层缓存”生效
COPY go.mod go.sum ./

# 3. 下载 Go 模块 (!!! 关键优化 !!!)
# 我们使用 --mount=type=cache 来挂载一个持久化的 Go 模块缓存
# 即使 Action runner 是全新的，BuildKit 也能重用这个缓存
# 这解决了你“下载mod”的痛点
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 4. 复制所有剩余的源代码
# .dockerignore 文件会在这里生效，排除不必要的文件
COPY . .

# 5. 编译 Go 应用 (!!! 关键优化 !!!)
# 我们同时挂载 Go 模块缓存和 Go 构建缓存
# 这解决了你“go build 缓存”的痛点
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd cmd/larkrobot && go build -ldflags="-w -s" -o /app/larkrobot


# -----------------
# Runner 阶段
# -----------------
FROM kevinmatt/libvips as runner

# 复制必要的文件
COPY --from=builder /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba/dict /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.6/deps/cppjieba/dict
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo
COPY --from=builder /app/larkrobot /larkrobot
COPY --from=builder /app/data/fonts /data/fonts
COPY --from=builder /app/data/assets/images /data/images

WORKDIR /

CMD ["./larkrobot"]