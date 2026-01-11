# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /ingest ./ingest

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /ingest /ingest

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/ingest"]
