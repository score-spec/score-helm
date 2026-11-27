FROM --platform=$BUILDPLATFORM dhi.io/golang:1.27.4-alpine3.24-dev@sha256:a76849b53b06b671dff57ebea6a83206c54a74e5fbd27ba365b9e2dc94a18195 AS builder

ARG VERSION=0.0.0
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

# Set the current working directory inside the container.
WORKDIR /go/src/github.com/score-spec/score-helm

# Copy just the module bits
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project and build it.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w \
        -X github.com/score-spec/score-helm/internal/version.Version=${VERSION} \
        -X github.com/score-spec/score-helm/internal/version.GitCommit=${GIT_COMMIT} \
        -X github.com/score-spec/score-helm/internal/version.BuildDate=${BUILD_DATE}" \
    -o /usr/local/bin/score-helm ./cmd/score-helm

# We can use static since we don't rely on any linux libs or state, but we need ca-certificates to connect to https/oci with the init command.
FROM dhi.io/static:20261020-alpine3.24@sha256:4767f889704f46bbf9b83bd0225e0e868ae0cb45ddc5627350ecf1449a019721

# Set the current working directory inside the container.
WORKDIR /score-helm

# Copy the binary from the builder image.
COPY --from=builder /usr/local/bin/score-helm /usr/local/bin/score-helm

# Run the binary.
ENTRYPOINT ["/usr/local/bin/score-helm"]
