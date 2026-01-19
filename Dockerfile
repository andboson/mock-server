# syntax = docker/dockerfile:1.2.1

# Start from the latest golang base image
FROM public.ecr.aws/docker/library/golang:1.25.5-bookworm@sha256:019c22232e57fda8ded2b10a8f201989e839f3d3f962d4931375069bbb927e03 as builder


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
FROM public.ecr.aws/bitnami/minideb:bullseye@sha256:a6f3a96622f1a2e0eb049e08e3aa4db29c725318c7f410a050f333e074793e88

RUN install_packages ca-certificates && \
    update-ca-certificates

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/ .

# ensure that we'll not be running the container as root
USER 1001:1001

# Command to run the executable
ENTRYPOINT ["./main"]
