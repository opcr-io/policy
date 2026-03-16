#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

./policy version

./policy images

./policy build ./tests/policy_v0   --rego-version=rego.v0   --tag test/policy_v0:test
./policy build ./tests/policy_v0v1 --rego-version=rego.v0v1 --tag test/policy_v0v1:test
./policy build ./tests/policy_v1   --rego-version=rego.v1   --tag test/policy_v1:test

./policy inspect test/policy_v0:test
./policy inspect test/policy_v0v1:test
./policy inspect test/policy_v1:test

mkdir -p /tmp/policy

./policy save test/policy_v0:test   --file /tmp/policy/policy_v0.bundle.tar.gz
./policy save test/policy_v0v1:test --file /tmp/policy/policy_v0v1.bundle.tar.gz
./policy save test/policy_v1:test   --file /tmp/policy/policy_v1.bundle.tar.gz

./policy rm test/policy_v0:test   --force 
./policy rm test/policy_v0v1:test --force
./policy rm test/policy_v1:test   --force

./policy images
