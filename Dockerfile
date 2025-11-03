FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/subs ./cmd/app

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
WORKDIR /home/app
COPY --from=builder /app/subs /usr/local/bin/subs
COPY --from=builder /src/migrations ./migrations
COPY --from=builder /src/config.yaml ./config.yaml

EXPOSE 8080
ENV APP_PORT=8080
ENTRYPOINT ["subs"]