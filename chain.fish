#!/usr/bin/env fish

set -l pass password

rm -rf .blockstore/ .keystore/

set -l own (./chain store init -p $pass --balance 1000)
echo owner: $own
set -l ben (./chain account create -p $pass)
echo benef: $ben
