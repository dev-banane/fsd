FROM node:22-alpine AS frontend

WORKDIR /build/src
COPY src/package.json src/package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY src/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend

WORKDIR /build
COPY go.mod ./
COPY *.go ./
COPY data/ ./data/
COPY --from=frontend /build/static ./static

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /webeye .

FROM alpine:3.23

RUN apk add --no-cache wget \
    && addgroup -S webeye \
    && adduser -S -G webeye webeye

COPY --from=backend /webeye /usr/local/bin/webeye

ENV WEBEYE_WHAZZUP=/data/whazzup.txt \
    WEBEYE_PORT=8080

EXPOSE 8080

USER webeye

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/webeye"]
