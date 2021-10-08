#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="error"

# validate if values are set
[ -z "${INPUT_TAG}" ]       && exit "INPUT_TAG is not set, exiting" && exit 2
[ -z "${INPUT_FILE}" ]      && echo "INPUT_FILE is not set, exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set, exiting" && exit 2

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
echo "POLICY-SAVE         $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_TAG           ${INPUT_TAG}"
echo "INPUT_FILE          ${INPUT_FILE}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY} (${VERBOSITY})"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy save ${INPUT_TAG} --file=${INPUT_FILE} --verbosity=${VERBOSITY}"

# execute command
eval "$CMD" || e_code=1
printf "\n"

exit $e_code
