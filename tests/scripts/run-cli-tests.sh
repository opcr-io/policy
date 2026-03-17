#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

./policy version

./policy images

./policy build ./tests/fixtures/policy_v0   --rego-version=rego.v0   --tag test/policy_v0:test
./policy build ./tests/fixtures/policy_v0v1 --rego-version=rego.v0v1 --tag test/policy_v0v1:test
./policy build ./tests/fixtures/policy_v1   --rego-version=rego.v1   --tag test/policy_v1:test

./policy inspect test/policy_v0:test
./policy inspect test/policy_v0v1:test
./policy inspect test/policy_v1:test

POLICY_TEST=${TMPDIR}/policy
mkdir -p $POLICY_TEST

./policy images

# ./policy save test/policy_v0:test   --file ${POLICY_TEST}/policy/policy_v0.bundle.tar.gz
# ./policy save test/policy_v0v1:test --file ${POLICY_TEST}/policy/policy_v0v1.bundle.tar.gz
# ./policy save test/policy_v1:test   --file ${POLICY_TEST}/policy/policy_v1.bundle.tar.gz

./policy images

# ./policy rm test/policy_v0:test   --force 
# ./policy rm test/policy_v0v1:test --force
# ./policy rm test/policy_v1:test   --force

./policy images
