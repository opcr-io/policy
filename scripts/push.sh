#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_SERVER}" ]    && INPUT_SERVER="opcr.io"
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="0"

# validate if values are set
[ -z "${INPUT_TAG}" ]       && echo "INPUT_TAG is not set exiting" && exit 2
[ -z "${INPUT_SERVER}" ]    && exit "INPUT_SERVER is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set exiting" && exit 2

# output all inputs env variables
echo "POLICY-PUSH         $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_TAG           ${INPUT_TAG}"
echo "INPUT_SERVER        ${INPUT_SERVER}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY}"
echo "GITHUB_WORKSPACE    ${GITHUB_WORKSPACE}"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy push ${INPUT_TAG}"

# execute command
eval "$CMD" || e_code=1
printf "\n"

/app/policy images --remote
printf "\n"

exit $e_code
