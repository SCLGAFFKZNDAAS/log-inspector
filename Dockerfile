FROM golang:1.25.7

# Set destination for COPY
WORKDIR /log-inspector

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o main

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
# EXPOSE 8080

# Run
CMD ["./main"]