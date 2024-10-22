FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN GOOS=linux GOARCH=amd64 go build -o main

# Base image to run DinD and Go service
FROM docker:20.10-dind

# Install Go (optional: if your service is not already compiled)
COPY --from=builder /app/main /main

RUN chmod 777 /main

# Expose the Docker Daemon port (optional, if needed)
EXPOSE 2375
EXPOSE 8000

# Run the Docker daemon in the background and then start your Go service
CMD ["sh", "-c", "dockerd-entrypoint.sh & ./main"]