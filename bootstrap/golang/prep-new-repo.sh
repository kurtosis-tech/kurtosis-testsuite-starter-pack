set -euo pipefail

# =============================================================================
#                                    Constants
# =============================================================================
TESTSUITE_IMPL_DIRNAME="testsuite"

# Constants 
GO_MOD_FILENAME="go.mod"
GO_MOD_MODULE_KEYWORD="module "  # The key we'll look for when replacing the module name in go.mod

# Frustratingly, there's no way to say "do in-place replacement" in sed that's compatible on both Mac and Linux
# Instead, we add this suffix and delete the backup files after
SED_INPLACE_FILE_SUFFIX=".sedreplace"

# =============================================================================
#                             Arg-Parsing & Validation
# =============================================================================
input_dirpath="${1:-}"
if [ -z "${input_dirpath}" ]; then
    echo "Error: Empty source directory to copy from" >&2
    exit 1
fi
if ! [ -d "${input_dirpath}" ]; then
    echo "Error: Dirpath to copy source from '${input_dirpath}' is nonexistent" >&2
    exit 1
fi

output_dirpath="${2:-}"
if [ -z "${output_dirpath}" ]; then
    echo "Error: Empty output directory to copy to" >&2
    exit 1
fi
if ! [ -d "${output_dirpath}" ]; then
    echo "Error: Output dirpath to copy to '${output_dirpath}' is nonexistent" >&2
    exit 1
fi


# =============================================================================
#                               Copying Files
# =============================================================================
cp "${input_dirpath}/${GO_MOD_FILENAME}" "${output_dirpath}/"
cp "${input_dirpath}/go.sum" "${output_dirpath}/"
cp -r "${input_dirpath}/${TESTSUITE_IMPL_DIRNAME}" "${output_dirpath}/"


# =============================================================================
#                         Post-Copy Modifications
# =============================================================================
# Allow setting the module name programatically, for testing in CI
new_module_name="${GO_NEW_MODULE_NAME:-}"
echo "Go uses Github as a its dependency management store. You'll now need to:"
echo "  1) Create a new Github repo to contain your testsuite, if you don't have one already"
echo "  2) Enter the URL of the repo on Github WITHOUT the leading 'https://' below (e.g. 'github.com/my-org/my-repo')"
echo "NOTE: This value is technically the Go module name. If you're unfamiliar with what this is, you can read more here: https://golang.org/ref/mod"
echo ""
while [ -z "${new_module_name}" ]; do
    read -p "Module name: " new_module_name
done

# Validation, to save us in case someone changes stuff in the future
go_mod_filepath="${output_dirpath}/${GO_MOD_FILENAME}"
if [ "$(grep "^${GO_MOD_MODULE_KEYWORD}" "${go_mod_filepath}" | wc -l)" -ne 1 ]; then
    echo "Validation failed: Could not find exactly one line in ${GO_MOD_FILENAME} with keyword '${GO_MOD_MODULE_KEYWORD}' for use when replacing with the user's module name" >&2
    exit 1
fi

# Replace module names in code (we need the "-i '' " argument because Mac sed requires it)
existing_module_name="$(grep "module" "${go_mod_filepath}" | awk '{print $2}')"
if ! sed -i"${SED_INPLACE_FILE_SUFFIX}" "s,${existing_module_name},${new_module_name},g" ${go_mod_filepath}; then
    echo "Error: Could not replace Go module name in mod file '${go_mod_filepath}'" >&2
    exit 1
fi
# We search for old_module_name/testsuite because we don't want the old_module_name/lib entries to get renamed
if ! sed -i"${SED_INPLACE_FILE_SUFFIX}" "s,${existing_module_name}/${TESTSUITE_IMPL_DIRNAME},${new_module_name}/${TESTSUITE_IMPL_DIRNAME},g" $(find "${output_dirpath}" -type f); then
    echo "Error: Could not replace Go module name in code files" >&2
    exit 1
fi

# NOTE: Leave this as the last command in the file!! It removes all the backup files created by our in-place sed (see above for why this is necessary)
if ! find "${output_dirpath}" -name "*${SED_INPLACE_FILE_SUFFIX}" -delete; then
    echo "Error: Failed to remove the backup files suffixed with '${SED_INPLACE_FILE_SUFFIX}' that we created doing in-place string replacement with sed" >&2
    exit 1
fi
