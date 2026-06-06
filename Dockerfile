FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -trimpath -ldflags="-s -w" -o uninaquiz-backend ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S appuser \
 && adduser -S -G appuser -H -s /sbin/nologin appuser

COPY --from=builder --chown=appuser:appuser /app/uninaquiz-backend /app/uninaquiz-backend

EXPOSE 8080

USER appuser

CMD ["./uninaquiz-backend"]
