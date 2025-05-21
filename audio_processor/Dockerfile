FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal internal/

ARG TARGETARCH
ARG TARGETOS
ARG ENV=aws

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o main cmd/${ENV}/main.go

FROM alpine:3.21.3
RUN apk --no-cache add ca-certificates ffmpeg python3 py3-pip curl \
    && python3 -m venv /venv \
    && /venv/bin/pip install --upgrade pip \
    && /venv/bin/pip install --upgrade yt-dlp

RUN /venv/bin/yt-dlp --version

WORKDIR /root/
COPY --from=builder /app/main .

ENV PATH="/venv/bin:$PATH"

EXPOSE 8080

CMD ["./main"]