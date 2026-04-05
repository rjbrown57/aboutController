FROM golang:1.25 AS builder

WORKDIR /workspace

# Cache dependencies first.
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the Go source needed to build the manager.
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build a static manager binary for the target platform.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
	go build -a -o manager ./cmd/main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
