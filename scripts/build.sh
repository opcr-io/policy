#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_REVISION}" ]  && INPUT_REVISION=${GITHUB_SHA}
[ -z "${INPUT_SERVER}" ]    && INPUT_SERVER="opcr.io"
[ -z "${INPUT_VERBOSITY}" ] && INPUT_VERBOSITY="0"

# validate if values are set
[ -z "${INPUT_SRC}" ]       && echo "INPUT_SRC is not set exiting" && exit 2
[ -z "${INPUT_TAG}" ]       && exit "INPUT_TAG is not set exiting" && exit 2
[ -z "${INPUT_REVISION}" ]  && exit "INPUT_REVISION is not set exiting" && exit 2
[ -z "${INPUT_SERVER}" ]    && exit "INPUT_SERVER is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ] && exit "INPUT_VERBOSITY is not set exiting" && exit 2

# calculate paths relative to the workspace (GITHUB_WORKSPACE).
SRC_PATH=$GITHUB_WORKSPACE/$INPUT_SRC

# validate path variables
if [[ ! -d ${SRC_PATH} ]]; then
    echo "INPUT_SRC path does not exist [${INPUT_SRC}]"
    exit 1
fi

# output all inputs env variables
echo "POLICY-BUILD        $(/app/policy version | sed 's/Policy CLI.//g')"
printf "\n"
printf "\n"
echo "INPUT_SRC           ${INPUT_SRC}"
echo "INPUT_TAG           ${INPUT_TAG}"
echo "INPUT_REVISION      ${INPUT_REVISION}"
echo "INPUT_SERVER        ${INPUT_SERVER}"
echo "INPUT_VERBOSITY     ${INPUT_VERBOSITY}"
echo "GITHUB_WORKSPACE    ${GITHUB_WORKSPACE}"
printf "\n"
echo ">>> CALCULATED VALUES"
echo "SRC_PATH            ${SRC_PATH}"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy build ${SRC_PATH} -t ${INPUT_TAG}"

# execute command
eval "$CMD" || e_code=1
printf "\n"

/app/policy images
printf "\n"

exit $e_code
