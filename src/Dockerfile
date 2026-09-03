# builder
FROM golang:1.27-alpine AS builder
WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download


COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o main


# runtime
FROM alpine:3 AS runtime

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/main .
COPY --from=builder /app/fonts ./fonts
COPY --from=builder /app/assets ./assets

# 実行ユーザーを非rootに変更
USER appuser

EXPOSE 8080
CMD ["/app/main"]
