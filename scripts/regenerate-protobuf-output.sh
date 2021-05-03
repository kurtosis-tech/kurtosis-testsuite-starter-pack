# This script regenerates Go bindings corresponding to the .proto files that define the API container's API
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                           Constants
# ==================================================================================================
# Relative to THE ROOT OF THE ENTIRE REPO
INPUT_RELATIVE_DIRPATH="core-api"
PROTOC_CMD="protoc"

PROTOBUF_FILE_EXT=".proto"

# ------------------------------------------- Golang -----------------------------------------------
GOLANG_DIRNAME="golang"
GO_MOD_FILENAME="go.mod"
GO_MOD_FILE_MODULE_KEYWORD="module"

# -------------------------------------------- Rust -----------------------------------------------
RUST_DIRNAME="rust"
RUST_BINDING_GENERATOR_CMD="rust-protobuf-binding-generator"
ERRONEOUS_GENERATED_FILENAME="google.protobuf.rs"
RUST_MOD_FILENAME="mod.rs"
RUST_FILE_EXT=".rs"

# -------------------------------------------- Shared -----------------------------------------------
# Each language's file extension so we know what files need deleting before we regenerate bindings
declare -A FILE_EXTENSIONS
FILE_EXTENSIONS["${GOLANG_DIRNAME}"]=".go"
FILE_EXTENSIONS["${RUST_DIRNAME}"]="${RUST_FILE_EXT}"

# Per-lang output locations for core API bindings, RELATIVE TO THE LANG ROOT!
declare -A CORE_API_OUTPUT_REL_DIRPATHS
CORE_API_OUTPUT_REL_DIRPATHS["${GOLANG_DIRNAME}"]="lib/core_api_bindings"
CORE_API_OUTPUT_REL_DIRPATHS["${RUST_DIRNAME}"]="lib/src/core_api_bindings"

# Per-lang output locations for suite API bindings
declare -A SUITE_API_OUTPUT_REL_DIRPATHS
SUITE_API_OUTPUT_REL_DIRPATHS["${GOLANG_DIRNAME}"]="lib/rpc_api/bindings"
SUITE_API_OUTPUT_REL_DIRPATHS["${RUST_DIRNAME}"]="lib/src/rpc_api/bindings"

# Maps path to directories containing Protobuf files RELATIVE TO REPO ROOT -> the name of the map variable containing the per-lang output directories
declare -A INPUT_REL_DIRPATHS
INPUT_REL_DIRPATHS["core-api"]="CORE_API_OUTPUT_REL_DIRPATHS"
INPUT_REL_DIRPATHS["suite-api"]="SUITE_API_OUTPUT_REL_DIRPATHS"

# ==================================================================================================
#                                           Main Logic
# ==================================================================================================
# ------------------------------------------- Golang -----------------------------------------------
go_mod_filepath="${root_dirpath}/${GOLANG_DIRNAME}/${GO_MOD_FILENAME}"
if ! [ -f "${go_mod_filepath}" ]; then
    echo "Error: Could not get Go module name; file '${go_mod_filepath}' doesn't exist" >&2
    exit 1
fi
go_module="$(grep "^${GO_MOD_FILE_MODULE_KEYWORD}" "${go_mod_filepath}" | awk '{print $2}')"
if [ "${go_module}" == "" ]; then
    echo "Error: Could not extract Go module from file '${go_mod_filepath}'" >&2
    exit 1
fi

generate_golang_bindings() {
    input_abs_dirpath="${1}"
    output_rel_dirpath="${2}"

    if ! command -v "${PROTOC_CMD}" > /dev/null; then
        echo "Error: No '${PROTOC_CMD}' command found; you'll need to install it via 'brew install protobuf'" >&2
        return 1
    fi

    output_abs_dirpath="${root_dirpath}/${GOLANG_DIRNAME}/${output_rel_dirpath}"

    grpc_flag="--go_out=plugins=grpc:${output_abs_dirpath}"

    fully_qualified_go_pkg="${go_module}/${output_rel_dirpath}"
    for input_filepath in $(find "${input_abs_dirpath}" -type f -name "*${PROTOBUF_FILE_EXT}"); do
        # Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),
        #  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated
        # See also: https://github.com/golang/protobuf/issues/1272
        protobuf_filename="$(basename "${input_filepath}")"
        go_module_flag="--go_opt=M${protobuf_filename}=${fully_qualified_go_pkg};$(basename "${fully_qualified_go_pkg}")"

        if ! "${PROTOC_CMD}" \
                -I="${input_abs_dirpath}" \
                "${grpc_flag}" \
                "${go_module_flag}" \
                "${input_filepath}"; then
            echo "Error: An error occurred generating Golang bindings for file '${input_filepath}'" >&2
            return 1
        fi
    done
}

