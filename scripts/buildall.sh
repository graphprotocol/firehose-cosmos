#!/usr/bin/env bash

NAME=$(basename $(pwd))
TARGETS="linux_amd64 linux_arm64 darwin_amd64 darwin_arm64"

for target in $TARGETS; do
  echo "Building for $target"

  parts=(${target//_/ })
  GOOS=${parts[0]} GOARCH=${parts[1]} go build -o dist/${NAME}_$target ./cmd/firehose-tendermint
done
