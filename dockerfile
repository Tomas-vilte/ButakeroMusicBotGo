FROM golang:1.21 AS builder

RUN apt-get update \
  && apt-get install -y build-essential libopus-dev libopusfile-dev \
  && go install github.com/bwmarrin/dca/cmd/dca@latest

WORKDIR /src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOARCH=amd64 go build -o /bin/butakero cmd/main.go

FROM ubuntu

ARG YT_DLP_VERSION="2023.10.13"

ENV DISCORDTOKEN=

RUN apt-get update \
  && apt-get install -y ffmpeg wget libopusfile0 \
  && wget "https://github.com/yt-dlp/yt-dlp/releases/download/${YT_DLP_VERSION}/yt-dlp_linux" -O /usr/local/bin/yt-dlp \
  && chmod +x /usr/local/bin/yt-dlp

COPY --from=builder /bin/butakero /bin/butakero
COPY --from=builder /go/bin/dca /usr/local/bin/dca

CMD ["/bin/butakero"]