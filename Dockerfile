FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Resizer"

COPY ./bin/csi-resizer csi-resizer
ENTRYPOINT ["/csi-resizer"]
