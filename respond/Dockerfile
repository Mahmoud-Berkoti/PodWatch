# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /respond ./respond

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /respond /respond

USER nonroot:nonroot

ENTRYPOINT ["/respond"]
