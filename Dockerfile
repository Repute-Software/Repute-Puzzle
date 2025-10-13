# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install templ CLI and generate templates
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o puzzle .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/puzzle .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/config.example.yaml ./config.yaml

# Create necessary directories
RUN mkdir -p /app/data /app/images /app/logs

# Expose port
EXPOSE 8080

# Run the application
CMD ["./puzzle"]

