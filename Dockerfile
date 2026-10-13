# Dynamic builds
ARG BUILDER_IMAGE=golang:1.26-bookworm
ARG FINAL_IMAGE=debian:bookworm-slim

# Build stage
FROM --platform=${BUILDPLATFORM} ${BUILDER_IMAGE} AS builder

# Build args
ARG GIT_REVISION=""
ARG BUILD_DATE=""

# Platform args
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

# Use modules for dependencies
WORKDIR $GOPATH/src/go.bengfort.dev/wilson

COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy source code
COPY pkg/ pkg/
COPY cmd/ cmd/

# Build the Wilson binary
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} BUILD_DATE=$(date +%Y-%m-%d) \
    go build -o /go/bin/wilson \
    -ldflags="-X 'go.bengfort.dev/wilson/pkg.GitVersion=${GIT_REVISION}' -X 'go.bengfort.dev/wilson/pkg.BuildDate=${BUILD_DATE}'" \
    ./cmd/wilson

# Bundle/minify web assets into pkg/web/dist (esbuild Go API; no Node/npm in image)
# RUN rm -rf pkg/web/dist
# RUN /go/bin/wilson staticfiles -o pkg/web/dist

# Final stage
FROM ${FINAL_IMAGE} AS final

LABEL maintainer="Benjamin Bengfort <benjamin@bengfort.com>"
LABEL description=""

# Ensure ca-certificates are installed for external API calls
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the Wilson binary from the builder stage
COPY --from=builder /go/bin/wilson /usr/local/bin/wilson

EXPOSE 8000

CMD [ "/usr/local/bin/wilson", "serve" ]
