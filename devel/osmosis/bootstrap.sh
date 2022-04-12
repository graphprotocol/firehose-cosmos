#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLEANUP=${CLEANUP:-"0"}

OS_PLATFORM=$(uname -s)
OS_ARCH=$(uname -m)

OSMOSIS_PLATFORM=${OSMOSIS_PLATFORM:-"linux_amd64"}
OSMOSIS_VERSION=${OSMOSIS_VERSION:-"v7.0.0"}
OSMOSIS_GENESIS="https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json"
OSMOSIS_GENESIS_HEIGHT=${OSMOSIS_GENESIS_HEIGHT:-"1"}
OSMOSIS_ADDRESS_BOOK="https://quicksync.io/addrbook.osmosis.json"

case $OS_PLATFORM-$OS_ARCH in
  Darwin-x86_64) GAIA_PLATFORM="darwin_amd64" ;;
  Darwin-arm64)  GAIA_PLATFORM="darwin_arm64" ;;
  Linux-x86_64)  GAIA_PLATFORM="linux_amd64"  ;;
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

if [ ! -f "osmosisd" ]; then
  echo "Downloading osmosisd $OSMOSIS_VERSION binary"
  #wget --quiet -O ./gaiad "https://github.com/figment-networks/gaia-dm/releases/download/$GAIA_VERSION/gaiad_${GAIA_VERSION}_deepmind_$GAIA_PLATFORM"
  #chmod +x ./gaiad
  cp /Users/sosedoff/go/src/github.com/osmosis-labs/osmosis/build/osmosisd .
fi

if [ ! -d "osmosis_home" ]; then
  echo "Configuring chain home directory"
  ./osmosisd --home=osmosis_home init $(hostname)
  rm -f \
    osmosis_home/config/genesis.json \
    osmosis_home/config/addrbook.json
fi

# if [ ! -f "osmosis_home/config/config.toml" ]; then
#   echo "Fetching gaia config"
#   # TODO
# fi

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
    - merger
    - firehose
    - relayer
  flags:
    common-first-streamable-block: $OSMOSIS_GENESIS_HEIGHT
    ingestor-mode: node
    ingestor-node-path: ./osmosisd
    ingestor-node-args: start --x-crisis-skip-assert-invariants --home=./osmosis_home
END
fi
