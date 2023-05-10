#!/usr/bin/env bash

set -o errexit
set -o pipefail

CLEANUP=${CLEANUP:-"0"}
CHAIN_ID=${CHAIN_ID:-"coreum-mainnet-1"}
OS_PLATFORM=$(uname -s)
OS_ARCH=$(uname -m)

COREUM_VERSION=${COREUM_VERSION:-"v1.0.0"}
COREUM_GENESIS_HEIGHT=${COREUM_GENESIS_HEIGHT:-"1"}

if [[ -z $(which "wget" || true) ]]; then
  echo "ERROR: wget is not installed"
  exit 1
fi

if [[ $CLEANUP -eq "1" ]]; then
  echo "Deleting all local data"
  rm -rf ./tmp/ > /dev/null
fi

echo "Setting up working directory"
mkdir -p tmp
pushd tmp

echo "Your platform is $OS_PLATFORM-$OS_ARCH"

if [ ! -f "cored" ]; then
  case $OS_PLATFORM-$OS_ARCH in
    Linux-x86_64)  COREUM_PLATFORM="linux-amd64"  ;;
    Linux-aarch64)  COREUM_PLATFORM="linux-arm64"  ;;
    *) echo "Invalid platform"; exit 1 ;;
  esac

  echo "Downloading cored $COREUM_VERSION binary"
  wget --quiet -O ./cored "https://github.com/CoreumFoundation/coreum/releases/download/$COREUM_VERSION/cored-$COREUM_PLATFORM"
  chmod +x ./cored
fi

if [ ! -d "coreum_home" ]; then
  echo "Configuring home directory"
  ./cored --home=coreum_home init $(hostname) --chain-id=$CHAIN_ID 2> /dev/null
fi

cat << END >> coreum_home/$CHAIN_ID/config/config.toml

#######################################################
###       Extractor Configuration Options     ###
#######################################################
[extractor]
enabled = true
output_file = "stdout"
END

if [ ! -f "firehose.yml" ]; then
  cat << END >> firehose.yml
start:
  args:
    - reader
    - merger
    - firehose
  flags:
    common-first-streamable-block: $COREUM_GENESIS_HEIGHT
    common-live-blocks-addr:
    reader-mode: node
    reader-node-path: ./cored
    reader-node-args: start --x-crisis-skip-assert-invariants --home=./coreum_home --chain-id=$CHAIN_ID
    reader-node-logs-filter: "module=(p2p|pex|consensus|x/bank)"
    relayer-max-source-latency: 99999h
    verbose: 1
END
fi
