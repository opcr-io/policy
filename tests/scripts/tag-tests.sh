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

# test 1 - delete in reverse order of creation
./policy build ./tests/fixtures/policy_v1    --tag test/policy_v1:v0.0.1
./policy images

./policy tag docker.io/test/policy_v1:v0.0.1 docker.io/test/policy_v1:test
./policy tag docker.io/test/policy_v1:v0.0.1 docker.io/test/policy_v1:latest
./policy images

./policy save docker.io/test/policy_v1:v0.0.1
./policy save docker.io/test/policy_v1:test
./policy save docker.io/test/policy_v1:latest

# delete in reverse order of creation
./policy rm docker.io/test/policy_v1:latest --force
./policy images
./policy save docker.io/test/policy_v1:v0.0.1
./policy save docker.io/test/policy_v1:test

./policy rm docker.io/test/policy_v1:test   --force
./policy images
./policy save docker.io/test/policy_v1:v0.0.1

./policy rm docker.io/test/policy_v1:v0.0.1 --force
./policy images

# test 2 - delete in order of creation 
./policy build ./tests/fixtures/policy_v1    --tag test/policy_v1:v0.0.1
./policy images

./policy tag docker.io/test/policy_v1:v0.0.1 docker.io/test/policy_v1:test
./policy tag docker.io/test/policy_v1:v0.0.1 docker.io/test/policy_v1:latest
./policy images

./policy save docker.io/test/policy_v1:v0.0.1
./policy save docker.io/test/policy_v1:test
./policy save docker.io/test/policy_v1:latest

./policy rm docker.io/test/policy_v1:v0.0.1 --force
./policy images
./policy save docker.io/test/policy_v1:test
./policy save docker.io/test/policy_v1:latest

./policy rm docker.io/test/policy_v1:test   --force
./policy images
./policy save docker.io/test/policy_v1:latest

./policy rm docker.io/test/policy_v1:latest --force
./policy images
