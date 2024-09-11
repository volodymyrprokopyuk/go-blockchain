#!/usr/bin/env fish

set -l pass password

# rm -rf .blockstore/ .keystore/
# set -l own (./chain store init -p $pass --balance 1000)
# echo owner: $own
set -l ben (./bcn account create --password $pass)
echo benef: $ben

# set -l own 86dcda3939811bda77aa2dd21e4ea255489ed49fd89c8142b18d5f09eb9095dd
# set -l ben 6aaee726c378c47e668e91d33a0722a18b6c1afbd3b3dc750f2d67e5c2ecc6a3

# set -l stx (./chain tx sign -p $pass --from $own --to $ben --value 12)
# echo $stx
# set -l hash (./chain tx send --sigtx $stx)
# echo $hash
