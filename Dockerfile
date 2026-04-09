FROM golang:1.26-alpine@sha256:c2a1f7b2095d046ae14b286b18413a05bb82c9bca9b25fe7ff5efef0f0826166 AS builder

ARG VERSION

# Set the current working directory inside the container.
WORKDIR /go/src/github.com/score-spec/score-helm

# Copy just the module bits
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project and build it.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/score-spec/score-helm/internal/version.Version=${VERSION}" -o /usr/local/bin/score-helm ./cmd/score-helm

# We can use gcr.io/distroless/static since we don't rely on any linux libs or state, but we need ca-certificates to connect to https/oci with the init command.
FROM gcr.io/distroless/static:530158861eebdbbf149f7e7e67bfe45eb433a35c@sha256:5c7e2b465ac6a2a4e5f4f7f722ce43b147dabe87cb21ac6c4007ae5178a1fa58

# Set the current working directory inside the container.
WORKDIR /score-helm

# Copy the binary from the builder image.
COPY --from=builder /usr/local/bin/score-helm /usr/local/bin/score-helm

# Run the binary.
ENTRYPOINT ["/usr/local/bin/score-helm"]