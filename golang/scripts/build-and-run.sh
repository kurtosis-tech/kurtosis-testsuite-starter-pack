set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"
repo_root_dirpath="$(dirname "${lang_root_dirpath}")"

KURTOSIS_DOCKERHUB_ORG="kurtosistech"

if [ "${#}" -eq 0 ]; then
    echo "Error: Must provide at least one argument (pass 'help' to see options)" >&2
    exit 1
fi
action="${1:-}"
shift 1

custom_params_json='{
    "apiServiceImage" :"'${KURTOSIS_DOCKERHUB_ORG}'/example-api-server",
    "datastoreServiceImage": "'${KURTOSIS_DOCKERHUB_ORG}'/example-datastore-server"
}'

bash "${repo_root_dirpath}/.kurtosis/build-and-run-core.sh" \
    "${action}" \
    "${KURTOSIS_DOCKERHUB_ORG}/kurtosis-golang-example" \
    "${lang_root_dirpath}" \
    "${lang_root_dirpath}/testsuite/Dockerfile" \
    --custom-params "${custom_params_json}" \
    ${1+"${@}"}
