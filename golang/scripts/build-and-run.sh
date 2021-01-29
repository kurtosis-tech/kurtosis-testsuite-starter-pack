set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"
repo_root_dirpath="$(dirname "${lang_root_dirpath}")"

# Arg-parsing
if [ "${#}" -eq 0 ]; then
    echo "Error: Must specify an action (help, build, run, all)" >&2
    exit 1
fi
action="${1:-}"
shift 1

# Main code
KURTOSIS_DOCKERHUB_ORG="kurtosistech"
api_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_api"
datastore_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_datastore"
custom_params_json='{
    "apiServiceImage" :"'${api_service_image}'",
    "datastoreServiceImage": "'${datastore_service_image}'",
    "isKurtosisCoreDevMode": true
}'

bash "${repo_root_dirpath}/scripts/build-and-run-core.sh" \
    "${action}" \
    "${KURTOSIS_DOCKERHUB_ORG}/kurtosis-go-example" \
    "${lang_root_dirpath}" \
    "${lang_root_dirpath}/testsuite/Dockerfile" \
    "${repo_root_dirpath}/scripts/kurtosis.sh" \
    --custom-params "${custom_params_json}" \
    ${1+"${@}"}
