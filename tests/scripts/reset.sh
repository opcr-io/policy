#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export POLICY_TEST=/tmp/policy
export POLICY_CONFIG=/tmp/policy/config.json
export POLICY_FILE_STORE_ROOT=/tmp/policy
export TRACE=1

# cleanup prev policy store and config
rm -rf $POLICY_FILE_STORE_ROOT
rm -rf $POLICY_TEST

mkdir -p $POLICY_TEST
mkdir -p $POLICY_FILE_STORE_ROOT

touch $POLICY_CONFIG

./policy version

./policy images
