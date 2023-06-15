# Builder image
FROM quay.io/centos/centos:stream8 as builder

USER root

# Dependencies
RUN dnf install -y golang git
RUN mkdir /build
WORKDIR /build

# Uncomment this block if the binary requires internal repositories
#ARG SSH_PRIVATE_KEY
#RUN mkdir /root/.ssh/
#RUN echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa
#RUN chmod 600 /root/.ssh/id_rsa
#RUN touch ~/.ssh/known_hosts
#RUN ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
#RUN git config --global url.ssh://org-25111032@github.com/smartcontractkit/.insteadOf https://github.com/smartcontractkit/

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
