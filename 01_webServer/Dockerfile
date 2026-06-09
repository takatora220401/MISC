# WARNING
# This image intentionally runs as root for debugging/development purposes.
# Do NOT use this configuration in production.
# In production, 
# - Create a dedicated non-root user and run the application as that user ex) 10001.
# - Consider using secret/sealed-secret/secret manager instead of copying cert and key.
# 
# ---------- Build Stage ----------

FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod .
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o webserver .

# ---------- Runtime Stage ----------

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /src/webserver .
COPY cert.pem .
COPY key.pem .

USER root

EXPOSE 8443

CMD ["./webserver"]