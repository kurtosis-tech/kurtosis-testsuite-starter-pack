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

OUTPUT_SCRIPTS_DIRNAME="scripts"

# =============================================================================
#                             Pre-Arg Parsing
# =============================================================================
supported_langs_filepath="${repo_root_dirpath}/${SUPPORTED_LANGS_FILENAME}"
if ! [ -f "${supported_langs_filepath}" ]; then
    echo "Error: Couldn't find supported languages file '${supported_langs_filepath}'; this is a bug in this script" >&2
    exit 1
fi

# Validate that the supported langs correspond to directories
while read lang; do
    if ! [ -d "${repo_root_dirpath}/${lang}" ]; then
        echo "Error: Supported languages file lists langauge '${lang}', but no directory was found corresponding to it; this is a bug in the supported languages file" >&2
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
#                               Arg Validation
# =============================================================================
lang="${1:-}"
output_dirpath="${2:-}"

if [ -z "${lang}" ]; then
    echo "Error: Lang cannot be empty" >&2
    show_help_and_exit
fi
if ! [ -d "${repo_root_dirpath}/${lang}" ]; then
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
if ! mkdir -p "${output_dirpath}"; then
    echo "Error: Could not create output directory '${output_dirpath}'" >&2
    exit 1
fi

lang_dirpath="${repo_root_dirpath}/${lang}"
if ! cp -r "${lang_dirpath}/" "${output_dirpath}/"; then
    echo "Error: Could not copy files from ${lang_dirpath} to ${output_dirpath}" >&2
    exit 1
fi

# TODO call the bootstrap.sh inside the new repo to remove unnecessary files and sed the appropriate files
# TODO Or maybe whitelist the files that get copied over??

output_scripts_dirpath="${output_dirpath}/${OUTPUT_SCRIPTS_DIRNAME}"
if ! mkdir -p "${output_scripts_dirpath}"; then
    echo "Error: Could not create the output scripts directory at '${output_scripts_dirpath}'" >&2
    exit 1
fi

input_scripts_dirpath="${repo_root_dirpath}/${ROOT_SCRIPTS_DIRNAME}"
if ! cp "${input_scripts_dirpath}/${WRAPPER_SCRIPT_FILENAME}" "${output_scripts_dirpath}/"; then
    echo "Error: Could not copy ${WRAPPER_SCRIPT_FILENAME} to ${output_scripts_dirpath}" >&2
    exit 1
fi
if ! cp "${input_scripts_dirpath}/${BUILD_AND_RUN_CORE_FILENAME}" "${output_scripts_dirpath}/"; then
    echo "Error: Could not copy ${BUILD_AND_RUN_CORE_FILENAME} to ${output_scripts_dirpath}" >&2 
    exit 1
fi


# TODO DEBUGGING
exit 99

