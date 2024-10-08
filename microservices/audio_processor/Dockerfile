FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
ARG TARGETARCH
ARG TARGETOS
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates ffmpeg python3 py3-pip curl \
    && python3 -m venv /venv \
    && /venv/bin/pip install --upgrade pip \
    && /venv/bin/pip install yt-dlp \
    && /venv/bin/pip install -U https://github.com/coletdjnz/yt-dlp-youtube-oauth2/archive/refs/heads/master.zip

WORKDIR /root/
COPY --from=builder /app/main .

ENV PATH="/venv/bin:$PATH"

# Verificar las versiones de ffmpeg, yt-dlp y curl
RUN ffmpeg -version && yt-dlp --version && curl --version

EXPOSE 8080

CMD ["./main"]