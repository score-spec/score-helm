FROM --platform=$BUILDPLATFORM dhi.io/golang:1.26.3-alpine3.23-dev@sha256:8c3d2807b83482ee8736d45c43239b8bfb2319837f22a87e56e20184299c94a7 AS builder

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
FROM dhi.io/static:20260413-alpine3.23@sha256:eeea5b5f4dc394069d2afb9e83af0b4af640709fda0c2cfbdbdbbb3a4b8ecf6f

# Set the current working directory inside the container.
WORKDIR /score-helm

# Copy the binary from the builder image.
COPY --from=builder /usr/local/bin/score-helm /usr/local/bin/score-helm

# Run the binary.
ENTRYPOINT ["/usr/local/bin/score-helm"]
