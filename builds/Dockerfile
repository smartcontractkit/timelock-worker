# Builder image: Go 1.22 on Alpine
FROM golang:1.24-alpine AS builder

# Install git (needed for go mod) and libc for static linking
RUN apk add --no-cache git build-base

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
FROM scratch
COPY --from=builder /build/timelock-worker /timelock-worker
ENTRYPOINT ["/timelock-worker"]
CMD ["start"]
