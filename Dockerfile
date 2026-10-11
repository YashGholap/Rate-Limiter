# Stage 1: Build the Go application
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /out/rate-limiter ./cmd/server


# Stage 2: Minimal runtime image
FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && adduser -D -H appuserWORKDIR /app

COPY --from=builder /out/rate-limiter /app/rate-limiter

USER appuser

EXPOSE 8082

ENTRYPOINT ["/app/rate-limiter"]