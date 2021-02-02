set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
repo_root_dirpath="$(dirname "${script_dirpath}")"

# =============================================================================
#                                 Constants
# =============================================================================
SUPPORTED_LANGS_FILENAME="supported-languages.txt"

ROOT_SCRIPTS_DIRNAME="scripts"
WRAPPER_SCRIPT_FILENAME="kurtosis.sh"
BUILD_AND_RUN_CORE_FILENAME="build-and-run-core.sh"

# Script for prepping a new testsuite repo
PREP_NEW_REPO_FILENAME="prep-new-repo.sh"
BOOTSTRAP_PARAMS_JSON_FILENAME="bootstrap-suite-params.json"

# Output repo constants
OUTPUT_README_FILENAME="README.md"
KURTOSIS_CORE_DIRNAME=".kurtosis"
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
    echo "$(basename "${0}") lang new_repo_dirpath"
    echo ""
    echo "  lang                Language that the new testsuite repo should be in ($(paste -d '|' "${supported_langs_filepath}"))"
    echo "  new_repo_dirpath    Path to the new directory to create to contain the testsuite repo"
    echo ""
    exit 1  # Exit with an error so CI fails if this was accidentally called
}

# =============================================================================
#                           Arg Parsing & Validation
# =============================================================================
lang="${1:-}"
output_dirpath="${2:-}"

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

# =============================================================================
#                                 Main Code
# =============================================================================
testsuite_image=""
while [ -z "${testsuite_image}" ]; do
    echo "Name for the testsuite Docker image that the repo will build, which must conform to the Docker image naming rules:"
    echo "  https://docs.docker.com/engine/reference/commandline/tag/#extended-description"
    read -p "Image name (e.g. your-dockerhub-org/your-image-name): " testsuite_image
done

if ! mkdir -p "${output_dirpath}"; then
    echo "Error: Could not create output directory '${output_dirpath}'" >&2
    exit 1
fi

lang_dirpath="${repo_root_dirpath}/${lang}"
if ! cp -r "${lang_dirpath}/" "${output_dirpath}/"; then
    echo "Error: Could not copy files from ${lang_dirpath} to ${output_dirpath}" >&2
    exit 1
fi

lang_bootstrap_dirpath="${script_dirpath}/${lang}"
prep_new_repo_script_filepath="${lang_bootstrap_dirpath}/${PREP_NEW_REPO_FILENAME}"
if ! bash "${prep_new_repo_script_filepath}" "${lang_dirpath}" "${output_dirpath}"; then
    echo "Error: Failed to prep new repo using script '${prep_new_repo_script_filepath}'" >&2
    exit 1
fi

output_scripts_dirpath="${output_dirpath}/${OUTPUT_SCRIPTS_DIRNAME}"
if ! mkdir -p "${output_scripts_dirpath}"; then
    echo "Error: Could not create the output scripts directory at '${output_scripts_dirpath}'" >&2
    exit 1
fi

# TODO PUT KURTOSIS STUFF IN ITS OWN DIRECTORY
input_scripts_dirpath="${repo_root_dirpath}/${ROOT_SCRIPTS_DIRNAME}"
if ! cp "${input_scripts_dirpath}/${WRAPPER_SCRIPT_FILENAME}" "${output_scripts_dirpath}/"; then
    echo "Error: Could not copy ${WRAPPER_SCRIPT_FILENAME} to ${output_scripts_dirpath}" >&2
    exit 1
fi
if ! cp "${input_scripts_dirpath}/${BUILD_AND_RUN_CORE_FILENAME}" "${output_scripts_dirpath}/"; then
    echo "Error: Could not copy ${BUILD_AND_RUN_CORE_FILENAME} to ${output_scripts_dirpath}" >&2 
    exit 1
fi

bootstrap_params_json_filepath="${lang_bootstrap_dirpath}/${BOOTSTRAP_PARAMS_JSON_FILENAME}"
bootstrap_params_json="$(cat "${bootstrap_params_json_filepath}")"
output_build_and_run_wrapper_filepath="${output_scripts_dirpath}/${BUILD_AND_RUN_FILENAME}"
cat <<- EOF > "${output_build_and_run_wrapper_filepath}"
    set -euo pipefail
    script_dirpath="\$(cd "\$(dirname "\${0}")" && pwd)"
    root_dirpath="\$(dirname "\${script_dirpath}")"
    kurtosis_core_dirpath="\${root_dirpath}/${KURTOSIS_CORE_DIRNAME}"

    # Arg-parsing
    if [ "\${#}" -eq 0 ]; then
        echo "Error: Must specify an action (help, build, run, all)" >&2
        exit 1
    fi
    action="\${1:-}"
    shift 1

    # Main code
    # >>>>>>>> Add custom testsuite parameters here <<<<<<<<<<<<<
    custom_params_json='${bootstrap_params_json}'
    # >>>>>>>> Add custom testsuite parameters here <<<<<<<<<<<<<

    bash "\${kurtosis_core_dirpath}/${BUILD_AND_RUN_CORE_FILENAME}" \
        "\${action}" \
        "${testsuite_image}" \
        "${output_dirpath}" \
        "\${root_dirpath}/testsuite/Dockerfile" \
        "\${kurtosis_core_dirpath}/${WRAPPER_SCRIPT_FILENAME}" \
        --custom-params "\${custom_params_json}" \
        \${1+"\${@}"}
EOF
if [ "${?}" -ne 0 ]; then
    echo "Error: Could not write build-and-run wrapper to '${output_build_and_run_wrapper_filepath}'" >&2
    exit 1
fi


# We have to initialize the new repo as a Git directory because bootstrapping depends on it
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
if ! git add . "${output_dirpath}"; then
    echo "Error: Could not stage files in new repo for committing" >&2
    exit 1
fi
if ! git commit -m "Initial commit"; then
    echo "Error: Could not create initial commit in new repo" >&2
    exit 1
fi

output_readme_filepath="${output_dirpath}/${OUTPUT_README_FILENAME}"
cat <<- EOF "${OUTPUT_README_FILENAME}" > "${output_readme_filepath}"
    My Kurtosis Testsuite
    =====================
    Welcome to your new Kurtosis testsuite! Now that you've bootstrapped, you can continue with the quickstart section from the "Run your testsuite" section.

    To run your testsuite, run 'bash ${OUTPUT_SCRIPTS_DIRNAME}/${BUILD_AND_RUN_FILENAME} all'
EOF

echo "Bootstrap complete! Your new testsuite can be run with 'bash ${output_scripts_dirpath}/${BUILD_AND_RUN_FILENAME} all'"
