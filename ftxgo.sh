#!/bin/bash

pushd /home/miguel/utils/ftxgo

source .keys
./ftxgo -avg_window=17 $*

popd
