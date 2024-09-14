# Stage 1: Build the Go binary
FROM golang:1.20-alpine AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go binary
RUN go build -o /app/main .

# Stage 2: Run the Go binary
FROM alpine:latest

# Set environment variables
ENV GIN_MODE=release

# Set the working directory in the final container
WORKDIR /app

# Copy the binary from the builder container
COPY --from=builder /app/main .

# Expose the port the service will run on
EXPOSE 8080

# Command to run the binary
CMD ["./main"]