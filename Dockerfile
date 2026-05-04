# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/neurofeed ./cmd/neurofeed

FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates cron tzdata \
	&& rm -rf /var/lib/apt/lists/*
COPY --from=build /out/neurofeed /usr/local/bin/neurofeed
COPY docker/run-daily.sh /usr/local/bin/run-daily.sh
COPY docker/neurofeed.cron /etc/cron.d/neurofeed
RUN chmod +x /usr/local/bin/neurofeed /usr/local/bin/run-daily.sh \
	&& chmod 0644 /etc/cron.d/neurofeed \
	&& sed -i 's/^session\s\+required\s\+pam_loginuid.so/#&/' /etc/pam.d/cron || true
CMD ["cron", "-f"]
