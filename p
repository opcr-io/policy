#!/usr/bin/env bash

set -e  # Exit on error

echo "$*"

$PWD/dist/build_darwin_arm64/policy $*