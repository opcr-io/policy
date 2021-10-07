#!/usr/bin/env bash

sudo rm -rf /tmp/policytest

docker run -ti \
-e INPUT_USERNAME=gertd \
-e INPUT_PASSWORD=${GIT_TOKEN} \
-e INPUT_SERVER= \
-e INPUT_VERBOSITY= \
-v "/tmp/policytest":"/github/workspace" \
--entrypoint=/app/login.sh \
policy:0.0.34-dirty 

