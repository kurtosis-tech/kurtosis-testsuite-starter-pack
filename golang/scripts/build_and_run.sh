set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ====================== CONSTANTS =======================================================
# TODO Extract these constants out into their own file so the bootstrap can easily replace it
# The name of to give the Docker image containing the testsuite
KURTOSIS_DOCKERHUB_ORG="kurtosistech"
SUITE_IMAGE="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-go-example"

# When enabled, additional tests are run that are used to verify Kurtosis Core functionality
# If using the free trial, this should be false else the test number limit will be hit
IS_KURTOSIS_CORE_DEV_MODE="true"

BUILD_ACTION="build"
RUN_ACTION="run"
BOTH_ACTION="all"
HELP_ACTION="help"

# ====================== ARG PARSING =======================================================
show_help() {
    echo "${0} <action> [<extra kurtosis.sh script args...>]"
    echo ""
    echo "  This script will optionally a) build your Kurtosis testsuite into a Docker image and/or b) run it via a call to the kurtosis.sh script"
    echo ""
    echo "  To select behaviour, choose from the following actions:"
    echo ""
    echo "    help    Displays this messages"
    echo "    build   Executes only the build step, skipping the run step"
    echo "    run     Executes only the run step, skipping the build step"
    echo "    all     Executes both build and run steps"
    echo ""
    echo "  To see the args the kurtosis.sh script accepts for the 'run' phase, call '$(basename ${0}) all --help'"
    echo ""
}

if [ "${#}" -eq 0 ]; then
    show_help
    exit 1     # Exit with error code so we dont't get spurious CI passes
fi

action="${1:-}"
shift 1

do_build=true
do_run=true
case "${action}" in
    ${HELP_ACTION})
        show_help
        exit 0
        ;;
    ${BUILD_ACTION})
        do_build=true
        do_run=false
        ;;
    ${RUN_ACTION})
        do_build=false
        do_run=true
        ;;
    ${BOTH_ACTION})
        do_build=true
        do_run=true
        ;;
    *)
        echo "Error: First argument must be one of '${HELP_ACTION}', '${BUILD_ACTION}', '${RUN_ACTION}', or '${BOTH_ACTION}'" >&2
        exit 1
        ;;
esac

# ====================== MAIN LOGIC =======================================================
# Captures the first of tag > branch > commit
git_ref="$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)"
docker_tag="$(echo "${git_ref}" | sed 's,[/:],_,g')"

lang_root_dirpath="$(dirname "${script_dirpath}")"
if "${do_build}"; then
    echo "Running unit tests..."
    # TODO Extract this go-specific logic out into a separate script so we can copy/paste the build_and_run.sh between various languages
    if ! go test "${lang_root_dirpath}/..."; then
        echo "Tests failed!"
        exit 1
    fi
    echo "Tests succeeded"

    if ! [ -f "${lang_root_dirpath}"/.dockerignore ]; then
        echo "Error: No .dockerignore file found in root; this is required so Docker caching works properly" >&2
        exit 1
    fi

    echo "Building ${SUITE_IMAGE} Docker image..."
    docker build -t "${SUITE_IMAGE}:${docker_tag}" -f "${lang_root_dirpath}/testsuite/Dockerfile" "${lang_root_dirpath}"
fi

if "${do_run}"; then
    # ======================================= Custom Docker environment variables ========================================================
    # NOTE: Replace these with whatever custom properties your service needs
    api_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_api"
    datastore_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_datastore"
    custom_params_json='{
        "apiServiceImage" :"'${api_service_image}'",
        "datastoreServiceImage": "'${datastore_service_image}'",
        "isKurtosisCoreDevMode": '${IS_KURTOSIS_CORE_DEV_MODE}'
    }'
    # ====================================== End custom Docker environment variables =====================================================
    # The funky ${1+"${@}"} incantation is how you you feed arguments exactly as-is to a child script in Bash
    # ${*} loses quoting and ${@} trips set -e if no arguments are passed, so this incantation says, "if and only if 
    #  ${1} exists, evaluate ${@}"
    bash "${script_dirpath}/kurtosis.sh" --custom-params "${custom_params_json}" ${1+"${@}"} "${SUITE_IMAGE}:${docker_tag}"
fi
