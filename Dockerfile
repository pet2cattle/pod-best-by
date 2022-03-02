ARG TARGETARCH
ARG TARGETOS

FROM --platform=${BUILDPLATFORM} golang:1.17.0 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -o best-by main.go

# container
FROM alpine:3.12.3

RUN adduser -u 1000 -D nonroot

WORKDIR /
COPY --from=builder /workspace/best-by .
USER nonroot:nonroot

ENTRYPOINT ["/best-by"]