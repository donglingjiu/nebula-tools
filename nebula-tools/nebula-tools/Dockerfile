# 构建阶段
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/nebula ./cmd/server

# 运行阶段（前端已 embed 进二进制，无需额外文件）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl tzdata && adduser -D nebula
WORKDIR /app
COPY --from=build /out/nebula /app/nebula
ENV PORT=8080 NEBULA_DATA_DIR=/app/data
EXPOSE 8080
VOLUME /app/data
USER nebula
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD curl -sf http://localhost:8080/api/health || exit 1
ENTRYPOINT ["/app/nebula"]
