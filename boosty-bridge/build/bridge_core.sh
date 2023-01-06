#!/bin/sh
set -e

export GIT_SSH_COMMAND="ssh -i /run/secrets/gitkey"
export CARGO_NET_GIT_FETCH_WITH_CLI=true

mkdir -p ~/.ssh
touch ~/.ssh/known_hosts
ssh-keyscan github.com >> ~/.ssh/known_hosts

cd /build/project

cargo build --release --bin bridge
cp /build/project/target/release/bridge /
