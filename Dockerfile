# This dockerfile is used by workflow to build the docker image
# It is not used to build the binary from scratch
FROM alpine:latest

WORKDIR /app

COPY release/gama-linux-amd64 /app/gama

# Set environment variable for color output
ENV TERM xterm-256color

ENTRYPOINT ["/app/gama"]
