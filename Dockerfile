# Use the official Golang image
FROM golang:1.20

# Set the working directory
WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Expose the API port
EXPOSE 8081

# Run the application
CMD ["go", "run", "main.go"]
