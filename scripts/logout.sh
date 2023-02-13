#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_SERVER}" ]    && INPUT_SERVER="ghcr.io"
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="error"

# validate if values are set
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
echo "POLICY-LOGOUT       $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_SERVER        ${INPUT_SERVER}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY} (${VERBOSITY})"
printf "\n"

#
# start execution block
#
e_code=0

# execute command
/app/policy logout --server=${INPUT_SERVER} --verbosity=${VERBOSITY}
e_code=$?

printf "\n"

exit $e_code
