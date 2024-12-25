#!/bin/bash

set -e

pushd $( dirname "${BASH_SOURCE[0]}")

hostname=$(hostname)

ip=$(grep "${hostname}" ./external_ip.txt | awk '{print $2}')

if [ -z "${ip}" ]; then
  echo "No external ip found for ${hostname}"
  exit 1
fi

echo "Replacing external ip with ${ip} for host ${hostname}"

if [ "$(uname)" == "Darwin" ]; then
  sed -i '' "s/<node_ip>/${ip}/g" ./config.yaml
else
  sed -i "s/<node_ip>/${ip}/g" ./config.yaml
fi

popd