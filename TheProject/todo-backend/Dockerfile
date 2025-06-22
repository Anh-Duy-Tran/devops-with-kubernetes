FROM golang:1.23 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-backend

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /usr/src/app

# Copy the binary
COPY --from=builder /todo-backend ./

# Expose port (default 3001, but configurable via PORT env var)
EXPOSE 3001

# Run
CMD ["./todo-backend"] 