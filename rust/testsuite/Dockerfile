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
RUN cargo test --release --bin kurtosis-rust-example
RUN cargo build --release --bin kurtosis-rust-example

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

# Copy the code into the container
COPY --from=builder /home/rust/target/x86_64-unknown-linux-musl/release/kurtosis-rust-example ./

# TODO Switch to exec command form, wrapping arguments with double-quote
CMD RUST_BACKTRACE=1 ./kurtosis-rust-example \
    --custom-params-json="${CUSTOM_PARAMS_JSON}" \
    --kurtosis-api-socket="${KURTOSIS_API_SOCKET}" \
    --log-level="${LOG_LEVEL}"
