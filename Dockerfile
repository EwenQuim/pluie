# Build stage
FROM golang:1.25-alpine AS builder

# Install git (needed for go mod download) and Node.js (for Tailwind CSS)
RUN apk add --no-cache git nodejs npm

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install Tailwind CSS and CLI locally (stable v4 versions)
RUN npm install --no-save tailwindcss@^4.0.0 @tailwindcss/cli@^4.0.0 @tailwindcss/typography

# Build Tailwind CSS
RUN npx @tailwindcss/cli -i ./src/input.css -o ./static/tailwind.min.css --minify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

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
