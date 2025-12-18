# Lightweight build suitable for single-binary deployment.

FROM golang:1.25 as builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO disabled to keep image small; SQLite 功能将不可用。
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/znp ./cmd/znp

FROM debian:bookworm-slim
ENV ZNP_CONFIG=/etc/znp/znp.yaml
RUN groupadd -r znp && useradd -r -g znp znp
COPY --from=builder /out/znp /usr/local/bin/znp
COPY etc/znp-api.yaml /etc/znp/znp.yaml.sample
USER znp
EXPOSE 8888
ENTRYPOINT ["/usr/local/bin/znp"]
CMD ["serve", "--config", "/etc/znp/znp.yaml", "--migrate-to", "latest"]
