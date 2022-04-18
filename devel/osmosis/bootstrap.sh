#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLEANUP=${CLEANUP:-"0"}

OS_PLATFORM=$(uname -s)
OS_ARCH=$(uname -m)

OSMOSIS_PLATFORM=${OSMOSIS_PLATFORM:-"linux_amd64"}
OSMOSIS_VERSION=${OSMOSIS_VERSION:-"v1.0.1"}
OSMOSIS_GENESIS="https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json"
OSMOSIS_GENESIS_HEIGHT=${OSMOSIS_GENESIS_HEIGHT:-"1"}
OSMOSIS_ADDRESS_BOOK="https://quicksync.io/addrbook.osmosis.json"
OSMOSIS_PATH="osmosisd"

case $OS_PLATFORM-$OS_ARCH in
  Darwin-x86_64) OSMOSIS_PLATFORM="darwin_amd64" ;;
  Darwin-arm64)  OSMOSIS_PLATFORM="darwin_arm64" ;;
  Linux-x86_64)  OSMOSIS_PLATFORM="linux_amd64"  ;;
  *) echo "Invalid platform"; exit 1 ;;
esac

if [[ -z $(which "wget" || true) ]]; then
  echo "ERROR: wget is not installed"
  exit 1
fi

if [[ $CLEANUP -eq "1" ]]; then
  echo "Deleting all local data"
  rm -rf ./tmp/
fi

echo "Setting up working directory"
mkdir -p tmp
pushd tmp

echo "Your platform is $OS_PLATFORM/$OS_ARCH"

if [ ! -f "$OSMOSIS_PATH" ]; then
  echo "NOTE: Downloading instrumented binaries is not implemented yet"
  # echo "Downloading osmosisd $OSMOSIS_VERSION binary"
  #wget --quiet -O ./gaiad "https://github.com/figment-networks/osmisis-dm/releases/download/$OSMOSIS_VERSION/gaiad_${OSMOSIS_VERSION}_deepmind_$OSMOSIS_PLATFORM"
  #chmod +x ./gaiad
fi

if [[ -z $(which osmosisd || true) ]]; then
  echo "Please make sure you have installed local osmosisd $OSMOSIS_VERSION binary"
  exit 1
fi

if [ ! -d "osmosis_home" ]; then
  echo "Configuring chain home directory"
  $OSMOSIS_PATH --home=osmosis_home init $(hostname)
  rm -f \
    osmosis_home/config/genesis.json \
    osmosis_home/config/addrbook.json
fi

if [ ! -f "osmosis_home/config/genesis.json" ]; then
  echo "Downloading osmosis genesis file"
  wget --quiet -O osmosis_home/config/genesis.json $OSMOSIS_GENESIS
fi

if [ ! -f "osmosis_home/config/addrbook.json" ]; then
  echo "Downloading address book"
  wget --quiet -O osmosis_home/config/addrbook.json $OSMOSIS_ADDRESS_BOOK
fi

cat << END >> osmosis_home/config/config.toml

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
    - ingestor
    #- merger
    #- firehose
    #- relayer
  flags:
    common-first-streamable-block: $OSMOSIS_GENESIS_HEIGHT
    ingestor-mode: node
    ingestor-node-path: $(which osmosisd)
    ingestor-node-args: start --x-crisis-skip-assert-invariants --home=./osmosis_home
END
fi
