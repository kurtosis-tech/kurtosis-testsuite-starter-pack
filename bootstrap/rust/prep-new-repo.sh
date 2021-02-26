set -euo pipefail

# =============================================================================
#                                    Constants
# =============================================================================
CARGO_TOML_VERSION_PATTERN='^version = ".*'
CARGO_TOML_FILENAME="Cargo.toml"
CARGO_LOCK_FILENAME="Cargo.lock"
LIB_DIRNAME="lib"
LIB_CRATE_NAME="kurtosis-rust-lib"
TESTSUITE_DIRNAME="testsuite"
DOCKERFILE_FILENAME="Dockerfile"
TESTSUITE_EXAMPLE_PACKAGE_NAME="kurtosis-rust-example"


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
cp "${input_dirpath}/.dockerignore" "${output_dirpath}/"
cp "${input_dirpath}/.gitignore" "${output_dirpath}/"
cp "${input_dirpath}/${CARGO_TOML_FILENAME}" "${output_dirpath}/"
cp "${input_dirpath}/${CARGO_LOCK_FILENAME}" "${output_dirpath}/"
cp -r "${input_dirpath}/${TESTSUITE_DIRNAME}" "${output_dirpath}/"



# =============================================================================
#                         Post-Copy Modifications
# =============================================================================
new_package_name=""
while [ -z "${new_package_name}" ]; do
    read -p "Name for the Rust package that will contain your testsuite project (e.g. my-new-package): " new_package_name
done

# Delete the "lib" entry from the root Cargo.toml file
root_cargo_toml_filepath="${output_dirpath}/${CARGO_TOML_FILENAME}"
lib_line_pattern="\"${LIB_DIRNAME}\","
num_lib_lines="$(grep -c "${lib_line_pattern}" "${root_cargo_toml_filepath}" || true)"
if [ "${num_lib_lines}" -ne 1 ]; then
    echo "Error: Expected exactly one line in '${root_cargo_toml_filepath}' matching pattern '${lib_line_pattern}', but got ${num_lib_lines}" >&2
    exit 1
fi
if ! sed -i '' "/${lib_line_pattern}/d" "${root_cargo_toml_filepath}"; then
    echo "Error: Could not delete line matching pattern '${lib_line_pattern}' from file '${root_cargo_toml_filepath}'" >&2
    exit 1
fi

# Grab the current version number of the lib, so that the bootstrapped repo can depend on it (since it won't have the "lib" directory inside it anymore)
lib_cargo_toml_filepath="${input_dirpath}/${LIB_DIRNAME}/${CARGO_TOML_FILENAME}"
num_lib_version_lines="$(grep -c "${CARGO_TOML_VERSION_PATTERN}" "${lib_cargo_toml_filepath}" || true)"
if [ "${num_lib_version_lines}" -ne 1 ]; then
    echo "Error: Expected exactly one line in '${lib_cargo_toml_filepath}' matching pattern '${CARGO_TOML_VERSION_PATTERN}', but got ${num_lib_version_lines}" >&2
    exit 1
fi
lib_version_line="$(grep "${CARGO_TOML_VERSION_PATTERN}" "${lib_cargo_toml_filepath}" || true)"
lib_version_string="$(echo "${lib_version_line}" | awk '{print $3}')"
if [ -z "${lib_version_string}" ]; then
    echo "Error: Could not extract lib version string from '${lib_cargo_toml_filepath}' by looking for pattern '${CARGO_TOML_VERSION_PATTERN}'" >&2
    exit 1
fi

# Substitute the lib version in for the relative-path dependency, so that the bootstrapped repo uses the version from crates.io
testsuite_cargo_toml_filepath="${output_dirpath}/${TESTSUITE_DIRNAME}/${CARGO_TOML_FILENAME}"
crate_line_pattern="^${LIB_CRATE_NAME} = .*"
num_crate_lines="$(grep -c "${crate_line_pattern}" "${testsuite_cargo_toml_filepath}" || true)"
if [ "${num_crate_lines}" -ne 1 ]; then
    echo "Error: Expected exactly one line in '${testsuite_cargo_toml_filepath}' matching pattern '${lib_line_pattern}', but got ${num_crate_lines}" >&2
    exit 1
fi
new_crate_line="${LIB_CRATE_NAME} = ${lib_version_string}"
if ! sed -i '' "s/${crate_line_pattern}/${new_crate_line}/" "${testsuite_cargo_toml_filepath}"; then
    echo "Error: Could not substitute lib line '${crate_line_pattern}' -> '${new_crate_line}'" >&2
    exit 1
fi

# Substitute the package name for the user's package name
num_package_name_lines="$(grep -c "${TESTSUITE_EXAMPLE_PACKAGE_NAME}" "${testsuite_cargo_toml_filepath}" || true)"
if [ "${num_package_name_lines}" -ne 1 ]; then
    echo "Error: Expected exactly one line in '${testsuite_cargo_toml_filepath}' containing example package name '${TESTSUITE_EXAMPLE_PACKAGE_NAME}, but got ${num_package_name_lines}" >&2
    exit 1
fi
if ! sed -i '' "s/${TESTSUITE_EXAMPLE_PACKAGE_NAME}/${new_package_name}/" "${testsuite_cargo_toml_filepath}"; then
    echo "Error: Could not substitute example package name '${TESTSUITE_EXAMPLE_PACKAGE_NAME}' -> '${new_package_name}'" >&2
    exit 1
fi

# Due to Rust's refusal to allow a "download dependencies only" command, we have special lines in the Dockerfile that deal with caching
# However, when we bootstrap the testsuite, the 'lib' directory will go away so we need to remove all references to it from the Dockerfile
# NOTE: This is a fairly aggressive removal, and one that could be prone to error!
dockerfile_filepath="${output_dirpath}/${TESTSUITE_DIRNAME}/${DOCKERFILE_FILENAME}"
pattern_to_remove="lib"
if ! sed -i '' "/${pattern_to_remove}/d" "${dockerfile_filepath}"; then
    echo "Error: Failed to remove all lines in '${dockerfile_filepath}' containing pattern '${pattern_to_remove}'" >&2
    exit 1
fi
