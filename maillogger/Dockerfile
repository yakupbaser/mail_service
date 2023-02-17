FROM golang:alpine AS builder

LABEL MAINTAINER "John Doe <john.doe@example.com>"

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create app directory
WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .

# Download all the dependencies that are required
RUN go mod download

# Copy all other source code to work directory
COPY . .

# Build the application
RUN go build -o maillogger maillogger.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/maillogger .

CMD ["./maillogger"]
