#Builder node
FROM golang:1.13.4-alpine3.10 as builder

# Add Maintainer Info
LABEL maintainer="Dilshat Aliev <dilshat.aliev@gmail.com>"

# Create working directory
RUN mkdir /app
# Copy the source from the current directory to the Working Directory inside the container
ADD . /app
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN GOINSECURE=1 go mod download

# Build sms service
RUN CGO_ENABLED=0 GOOS=linux go build -o smsservice

#Production node
FROM alpine:latest AS production

COPY --from=builder /app/.env .
COPY --from=builder /app/smsservice .

CMD ["./smsservice"]