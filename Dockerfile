# Stage 1: Build the Go application
FROM golang:1.22 AS builder

WORKDIR /app

# Copy the Go application files
COPY . .

# Install dependencies
RUN go mod tidy

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

# Stage 2: Docker with Go installed
FROM docker:20.10-dind

# Install Go in the final image
ENV GOLANG_VERSION 1.22.8
RUN apk add --no-cache curl git bash && \
    curl -OL https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# Add Go binary to PATH
ENV PATH="/usr/local/go/bin:$PATH"

# Copy the compiled Go binary from the build stage
COPY --from=builder /app/main /main

# Set permissions for the Go binary
RUN chmod 777 /main

# Expose the Docker Daemon port (optional, if needed)
EXPOSE 2375
EXPOSE 8000

# Run Docker Daemon and Go service concurrently
CMD ["sh", "-c", "dockerd-entrypoint.sh & ./main"]
