#!/usr/bin/env bash

[ -z "$1" ] && echo "provide container version" && exit 1

CONTAINER=policy:$1

# reset
sudo rm -rf /tmp/policytest

# init
mkdir -p /tmp/policytest
gh repo clone aserto-dev/policy-peoplefinder /tmp/policytest/peoplefinder

docker run -ti \
-e INPUT_USERNAME=gertd \
-e INPUT_PASSWORD=${GIT_TOKEN} \
-e INPUT_SERVER= \
-e INPUT_VERBOSITY= \
-e GITHUB_WORKSPACE=/github/workspace \
-v "/tmp/policytest":"/github/workspace" \
--entrypoint=/app/login.sh \
${CONTAINER}

docker run -ti \
-e INPUT_SRC=peoplefinder/src \
-e INPUT_TAG=peoplefinder:$(sver -n patch) \
-e INPUT_REVISION=$(sver) \
-e INPUT_SERVER= \
-e INPUT_VERBOSITY= \
-e GITHUB_WORKSPACE=/github/workspace \
-v "/tmp/policytest":"/github/workspace" \
--entrypoint=/app/build.sh \
${CONTAINER}

docker run -ti \
-e INPUT_TAG=peoplefinder:$(sver -n patch) \
-e INPUT_SERVER= \
-e INPUT_VERBOSITY= \
-e GITHUB_WORKSPACE=/github/workspace \
-v "/tmp/policytest":"/github/workspace" \
--entrypoint=/app/push.sh \
${CONTAINER}
