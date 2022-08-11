#!/bin/bash

pushd /home/miguel/utils/ftxgo

source .keys
./ftxgo $*

popd