# -------------------------------------------- Rust -----------------------------------------------
generate_rust_bindings() {
    input_abs_dirpath="${1}"
    output_rel_dirpath="${2}"

    if ! command -v "${RUST_BINDING_GENERATOR_CMD}" > /dev/null; then
        echo "Error: No '${RUST_BINDING_GENERATOR_CMD}' command found; you'll need to install it from https://github.com/kurtosis-tech/rust-protobuf-binding-generator" >&2
        return 1
    fi

    output_abs_dirpath="${root_dirpath}/${RUST_DIRNAME}/${output_rel_dirpath}"
    if ! "${RUST_BINDING_GENERATOR_CMD}" "${input_abs_dirpath}" "${output_abs_dirpath}" $(find "${input_abs_dirpath}" -type f -name "*${PROTOBUF_FILE_EXT}"); then
        echo "Error: An error occurred generating Rust bindings from Protobufs in '${input_abs_dirpath}' to '${output_abs_dirpath}'" >&2
        exit 1
    fi

    # Due to https://github.com/danburkert/prost/issues/228#event-4306587537 , an erroneous empty file gets generated
    # We remove it so as not to confuse the user
    erroneous_filepath="${output_abs_dirpath}/${ERRONEOUS_GENERATED_FILENAME}"
    if [ -f "${erroneous_filepath}" ]; then
        if ! rm "${erroneous_filepath}"; then
            echo "Warning: Could not remove erroneous file '${erroneous_filepath}'"
        fi
    fi

    # Regenerate mod.rs file
    mod_filepath="${output_abs_dirpath}/${RUST_MOD_FILENAME}"
    if [ -f "${mod_filepath}" ]; then
        if ! rm "${mod_filepath}"; then
            echo "Warning: Could not remove ${RUST_MOD_FILENAME} file for output directory '${output_abs_dirpath}'"
        fi
    fi
    for generated_filepath in $(find "${output_abs_dirpath}" -name "*${RUST_FILE_EXT}" -maxdepth 1); do
        generated_filename="$(basename "${generated_filepath}")"
        generated_module="${generated_filename%%${RUST_FILE_EXT}}"
        echo "pub mod ${generated_module};" >> "${mod_filepath}"
    done
}


# ------------------------------------------ Shared Code-----------------------------------------------
# "Schema" of the function provided as a value of this map:
# generate_XXX_bindings(input_abs_dirpath, output_rel_dirpath) where:
#  1. The input_abs_dirpath is an ABSOLUTE path to a directory containing .proto files
#  2. The output_rel_dirpath is a path (relative to the LANG root!!!) where the bindings should be generated
declare -A generators
generators["${GOLANG_DIRNAME}"]="generate_golang_bindings"
generators["${RUST_DIRNAME}"]="generate_rust_bindings"

for input_rel_dirpath in "${!INPUT_REL_DIRPATHS[@]}"; do
    # Name of array containing per-lang output directories where the bindings should be generated
    output_rel_dirpaths_var_name="${INPUT_REL_DIRPATHS["${input_rel_dirpath}"]}"

    input_abs_dirpath="${root_dirpath}/${input_rel_dirpath}"

    for lang in "${!FILE_EXTENSIONS[@]}"; do
        file_ext="${FILE_EXTENSIONS["${lang}"]}"

        eval 'output_rel_dirpath="${'${output_rel_dirpaths_var_name}'["'${lang}'"]}"'
        output_abs_dirpath="${root_dirpath}/${lang}/${output_rel_dirpath}"

        if ! mkdir -p "${output_abs_dirpath}"; then
            echo "Error: Couldn't create ${lang} bindings output directory '${output_abs_dirpath}'" >&2
            exit 1
        fi

        if [ "${output_abs_dirpath}/" == "/" ]; then
            echo "Error: ${lang} output dirpath for input '${input_abs_dirpath}' must not be empty!" >&2
            exit 1
        fi

        if ! find "${output_abs_dirpath}" -name "*${file_ext}" -delete; then
            echo "Error: An error occurred removing the existing ${lang} bindings at '${output_abs_dirpath}'" >&2
            exit 1
        fi

        generator_func="${generators["${lang}"]}"

        # NOTE: When multiple people start developing on this, we won't be able to rely on using the user's local environment for generating bindings because the environments
        # might differ across users
        # We'll need to standardize by:
        #  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
        #  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
        #  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
        #  4) Adding the locally-generated Go output files to .gitignore
        #  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
        if ! "${generator_func}" "${input_abs_dirpath}" "${output_rel_dirpath}"; then
            echo "Error: An error occurred generating ${lang} bindings from input directory '${input_abs_dirpath}' to output directory '${output_abs_dirpath}'" >&2           
            exit 1
        fi

        echo "Successfully generated ${lang} bindings for ${PROTOBUF_FILE_EXT} files in '${input_abs_dirpath}' to '${output_abs_dirpath}'"
    done
done
