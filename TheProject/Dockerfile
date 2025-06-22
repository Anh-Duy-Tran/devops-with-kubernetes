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

FROM scratch

COPY --from=builder /todo-app /
# Copy the HTML template file
COPY index.html /

# Expose port (default 8080, but configurable via PORT env var)
EXPOSE 8080

# Run
CMD ["/todo-app"]
