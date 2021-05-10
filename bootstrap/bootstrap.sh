set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
repo_root_dirpath="$(dirname "${script_dirpath}")"

# =============================================================================
#                                 Constants
# =============================================================================
# A sed regex that will be used to determine if the user-supplied image name matches the regex
ALLOWED_IMAGE_NAME_CHARS='a-z0-9._/-'

SUPPORTED_LANGS_FILENAME="supported-languages.txt"
INPUT_KURTOSIS_CORE_DIRNAME=".kurtosis"
WRAPPER_SCRIPT_FILENAME="kurtosis.sh"
BUILD_AND_RUN_CORE_FILENAME="build-and-run-core.sh"

# Script for prepping a new testsuite repo
PREP_NEW_REPO_FILENAME="prep-new-repo.sh"
BOOTSTRAP_PARAMS_JSON_FILENAME="bootstrap-suite-params.json"

# Output repo constants
OUTPUT_README_FILENAME="README.md"
OUTPUT_KURTOSIS_CORE_DIRNAME=".kurtosis"
OUTPUT_SCRIPTS_DIRNAME="scripts"
BUILD_AND_RUN_FILENAME="build-and-run.sh"

# =============================================================================
#                             Pre-Arg Parsing
# =============================================================================
supported_langs_filepath="${repo_root_dirpath}/${SUPPORTED_LANGS_FILENAME}"
if ! [ -f "${supported_langs_filepath}" ]; then
    echo "Error: Couldn't find supported languages file '${supported_langs_filepath}'; this is a bug in this script" >&2
    exit 1
fi

# Validate that the supported langs correspond to directories
while read supported_lang; do
    supported_lang_dirpath="${repo_root_dirpath}/${supported_lang}"
    if ! [ -d "${supported_lang_dirpath}" ]; then
        echo "Error: Supported languages file lists langauge '${supported_lang}', but no lang directory '${supported_lang_dirpath}' found corresponding to it; this is a bug in the supported languages file" >&2
        exit 1
    fi
    supported_lang_bootstrap_dirpath="${script_dirpath}/${supported_lang}"
    if ! [ -d "${supported_lang_bootstrap_dirpath}" ]; then
        echo "Error: Supported languages file lists langauge '${supported_lang}', but no lang bootstrap directory '${supported_lang_bootstrap_dirpath}' found corresponding to it; this is a bug in the supported languages file" >&2
        exit 1
    fi
done < "${supported_langs_filepath}"

show_help_and_exit() {
    echo ""
    echo "Usage: $(basename "${0}") lang new_repo_dirpath testsuite_image_name"
    echo ""
    # NOTE: We *could* extract the arg names to variables since they're repeated, but then we wouldn't be able to visually align the indentation here
    echo "  lang                  Language that the new testsuite repo should be in ($(paste -sd '|' "${supported_langs_filepath}"))"
    echo "  new_repo_dirpath      Path to new dirpath where the testsuite repo should be created"
    echo "  testsuite_image_name  Name of the Docker image that will be built to contain the testsuite (must match regex [${ALLOWED_IMAGE_NAME_CHARS}]+, e.g. 'my-test-image')"
    echo ""
    exit 1  # Exit with an error so CI fails if this was accidentally called
}

# =============================================================================
#                           Arg Parsing & Validation
# =============================================================================
lang="${1:-}"
output_dirpath="${2:-}"
testsuite_image="${3:-}"

if [ -z "${lang}" ]; then
    echo "Error: Lang cannot be empty" >&2
    show_help_and_exit
fi
if ! grep -q "^${lang}$" "${supported_langs_filepath}"; then
    echo "Error: Unrecognized lang '${lang}'" >&2
    show_help_and_exit
fi
if [ -z "${output_dirpath}" ]; then
    echo "Error: Output dirpath must not be empty" >&2
    show_help_and_exit
fi
if [ -d "${output_dirpath}" ] && [ "$(ls -A "${output_dirpath}")" ]; then
    echo "Error: Output directory '${output_dirpath}' exists, but is not empty"
    exit 1
