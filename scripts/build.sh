#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_REVISION}" ]  && INPUT_REVISION=${GITHUB_SHA}
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="error"

# validate if values are set
[ -z "${INPUT_SRC}" ]       && echo "INPUT_SRC is not set exiting" && exit 2
[ -z "${INPUT_TAG}" ]       && exit "INPUT_TAG is not set exiting" && exit 2
[ -z "${INPUT_REVISION}" ]  && exit "INPUT_REVISION is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set exiting" && exit 2

if [[ -z "${GITHUB_WORKSPACE}" ]]; then
  SRC_PATH=$PWD/$INPUT_SRC
else
  # calculate paths relative to the workspace (GITHUB_WORKSPACE).
  SRC_PATH=$GITHUB_WORKSPACE/$INPUT_SRC
fi

# validate path variables
if [[ ! -d ${SRC_PATH} ]]; then
    echo "INPUT_SRC path does not exist [${INPUT_SRC}]"
    exit 1
fi

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
echo "POLICY-BUILD        $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_SRC           ${INPUT_SRC}"
echo "INPUT_TAG           ${INPUT_TAG}"
echo "INPUT_REVISION      ${INPUT_REVISION}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY} (${VERBOSITY})"
echo "SRC_PATH            ${SRC_PATH}"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy build ${SRC_PATH} --tag ${INPUT_TAG} --verbosity=${VERBOSITY}"

# execute command
eval "$CMD" || e_code=1
printf "\n"

if [ "${VERBOSITY}" -ge "1" ]; then 
  /app/policy images
  printf "\n"
fi 

exit $e_code
