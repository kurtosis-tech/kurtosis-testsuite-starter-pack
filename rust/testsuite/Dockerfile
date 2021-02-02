FROM rust:1.49-alpine3.12 AS builder

WORKDIR /build

# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

RUN go test ./...

# Build the application
RUN GOOS=linux rustc -o testsuite.bin testsuite/main.go

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/testsuite.bin .

# TODO Switch to exec command form, wrapping arguments with double-quote
CMD ./testsuite.bin \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"
