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

P=./tests/fixtures/policy_

declare -a policies=("v1" "v0v1" "v0")

declare -a variations=("default" "docker.io/default" "docker.io/library/default" "docker.io/library/default:latest")

for p in "${policies[@]}"
do
  
  for v in "${variations[@]}"
  do
    echo "policy $P$p - variation $v"
   
    # build
    ./policy build $P$p --rego-version=rego.$p

    # list images
    ./policy images

    # inspect
    ./policy inspect $v

    # save
    ./policy save $v

    # rm
    ./policy rm $v --force

    # list images
    ./policy images

    # cleanup save output
    if [ -f "./bundle.tar.gz" ]; then
      rm -f ./bundle.tar.gz
    fi
    
  done

done
