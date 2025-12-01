# Build stage
FROM golang:1.25-alpine AS builder

# Install git (needed for go mod download) and Node.js (for Tailwind CSS)
RUN apk add --no-cache git nodejs npm

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with module cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy package.json for npm dependency caching
COPY package.json ./

# Install Tailwind CSS with npm cache
RUN --mount=type=cache,target=/root/.npm \
    npm install

# Copy only files needed for CSS build first
COPY src/input.css ./src/input.css

# Build Tailwind CSS with minification
RUN npx @tailwindcss/cli -i ./src/input.css -o ./static/tailwind.min.css --minify

# Copy the rest of the source code
COPY . .

# Build the application with build cache and strip symbols
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static assets (including generated CSS)
COPY --from=builder /app/static ./static

# Create vault directory for data storage
RUN mkdir -p /vault

# Run the application (file watching is enabled by default)
# To disable watching, override with: -watch=false
CMD ["./main", "-path", "/vault"]
