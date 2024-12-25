# Build stage
FROM --platform=$BUILDPLATFORM golang:latest AS builder
WORKDIR /app
COPY . .
ARG TARGETARCH

# Build a fully static binary
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /app/server ./cmd

# Final stage
FROM scratch
COPY --from=builder /app/server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/server"]