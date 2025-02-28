# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with static linking and combined environment variables
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Create the final, minimal image
FROM alpine:latest

# Add a non-root user
RUN adduser -D appuser

# Set the working directory for the application
WORKDIR /usr/src/app

# Copy the static binary from the builder stage
COPY --from=builder /app/main /usr/src/app/
#COPY --from=builder /app/.env /usr/src/app/

# Set the user to non-root `appuser`
USER appuser

# Command to run the executable
CMD ["/usr/src/app/main"]
