#!/usr/bin/env bash

# set defaults when not set
[ -z "${INPUT_REVISION}" ]    && INPUT_REVISION=${GITHUB_SHA}
[ -z "${INPUT_VERBOSITY}" ]   && INPUT_VERBOSITY="error"
[ -z "${INPUT_SOURCE_URL}" ]  && INPUT_SOURCE_URL="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}"
[ -z "${INPUT_REGO_VERSION}"] && INPUT_REGO_VERSION="rego.v1"
[ -z "${INPUT_OPTIMIZE}"]     && INPUT_OPTIMIZE="disabled"
[ -z "${INPUT_ENTRYPOINT}"]   && INPUT_ENTRYPOINT=""

# validate if values are set
[ -z "${INPUT_SRC}" ]        && echo "INPUT_SRC is not set exiting" && exit 2
[ -z "${INPUT_TAG}" ]        && exit "INPUT_TAG is not set exiting" && exit 2
[ -z "${INPUT_REVISION}" ]   && exit "INPUT_REVISION is not set exiting" && exit 2
[ -z "${INPUT_VERBOSITY}" ]  && exit "INPUT_VERBOSITY is not set exiting" && exit 2
[ -z "${INPUT_SOURCE_URL}" ] && exit "INPUT_SOURCE_URL is not set exiting" && exit 2
[ -z "${INPUT_OPTIMIZE}" ]   && exit "INPUT_OPTIMIZE is not set exiting" && exit 2

if [[ -z "${GITHUB_WORKSPACE}" ]]; then
  SRC_PATH=$PWD/$INPUT_SRC
else
  # calculate paths relative to the workspace (GITHUB_WORKSPACE).
  SRC_PATH=$GITHUB_WORKSPACE/$INPUT_SRC
fi

# validate path variables
if [[ ! -d ${SRC_PATH} ]]; then
    echo "SRC_PATH path does not exist [${SRC_PATH}]"
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

OPTIMIZE=0
case ${INPUT_OPTIMIZE} in
  "disabled")
    OPTIMIZE=0
    ;;
  "recommended")
    OPTIMIZE=1
    ;;
  "aggressive")
    OPTIMIZE=2
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
echo "INPUT_SOURCE_URL    ${INPUT_SOURCE_URL}"
echo "INPUT_REGO_VERSION  ${INPUT_REGO_VERSION}"
echo "INPUT_OPTIMIZE      ${INPUT_OPTIMIZE} (${OPTIMIZE})"
echo "INPUT_ENTRYPOINT    ${INPUT_ENTRYPOINT}"
echo "SRC_PATH            ${SRC_PATH}"
printf "\n"

#
# start execution block
#
e_code=0

# construct commandline arguments 
CMD="/app/policy build ${SRC_PATH} --tag ${INPUT_TAG} --rego-version=${INPUT_REGO_VERSION} --verbosity=${VERBOSITY} --optimize=${OPTIMIZE} --annotations=org.opencontainers.image.source=${INPUT_SOURCE_URL} --entrypoint=${INPUT_ENTRYPOINT}"

# execute command
eval "$CMD" || e_code=1
printf "\n"

if [ "${VERBOSITY}" -ge "1" ]; then 
  /app/policy images
  printf "\n"
fi 

exit $e_code
