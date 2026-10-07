FROM rust:1.82-bookworm AS build
WORKDIR /src
COPY api-rust/Cargo.toml ./api-rust/Cargo.toml
COPY api-rust/src ./api-rust/src
WORKDIR /src/api-rust
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=build /src/api-rust/target/release/api-rust /usr/local/bin/api-rust
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/api-rust"]
