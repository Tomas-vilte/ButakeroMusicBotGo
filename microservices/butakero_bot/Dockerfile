FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/

ARG TARGETARCH
ARG TARGETOS
ARG ENV=bot_aws

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o main cmd/${ENV}/main.go

FROM alpine:3.21.3

RUN apk add --no-cache curl


WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8081

CMD ["./main"]