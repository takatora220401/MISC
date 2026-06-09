
FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod .
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o webserver .



FROM alpine:3.22

RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -D appuser

WORKDIR /app

COPY --from=builder /src/webserver .

RUN chown appuser:appgroup webserver

USER 1000

EXPOSE 8443

CMD ["./webserver"]