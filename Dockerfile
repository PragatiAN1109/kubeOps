
# Stage 1: Build the Go application
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Build a static Go binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /app/server

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy only the compiled application
COPY --from=builder /app/server /server

# Run as an unprivileged user
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/server"]
