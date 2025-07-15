FROM alpine:3.18

# Install required packages
RUN apk add --no-cache \
    curl \
    bash \
    ca-certificates

# Create app directory
WORKDIR /app

# Copy the script
COPY create-todo.sh /app/create-todo.sh

# Make script executable
RUN chmod +x /app/create-todo.sh

# Run the script
CMD ["/app/create-todo.sh"] 