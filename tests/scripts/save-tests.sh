#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export POLICY_TEST=/tmp/policy
export POLICY_CONFIG=/tmp/policy/config.json
export POLICY_FILE_STORE_ROOT=/tmp/policy

# cleanup prev policy store and config
rm -rf $POLICY_FILE_STORE_ROOT
rm -rf $POLICY_TEST

mkdir -p $POLICY_TEST
mkdir -p $POLICY_FILE_STORE_ROOT

touch $POLICY_CONFIG

./policy version

./policy images

./policy build ./tests/fixtures/policy_v0   --rego-version=rego.v0   --tag test/policy_v0:test
./policy build ./tests/fixtures/policy_v0v1 --rego-version=rego.v0v1 --tag test/policy_v0v1:test
./policy build ./tests/fixtures/policy_v1   --rego-version=rego.v1   --tag test/policy_v1:test

./policy inspect test/policy_v0:test
./policy inspect test/policy_v0v1:test
./policy inspect test/policy_v1:test


./policy images

./policy save test/policy_v0:test   --file ${POLICY_TEST}/policy_v0.bundle.tar.gz
./policy save test/policy_v0v1:test --file ${POLICY_TEST}/policy_v0v1.bundle.tar.gz
./policy save test/policy_v1:test   --file ${POLICY_TEST}/policy_v1.bundle.tar.gz

./policy images

./policy rm test/policy_v0:test   --force
./policy rm test/policy_v0v1:test --force
./policy rm test/policy_v1:test   --force

./policy images
