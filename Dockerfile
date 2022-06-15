# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details

FROM gcr.io/distroless/static:nonroot
ARG VERSION="dev"
ARG TARGETARCH

ENV SUMMARY="Modela Operator Container Image." \
    DESCRIPTION="This Docker image contains Modela Operator."

LABEL summary="$SUMMARY" \
      description="$DESCRIPTION" \
      io.k8s.display-name="$SUMMARY" \
      io.k8s.description="$DESCRIPTION" \
      name="Modela Operator" \
      vendor="Modela Contributors" \
      url="https://modela.ai/" \
      version="$VERSION" \
      release="1"

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
