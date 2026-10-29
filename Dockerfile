FROM --platform=$BUILDPLATFORM golang:1.26-trixie AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=vendor -o /out/serial-driver ./cmd/service

FROM debian:trixie-slim
WORKDIR /app
COPY autogen/oapi3.yaml /app/autogen/oapi3.yaml
COPY --from=build /out/serial-driver /usr/local/bin/serial-driver
USER 65532:65532
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/serial-driver"]