fi
if [ -z "${testsuite_image}" ]; then
    echo "Error: Testsuite image cannot be empty" >&2
    exit 1
fi
sanitized_image="$(echo "${testsuite_image}" | sed "s|[^${ALLOWED_IMAGE_NAME_CHARS}]||g")"
if [ "${sanitized_image}" != "${testsuite_image}" ]; then
    echo "Error: Testsuite image name '${testsuite_image}' doesn't match regex [${ALLOWED_IMAGE_NAME_CHARS}]+" >&2
fi

# =============================================================================
#                                 Main Code
# =============================================================================
# Use language-specific prep script to populate contents of output directory
if ! mkdir -p "${output_dirpath}"; then
    echo "Error: Could not create output directory '${output_dirpath}'" >&2
    exit 1
fi
lang_dirpath="${repo_root_dirpath}/${lang}"
lang_bootstrap_dirpath="${script_dirpath}/${lang}"
prep_new_repo_script_filepath="${lang_bootstrap_dirpath}/${PREP_NEW_REPO_FILENAME}"
if ! bash "${prep_new_repo_script_filepath}" "${lang_dirpath}" "${output_dirpath}"; then
    echo "Error: Failed to prep new repo using script '${prep_new_repo_script_filepath}'" >&2
    exit 1
fi

# Copy over Kurtosis Core scripts
input_kurtosis_core_dirpath="${repo_root_dirpath}/${INPUT_KURTOSIS_CORE_DIRNAME}"
output_kurtosis_core_dirpath="${output_dirpath}/${OUTPUT_KURTOSIS_CORE_DIRNAME}"
if ! mkdir -p "${output_kurtosis_core_dirpath}"; then
    echo "Error: Could not create Kurtosis Core directory '${output_kurtosis_core_dirpath}'" >&2
    exit 1
fi
if ! cp "${input_kurtosis_core_dirpath}/${WRAPPER_SCRIPT_FILENAME}" "${output_kurtosis_core_dirpath}/"; then
    echo "Error: Could not copy ${WRAPPER_SCRIPT_FILENAME} to ${output_kurtosis_core_dirpath}" >&2
    exit 1
fi
if ! cp "${input_kurtosis_core_dirpath}/${BUILD_AND_RUN_CORE_FILENAME}" "${output_kurtosis_core_dirpath}/"; then
    echo "Error: Could not copy ${BUILD_AND_RUN_CORE_FILENAME} to ${output_kurtosis_core_dirpath}" >&2 
    exit 1
fi

# Create build-and-run wrapper over build-and-run-core
output_scripts_dirpath="${output_dirpath}/${OUTPUT_SCRIPTS_DIRNAME}"
if ! mkdir -p "${output_scripts_dirpath}"; then
    echo "Error: Could not create the output scripts directory at '${output_scripts_dirpath}'" >&2
    exit 1
fi
bootstrap_params_json_filepath="${lang_bootstrap_dirpath}/${BOOTSTRAP_PARAMS_JSON_FILENAME}"
if ! [ -f "${bootstrap_params_json_filepath}" ]; then
    echo "Error: Could not find bootstrap testsuite params at '${bootstrap_params_json_filepath}'; this is a bug with the bootstrapping process" >&2
    exit 1
fi
bootstrap_params_json="$(cat "${bootstrap_params_json_filepath}")"
output_build_and_run_wrapper_filepath="${output_scripts_dirpath}/${BUILD_AND_RUN_FILENAME}"
cat << EOF > "${output_build_and_run_wrapper_filepath}"
set -euo pipefail
script_dirpath="\$(cd "\$(dirname "\${0}")" && pwd)"
root_dirpath="\$(dirname "\${script_dirpath}")"
kurtosis_core_dirpath="\${root_dirpath}/${OUTPUT_KURTOSIS_CORE_DIRNAME}"

