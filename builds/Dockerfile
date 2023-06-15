# Builder image
FROM quay.io/centos/centos:stream8 as builder

USER root

# Dependencies
RUN dnf install -y golang git
RUN mkdir /build
WORKDIR /build

# Golang build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o timelock-worker .

# Final image
FROM quay.io/centos/centos:stream8
COPY --from=builder /build/timelock-worker /app/
WORKDIR /app

ENTRYPOINT ["./timelock-worker"]
CMD ["start", "--log-level", "${LOGLEVEL:-debug}"]
