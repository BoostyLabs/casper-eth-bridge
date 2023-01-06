# syntax=docker/dockerfile:experimental
FROM rustlang/rust:nightly as builder
RUN apt-get update && \
    apt-get install -y libudev-dev
WORKDIR /build
COPY ./build ./build
COPY ./poc/bridge-core ./project
RUN chown 666 build/bridge_core.sh && chmod +x build/bridge_core.sh
RUN # \
    --mount=type=secret,id=gitkey \
    ./build/bridge_core.sh

FROM debian:sid
ENV RUST_LOG=info
WORKDIR /
RUN apt-get update && \
    apt-get -y install openssl && \
    apt-get -y install libssl-dev
COPY --from=builder /build/project/target/release/bridge /build/project/target/release/bridge
