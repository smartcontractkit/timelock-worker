# Builder image: Go 1.22 on Alpine
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /build

# Cache go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source code
COPY . .

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-w -s -extldflags "-static"' -o timelock-worker .

# Final image: minimal scratch with only the binary
FROM alpine:latest
# Install CA certs
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/timelock-worker /timelock-worker
ENTRYPOINT ["/timelock-worker"]
CMD ["start"]
