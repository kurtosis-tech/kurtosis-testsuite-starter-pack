set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==========================================================================================
#                                         Constants
# ==========================================================================================
BOOTSTRAP_SCRIPTS_DIRNAME="bootstrap"
BOOTSTRAP_SCRIPT_FILENAME="bootstrap.sh"
SUPPORTED_LANGS_FILENAME="supported-languages.txt"
SCRIPTS_DIRNAME_INSIDE_TESTSUITE="scripts"
BUILD_AND_RUN_FILENAME="build-and-run.sh"
BUILD_AND_RUN_ALL_CMD="all"

# Bootstrapping normally requires input from STDIN, but we can set
#  certain variables so this isn't required for CI
# NOTE: This won't handle flag values that contain spaces, though it can handle multiple flags separated by a space
declare -A CUSTOM_LANG_BOOTSTRAP_FLAGS 
CUSTOM_LANG_BOOTSTRAP_FLAGS[golang]="GO_NEW_MODULE_NAME=github.com/test/test-module"
CUSTOM_LANG_BOOTSTRAP_FLAGS[rust]="RUST_NEW_PACKAGE_NAME=test-package"


# ==========================================================================================
#                                           Main code
# ==========================================================================================
bootstrap_script_filepath="${root_dirpath}/${BOOTSTRAP_SCRIPTS_DIRNAME}/${BOOTSTRAP_SCRIPT_FILENAME}"
echo "Bootstrapping and running new testsuites for all languages..."
for lang in $(cat "${root_dirpath}/${SUPPORTED_LANGS_FILENAME}"); do
    echo "Bootstrapping and running ${lang} testsuite..."
    output_dirpath="$(mktemp -d)"
    testsuite_image="bootstrap-test-${lang}-image"
    lang_specific_vars_to_set="${CUSTOM_LANG_BOOTSTRAP_FLAGS[${lang}]}"
    command="${lang_specific_vars_to_set} ${bootstrap_script_filepath} ${lang} ${output_dirpath} ${testsuite_image}"
    if ! eval "${command}"; then
        echo "Error: Bootstrapping ${lang} testsuite failed" >&2
        exit 1
    fi

    build_and_run_filepath="${output_dirpath}/${SCRIPTS_DIRNAME_INSIDE_TESTSUITE}/${BUILD_AND_RUN_FILENAME}"
    if ! "${build_and_run_filepath}" "${BUILD_AND_RUN_ALL_CMD}"; then
        echo "Error: The bootstrapped ${lang} testsuite failed" >&2
        exit 1
    fi
    echo "Successfully bootstrapped and ran new ${lang} testsuite"
done
echo "Successfully bootstrapped and ran new testsuites for all languages!"
