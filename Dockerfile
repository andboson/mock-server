# syntax = docker/dockerfile:1.2.1

# Start from the latest golang base image
FROM public.ecr.aws/docker/library/golang:1.27.4-bookworm@sha256:4fa7c968ba93f1635670522dbb01a2c78150443e1385746924566079f36570ea as builder


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the Working Directory inside the container
COPY . .

WORKDIR /app

ARG VERSION=dev
ARG REVISION=unknown
ENV VERSION ${VERSION}
ENV REVISION ${REVISION}

# Build the Go app
RUN  make build

######## Start a new stage from scratch #######
FROM gcr.io/distroless/static-debian12@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/ .

# ensure that we'll not be running the container as root
USER 1001:1001

# Command to run the executable
ENTRYPOINT ["./main"]
