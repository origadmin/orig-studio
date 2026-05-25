FROM docker.1ms.run/oven/bun:1 AS frontend-builder

WORKDIR /web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

FROM docker.1ms.run/golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

ENV GOWORK=off
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
COPY --from=frontend-builder /web/dist ./web/dist/

ARG SERVICE_NAME=server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/origcms-${SERVICE_NAME} ./cmd/${SERVICE_NAME}

FROM docker.1ms.run/alpine:3.20

RUN apk --no-cache add ca-certificates tzdata ffmpeg

WORKDIR /app

ARG SERVICE_NAME=server
COPY --from=builder /bin/origcms-${SERVICE_NAME} /app/origcms
COPY --from=builder /src/resources/ /app/resources/

ENV TZ=UTC

EXPOSE 8080

ENTRYPOINT ["/app/origcms"]
