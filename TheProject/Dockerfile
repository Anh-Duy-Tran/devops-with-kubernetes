FROM golang:1.23 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-app

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /usr/src/app

# Copy the binary and template
COPY --from=builder /todo-app ./
COPY index.html ./

# Create directory for image cache (will be mounted as volume)
RUN mkdir -p ./images

# Expose port (default 8080, but configurable via PORT env var)
EXPOSE 8080

# Define volume for persistent image storage
VOLUME ["./images"]

# Run
CMD ["./todo-app"]
