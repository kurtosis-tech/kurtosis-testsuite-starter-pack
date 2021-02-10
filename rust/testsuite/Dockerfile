FROM rust:1.49-alpine3.12 AS builder

WORKDIR /build

# TODO Copy dependencies in the same way we copy go.mod and do mod download???

# Copy the code into the container
COPY . .

# TODO Rust tests

# Build the application
RUN rustc -o testsuite.bin testsuite/main.rs

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/testsuite.bin .

# TODO Switch to exec command form, wrapping arguments with double-quote
# TODO Use RUST_LOG envvar to configure logging
CMD ./testsuite.bin \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"
