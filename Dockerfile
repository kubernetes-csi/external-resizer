FROM gcr.io/distroless/static:latest-amd64
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Resizer"
ARG binary=./bin/csi-resizer

COPY ${binary} csi-resizer
ENTRYPOINT ["/csi-resizer"]
