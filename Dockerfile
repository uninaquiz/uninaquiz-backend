FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go mod download && \
    go build trimpath -ldflags="-s -w" -o uninaquiz-backend .

FROM alpine:latest AS uninaquiz-backend

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S appuser \
 && adduser -S -G appuser -H -s /sbin/nologin appuser

COPY --from=builder --chown=appuser:appuser /app/uninaquiz-backend /app/uninaquiz-backend
COPY --from=builder /app/uninaquiz-backend .

CMD ["./uninaquiz-backend"]
