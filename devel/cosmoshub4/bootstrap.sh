#!/usr/bin/env bash

set -o errexit
set -o pipefail

CLEANUP=${CLEANUP:-"0"}

OS_PLATFORM=$(uname -s)
OS_ARCH=$(uname -m)

GAIA_PLATFORM=${GAIA_PLATFORM:-"linux_amd64"}
GAIA_VERSION=${GAIA_VERSION:-"v4.2.1"}
GAIA_GENESIS="https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz"
GAIA_GENESIS_HEIGHT=${GAIA_GENESIS_HEIGHT:-"5200791"}
GAIA_ADDRESS_BOOK="https://quicksync.io/addrbook.cosmos.json"

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

if [ ! -f "gaiad" ]; then
  echo "Downloading gaiad $GAIA_VERSION binary"
  wget -O ./gaiad "https://github.com/figment-networks/gaia-dm/releases/download/$GAIA_VERSION/gaiad_${GAIA_VERSION}_firehose_$GAIA_PLATFORM"
  chmod +x ./gaiad
fi

if [ ! -d "gaia_home" ]; then
  echo "Configuring gaia home directory"
  ./gaiad --home=gaia_home init $(hostname) 2> /dev/null
  rm -f \
    gaia_home/config/genesis.json \
    gaia_home/config/addrbook.json
fi

if [ ! -f "gaia_home/config/config.toml" ]; then
  echo "Fetching gaia config"
  # TODO
fi

if [ ! -f "gaia_home/config/genesis.json" ]; then
  echo "Downloading gaia genesis file"
  wget --quiet -O gaia_home/config/genesis.json.gz $GAIA_GENESIS
  gunzip gaia_home/config/genesis.json.gz
fi

if [ ! -f "gaia_home/config/addrbook.json" ]; then
  echo "Downloading address book"
  wget --quiet -O gaia_home/config/addrbook.json $GAIA_ADDRESS_BOOK
fi

cat << END >> gaia_home/config/config.toml

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
    common-first-streamable-block: $GAIA_GENESIS_HEIGHT
    ingestor-mode: node
    ingestor-node-path: ./gaiad
    ingestor-node-args: start --x-crisis-skip-assert-invariants --home=./gaia_home
    ingestor-node-logs-filter: "module=(p2p|pex|consensus|x/bank)"
    firehose-real-time-tolerance: 99999h
    relayer-max-source-latency: 99999h
    verbose: 1
END
fi
