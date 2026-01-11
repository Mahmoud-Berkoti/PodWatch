# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /enrich ./enrich

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /enrich /enrich

USER nonroot:nonroot

ENTRYPOINT ["/enrich"]
