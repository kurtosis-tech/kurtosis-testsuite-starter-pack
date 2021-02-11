# # This entire Dockerfile inspired by https://github.com/LukeMathWalker/cargo-chef, to cache
# # Rust dependencies
# ARG RUST_IMAGE=ekidd/rust-musl-builder:stable
# ARG CHEF_VERSION=0.1.14
# 
# FROM ${RUST_IMAGE} as planner
# ARG CHEF_VERSION
# WORKDIR /app
# # We only pay the installation cost once, 
# # it will be cached from the second build onwards
# RUN cargo install cargo-chef --version ${CHEF_VERSION}
# COPY . .
# # The 'chown' is necessary due to https://github.com/emk/rust-musl-builder/issues/26
# RUN sudo chmod -R a+rwX .
# RUN cargo chef prepare  --recipe-path recipe.json
# 
# FROM ${RUST_IMAGE} as cacher
# ARG CHEF_VERSION
# WORKDIR /app
# RUN cargo install cargo-chef --version ${CHEF_VERSION}
# COPY --from=planner /app/recipe.json recipe.json
# # The 'chown' is necessary due to https://github.com/emk/rust-musl-builder/issues/26
# RUN sudo chmod -R a+rwX .
# RUN cargo chef cook --release --recipe-path recipe.json
# 
# FROM ${RUST_IMAGE} as builder
# WORKDIR /app
# ENV TARGET=x86_64-unknown-linux-musl
# COPY . .
# COPY --from=cacher /app/target target
# COPY --from=cacher /usr/local/cargo /usr/local/cargo
# # The 'chown' is necessary due to https://github.com/emk/rust-musl-builder/issues/26
# RUN sudo chmod -R a+rwX .
# RUN cargo test --release
# RUN cargo build --release --target ${TARGET} --bin kurtosis_rust_example
# 
# FROM alpine:3.12 as runtime
# WORKDIR /app
# COPY --from=builder /app/target/release/kurtosis_rust_example .
# # TODO Switch to exec command form, wrapping arguments with double-quote
# # TODO Use RUST_LOG envvar to configure logging
# CMD ./kurtosis_rust_example \
#     --custom-params-json="${CUSTOM_PARAMS_JSON}" \
#     --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
#     --log-level="${LOG_LEVEL}"


# ARG TARGET=x86_64-unknown-linux-musl

# FROM rust:1.49-alpine3.12 AS builder
FROM ekidd/rust-musl-builder:stable AS builder
WORKDIR /home/rust

# Rust doesn't yet have a way to cache build dependencies as Docker layers, so we use this hack so 
# that building doesn't take 5 minutes every time
# The 'chown' is necessary due to https://github.com/emk/rust-musl-builder/issues/26
COPY --chown=rust:rust Cargo.toml ./
COPY --chown=rust:rust Cargo.lock ./
COPY --chown=rust:rust lib/Cargo.toml lib/
COPY --chown=rust:rust lib/dummy_lib_for_build_cache.rs lib/src/lib.rs
COPY --chown=rust:rust testsuite/Cargo.toml testsuite/
COPY --chown=rust:rust testsuite/dummy_main_for_build_cache.rs testsuite/src/main.rs
RUN cargo build --release

# Copy & build real lib code first, so Docker can do caching
COPY --chown=rust:rust lib lib
RUN cargo test --release --lib
RUN cargo build --release --lib

# Next copy the example testsuite code, so that if we make changes there we don't need to rebuild the lib
COPY --chown=rust:rust testsuite testsuite
RUN cargo test --release --bin kurtosis_rust_example
RUN cargo build --release --bin kurtosis_rust_example

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

# Copy the code into the container
COPY --from=builder /home/rust/target/x86_64-unknown-linux-musl/release/kurtosis_rust_example ./

# TODO Switch to exec command form, wrapping arguments with double-quote
# TODO Use RUST_LOG envvar to configure logging
CMD RUST_BACKTRACE=1 ./kurtosis_rust_example \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"
