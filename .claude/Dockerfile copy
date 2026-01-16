# Multi-stage build for Go ONVIF library
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/onvif-cli ./cmd/onvif-cli
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/onvif-quick ./cmd/onvif-quick

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S onvif && \
    adduser -u 1001 -S onvif -G onvif

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /bin/onvif-cli /usr/local/bin/
COPY --from=builder /bin/onvif-quick /usr/local/bin/

# Copy examples (optional)
COPY --from=builder /src/examples ./examples/

# Set ownership
RUN chown -R onvif:onvif /app

# Switch to non-root user
USER onvif

# Default command (run the quick tool)
CMD ["onvif-quick"]

# Labels
LABEL maintainer="ONVIF Library Team"
LABEL description="Go ONVIF library with CLI tools"
LABEL version="1.0.0"