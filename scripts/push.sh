#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="error"

# validate if values are set
[ -z "${INPUT_TAGS}" ]       && echo "INPUT_TAGS is not set exiting" && exit 2
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
echo "POLICY-PUSH         $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_TAGS          ${INPUT_TAGS}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY} (${VERBOSITY})"
printf "\n"

#
# start execution block
#
e_code=0

for TAG in ${INPUT_TAGS}; do
  # construct commandline arguments
  CMD="/app/policy push ${TAG} --verbosity=${VERBOSITY}"

  # execute command
  eval "$CMD" || e_code=1
  printf "\n"
done

exit $e_code
