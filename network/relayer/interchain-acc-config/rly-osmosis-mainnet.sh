#!/bin/bash

# Configure predefined mnemonic pharses
BINARY=rly
CHAIN_DIR=./data
RELAYER_DIR=./relayer
MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
MNEMONIC_2="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"

# Ensure rly is installed
if ! [ -x "$(command -v $BINARY)" ]; then
    echo "$BINARY is required to run this script..."
    echo "You can download at https://github.com/cosmos/relayer"
    exit 1
fi

echo "Initializing $BINARY..."
$BINARY config init --home $CHAIN_DIR/$RELAYER_DIR

echo "Adding configurations for both chains..."
$BINARY chains add -f $PWD/network/relayer/interchain-acc-config/chains/test-1.json --home $CHAIN_DIR/$RELAYER_DIR
$BINARY chains add -f $PWD/network/relayer/interchain-acc-config/chains/osmosis-1.json --home $CHAIN_DIR/$RELAYER_DIR
echo "Adding configurations for paths..."
#$BINARY paths add-dir $PWD/network/relayer/interchain-acc-config/paths --home $CHAIN_DIR/$RELAYER_DIR
$BINARY paths add test-1 osmosis-1 test1-osmo1 -f $PWD/network/relayer/interchain-acc-config/paths/test1-osmosis-mainnet.json --home $CHAIN_DIR/$RELAYER_DIR

echo "Restoring accounts..."
$BINARY keys restore test-1 test-1 "$MNEMONIC_1" --home $CHAIN_DIR/$RELAYER_DIR
$BINARY keys restore osmosis-1 test-1 "$MNEMONIC_1" --home $CHAIN_DIR/$RELAYER_DIR


echo "Linking both chains..."
$BINARY tx link test1-osmo1--home $CHAIN_DIR/$RELAYER_DIR --debug
#
#echo "Starting to listen relayer..."
$BINARY start test1-osmo1 --home $CHAIN_DIR/$RELAYER_DIR