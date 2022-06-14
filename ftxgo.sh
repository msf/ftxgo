#!/bin/bash

pushd /home/miguel/utils/ftxgo

source .keys
./ftxgo -budget=53 -avg_window=30 $*

popd
