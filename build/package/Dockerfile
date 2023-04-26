FROM golang:1.20.0-bullseye AS BUILD-ENV

RUN git config --global url."ssh://git@github.com".insteadOf "https://github.com"
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR /app

# Download dependencies
COPY ./go.mod ./go.sum ./
RUN --mount=type=ssh go mod download

# Copy source
COPY ./ ./

# Build binary
RUN env GOOS=linux GOARCH=amd64 go build -o /go/bin/admgr ./cmd/admgr

# Next Stage
FROM debian:bullseye-slim

COPY --from=BUILD-ENV /go/bin/admgr /go/bin/admgr

ENTRYPOINT ["/go/bin/admgr"]
