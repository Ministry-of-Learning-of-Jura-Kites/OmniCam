FROM alpine:latest
COPY . /app
WORKDIR /app

# Recursively list all files in the current work directory during build
CMD ["ls", "-R"]

