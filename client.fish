#!/usr/bin/env fish

set -l pass password

function chainInit
  rm -rf .blockstore/ .keystore/
  set -l own (./bcn store init --password $pass --balance 1000)
  echo owner: $own
  set -l ben (./bcn account create --password $pass)
  echo benef: $ben
end

function txSignAndSend -a pass from to value
  set -l stx (./bcn tx sign --password $pass --from $from --to $to --value $value)
  echo $stx
  set -l hash (./bcn tx send --sigtx $stx)
  echo $hash
end

set -l own 3351fcb0bdc66f3e53d0a2f8e768b9351849b64a60c43589f3dcab0807af363f
set -l ben f34ec96f232e6d3f0ba0a998a7e81cea3b01463cfd772c9774e980e8f771e70f

txSignAndSend $pass $own $ben 12
txSignAndSend $pass $own $ben 34
