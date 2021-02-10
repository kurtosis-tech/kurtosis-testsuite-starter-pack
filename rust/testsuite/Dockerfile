# This entire Dockerfile inspired by https://github.com/LukeMathWalker/cargo-chef, to cache
# Rust dependencies
ARG CARGO_CHEF_IMAGE=lukemathwalker/cargo-chef

FROM ${CARGO_CHEF_IMAGE} as planner
WORKDIR /app
COPY . .
RUN cargo chef prepare  --recipe-path recipe.json

FROM ${CARGO_CHEF_IMAGE} as cacher
WORKDIR /app
COPY --from=planner /app/recipe.json recipe.json
RUN cargo chef cook --release --recipe-path recipe.json

FROM rust:1.49 as builder
WORKDIR /app
COPY . .
# Copy over the cached dependencies
COPY --from=cacher /app/target target
COPY --from=cacher /usr/local/cargo /usr/local/cargo
RUN cargo test --release
RUN cargo build --release --bin kurtosis_rust_example

FROM alpine:3.12 as runtime
WORKDIR /app
COPY --from=builder /app/target/release/kurtosis_rust_example ./
# TODO Switch to exec command form, wrapping arguments with double-quote
# TODO Use RUST_LOG envvar to configure logging
CMD ./kurtosis_rust_example \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"




# # FROM rust:1.49-alpine3.12 AS builder
# FROM ekidd/rust-musl-builder:stable AS builder
# 
# ENV TARGET=x86_64-unknown-linux-musl
# 
# # TODO Copy dependencies in the same way we copy go.mod and do mod download???
# 
# # Copy the code into the container
# COPY . .
# 
# RUN cargo test
# 
# RUN cargo install --path .
# 
# # ============= Execution Stage ================
# FROM alpine:3.12 AS execution
# 
# RUN apt-get update && apt-get install -y extra-runtime-dependencies && rm -rf /var/lib/apt/lists/*
# 
# WORKDIR /run
# 
# # Copy the code into the container
# COPY --from=builder /usr/local/cargo/bin/kurtosis_rust_example ./
# 
# # TODO Switch to exec command form, wrapping arguments with double-quote
# # TODO Use RUST_LOG envvar to configure logging
# CMD ./kurtosis_rust_example \
#     --custom-params-json="${CUSTOM_PARAMS_JSON}" \
#     --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
#     --log-level="${LOG_LEVEL}"
# 