show_help_and_exit() {
    echo ""
    echo "Usage: \$(basename "\${0}") action [kurtosis.sh_arg1] [kurtosis.sh_arg2]..."
    echo ""
    echo "  action              The action that should be passed to the underlying ${BUILD_AND_RUN_CORE_FILENAME} script to tell it which action should be taken (call"
    echo "                          'bash \${kurtosis_core_dirpath}/${BUILD_AND_RUN_CORE_FILENAME} help' directly for all available actions)"
    echo "  kurtosis.sh_args    Optional, supplemental args that should be passed to the ${WRAPPER_SCRIPT_FILENAME} script to modify testsuite execution behaviour (call"
    echo "                          'bash \${kurtosis_core_dirpath}/${WRAPPER_SCRIPT_FILENAME} --help' directly for all available args)"
    echo ""
    exit 1  # Exit with error so CI will fail if it accidentally calls this
}

if [ "\${#}" -eq 0 ]; then
    show_help_and_exit
fi
action="\${1:-}"
shift 1
if [ "\${action}" == "help" ]; then
    show_help_and_exit
fi

# >>>>>>>> Add custom testsuite parameters here <<<<<<<<<<<<<
custom_params_json='${bootstrap_params_json}'
# >>>>>>>> Add custom testsuite parameters here <<<<<<<<<<<<<

bash "\${kurtosis_core_dirpath}/${BUILD_AND_RUN_CORE_FILENAME}" \\
    "\${action}" \\
    "${testsuite_image}" \\
    "\${root_dirpath}" \\
    "\${root_dirpath}/testsuite/Dockerfile" \\
    "\${kurtosis_core_dirpath}/${WRAPPER_SCRIPT_FILENAME}" \\
    --custom-params "\${custom_params_json}" \\
    \${1+"\${@}"}
EOF
if [ "${?}" -ne 0 ]; then
    echo "Error: Could not write build-and-run wrapper to '${output_build_and_run_wrapper_filepath}'" >&2
    exit 1
fi
if ! chmod u+x "${output_build_and_run_wrapper_filepath}"; then
    echo "Error: Could not make build-and-run wrapper '${output_build_and_run_wrapper_filepath}' executable" >&2
    exit 1
fi

# README file
output_readme_filepath="${output_dirpath}/${OUTPUT_README_FILENAME}"
cat << EOF > "${output_readme_filepath}"
My Kurtosis Testsuite
=====================
Welcome to your new Kurtosis testsuite! To run your testsuite, run 'bash ${OUTPUT_SCRIPTS_DIRNAME}/${BUILD_AND_RUN_FILENAME} all'. To see help information, run 'bash ${OUTPUT_SCRIPTS_DIRNAME}/${BUILD_AND_RUN_FILENAME} help'.
EOF
if [ "${?}" -ne 0 ]; then
    echo "Error: Could not write README file to '${output_readme_filepath}'" >&2
    exit 1
fi

#Initialize the new repo as a Git directory, because running the testsuite depends on it
if ! command -v git &> /dev/null; then
    echo "Error: Git is required to create a new testsuite repo, but it is not installed" >&2
    exit 1
fi
if ! cd "${output_dirpath}"; then
    echo "Error: Could not cd to new testsuite repo '${output_dirpath}', which is necessary for initializing it as a Git repo" >&2
    exit 1
fi
if ! git init; then
    echo "Error: Could not initialize the new repo as a Git repository" >&2
    exit 1
fi
if ! git add .; then
    echo "Error: Could not stage files in new repo for committing" >&2
    exit 1
fi
if ! git commit -m "Initial commit" > /dev/null; then
    echo "Error: Could not create initial commit in new repo" >&2
    exit 1
fi

echo "Bootstrap successful!"
echo " - Your new testsuite can be run with 'bash ${output_scripts_dirpath}/${BUILD_AND_RUN_FILENAME} all'"
echo " - To continue with the quickstart, head back to the quickstart steps: https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart"
