# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory in the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o app ./cmd/api

# Expose the port your application listens on
EXPOSE 3000

# Command to run your application
CMD ["./app"]
