# Build the manager binary
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download


ENV K8S_VERSION=v1.24.1
ENV HELM_VERSION=v3.9.0
ENV HELM_FILENAME=helm-${HELM_VERSION}-linux-amd64.tar.gz

# install helm and kubectl
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl 
RUN curl -L https://get.helm.sh/${HELM_FILENAME} | tar xz && mv linux-amd64/helm /bin/helm && rm -rf linux-amd64
# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static
WORKDIR /

COPY manifests /workspace/manifests
COPY assets /assets
COPY --from=builder /workspace/manager .
COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/
COPY --from=builder /bin/helm /usr/local/bin/


ENTRYPOINT ["/manager"]
