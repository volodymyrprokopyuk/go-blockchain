#!/usr/bin/env fish

set -g node localhost:1122
set -g pass password

# ./bcn node start --node $node --bootstrap --authpass $pass \
#   --ownerpass $pass --balance 1000
# ./bcn account create --node $node --ownerpass $pass
# ./bcn node start --node localhost:1123 --seed localhost:1122

function txSignAndSend -a node pass from to value
  set -l tx (./bcn tx sign --node $node --from $from --to $to --value $value \
    --ownerpass $pass)
  echo $tx
  set -l hash (./bcn tx send --node $node --sigtx $tx)
  echo $hash
end

set -l own 23309c0c52fe0bef535ddd439fa6ffe63363337d92f530a84137f752a524a4e7
set -l ben 2beebae29ca76b2c70afe5b981daa95bf325ee6324671203246258cbb5d0f57f

txSignAndSend $node $pass $own $ben 2
txSignAndSend $node $pass $ben $own 1

# ./bcn account balance --node $node --account $own
# ./bcn account balance --node $node --account $ben

# ./bcn block search --node $node --number 2
# ./bcn block search --node $node --hash 96b3d1d
# ./bcn block search --node $node --parent 76088e0

# ./bcn tx search --hash 6fe5fff
# ./bcn tx search --from d8a05ac
# ./bcn tx search --to fd29d48
# ./bcn tx search --account d8a05ac
