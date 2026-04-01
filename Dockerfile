# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates build-base

# Set the working directory to the project root
WORKDIR /app

# Enable Go modules
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire source code
# This allows the builder to access all service code
COPY . .

# Define a build argument for the service name
ARG service_name

# Build the specified application
# The binary will be named after the service_name
RUN go build -a -o /app/bin/${service_name} ./cmd/${service_name}

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Define a build argument for the service name (again for this stage)
ARG service_name

# Copy the specific binary from the builder stage
COPY --from=builder /app/bin/${service_name} /app/

# Expose the port the app runs on (default, can be overridden by service)
EXPOSE 8080

# Command to run the application
# The actual command will be set in Kubernetes deployment or docker run command
ENTRYPOINT ["/app/${service_name}"]
