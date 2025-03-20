FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
ARG TARGETARCH
ARG TARGETOS
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o main cmd/main.go

FROM alpine:3.21.3

WORKDIR /root/
COPY --from=builder /app/main .

CMD ["./main"]