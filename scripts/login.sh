#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_SERVER}" ]    && INPUT_SERVER="ghcr.io"
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="error"

# validate if values are set
[ -z "${INPUT_USERNAME}" ]  && echo "INPUT_USERNAME is not set exiting" && exit 2
[ -z "${INPUT_PASSWORD}" ]  && exit "INPUT_PASSWORD is not set exiting" && exit 2
[ -z "${INPUT_SERVER}" ]    && exit "INPUT_SERVER is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set exiting" && exit 2

VERBOSITY=0
case ${INPUT_VERBOSITY} in
  "info")
    VERBOSITY=1
    ;;
  "error")
    VERBOSITY=0
    ;;
  "debug")
    VERBOSITY=2
    ;;
  "trace")
    VERBOSITY=3
    ;;
esac

# output all inputs env variables
echo "POLICY-LOGIN        $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_USERNAME      ${INPUT_USERNAME}"
echo "INPUT_PASSWORD      **********"
echo "INPUT_SERVER        ${INPUT_SERVER}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY} (${VERBOSITY})"
printf "\n"

#
# start execution block
#
e_code=0

# execute command
echo ${INPUT_PASSWORD} | /app/policy login --username=${INPUT_USERNAME} --password-stdin --server=${INPUT_SERVER} --verbosity=${VERBOSITY} --default-domain
e_code=$?

printf "\n"

exit $e_code
