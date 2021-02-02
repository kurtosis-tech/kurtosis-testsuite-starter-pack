set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"
repo_root_dirpath="$(dirname "${lang_root_dirpath}")"

action="${1:-}"
shift 1

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
    "${lang_root_dirpath}/Dockerfile" \
    "${repo_root_dirpath}/.kurtosis/kurtosis.sh" \
    --custom-params "${custom_params_json}" \
    ${1+"${@}"}
