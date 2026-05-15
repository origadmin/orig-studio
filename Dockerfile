FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /src

ENV GOWORK=off

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME=server
RUN CGO_ENABLED=1 GOOS=linux go build -tags dev -a -ldflags '-linkmode external -extldflags "-static"' -o /bin/origcms-${SERVICE_NAME} ./cmd/${SERVICE_NAME}

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata ffmpeg

WORKDIR /app

ARG SERVICE_NAME=server
COPY --from=builder /bin/origcms-${SERVICE_NAME} /app/origcms
COPY resources/ /app/resources/

ENV TZ=UTC

EXPOSE 8080

ENTRYPOINT ["/app/origcms"]
