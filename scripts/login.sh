#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_SERVER}" ]    && INPUT_SERVER="opcr.io"
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="0"

# validate if values are set
[ -z "${INPUT_USERNAME}" ]  && echo "INPUT_USERNAME is not set exiting" && exit 2
[ -z "${INPUT_PASSWORD}" ]  && exit "INPUT_PASSWORD is not set exiting" && exit 2
[ -z "${INPUT_SERVER}" ]    && exit "INPUT_SERVER is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set exiting" && exit 2

# output all inputs env variables
echo "POLICY-LOGIN        $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "GITHUB_WORKSPACE    ${GITHUB_WORKSPACE}"
echo "INPUT_SERVER        ${INPUT_SERVER}"
echo "INPUT_USERNAME      ${INPUT_USERNAME}"
echo "INPUT_PASSWORD      **********"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY}"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy login --username=${INPUT_USERNAME} --password=${INPUT_PASSWORD} --server=${INPUT_SERVER} --verbosity=${INPUT_VERBOSITY}"

# execute command
eval "$CMD" || e_code=1

exit $e_code
