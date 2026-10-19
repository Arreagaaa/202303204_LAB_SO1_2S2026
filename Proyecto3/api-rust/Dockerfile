FROM rust:1.99-bookworm AS build
WORKDIR /src
COPY api-rust/Cargo.toml ./api-rust/Cargo.toml
COPY api-rust/Cargo.lock ./api-rust/Cargo.lock
COPY api-rust/src ./api-rust/src
WORKDIR /src/api-rust
RUN cargo build --release --locked

FROM debian:bookworm-slim
COPY --from=build /src/api-rust/target/release/api-rust /usr/local/bin/api-rust
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/api-rust"]
