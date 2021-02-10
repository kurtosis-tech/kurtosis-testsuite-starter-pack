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

# ------------------------------------------- Golang -----------------------------------------------
GOLANG_DIRNAME="golang"
GO_MOD_FILENAME="go.mod"
GO_MOD_FILE_MODULE_KEYWORD="module"
GO_RELATIVE_OUTPUT_DIRPATH="lib/core_api_bindings"   # Relative to the root of the lang dir!

# -------------------------------------------- Rust -----------------------------------------------
RUST_DIRNAME="rust"
RUST_BINDING_GENERATOR_CMD="rust-protobuf-binding-generator"
RUST_RELATIVE_OUTPUT_DIRPATH="lib/src/core_api_bindings"



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
go_bindings_pkg="${go_module}/${GO_RELATIVE_OUTPUT_DIRPATH}"

generate_golang_bindings() {
    input_dirpath="${1}"
    shift 1

    output_dirpath="${1}"
    shift 1

    if ! command -v "${PROTOC_CMD}"; then
        echo "Error: No '${PROTOC_CMD}' command found; you'll need to install it via 'brew install protobuf'" >&2
        return 1
    fi

    grpc_flag="--go_out=plugins=grpc:${output_dirpath}"

    for input_filepath in "${@}"; do
        # Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),
        #  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated
        # See also: https://github.com/golang/protobuf/issues/1272
        protobuf_filename="$(basename "${input_filepath}")"
        go_module_flag="--go_opt=M${protobuf_filename}=${go_bindings_pkg};$(basename "${go_bindings_pkg}")"

        if ! "${PROTOC_CMD}" \
                -I="${input_dirpath}" \
                "${grpc_flag}" \
                "${go_module_flag}" \
                "${@}"; then
            echo "Error: An error occurred generating Golang bindings for file '${input_filepath}'" >&2
            return 1
        fi
    done
}

# -------------------------------------------- Rust -----------------------------------------------
generate_rust_bindings() {
    input_dirpath="${1}"
    shift 1

    output_dirpath="${1}"
    shift 1

    if ! command -v "${RUST_BINDING_GENERATOR_CMD}"; then
        echo "Error: No '${RUST_BINDING_GENERATOR_CMD}' command found; you'll need to install it from https://github.com/kurtosis-tech/rust-protobuf-binding-generator" >&2
        return 1
    fi

    "${RUST_BINDING_GENERATOR_CMD}" "${input_dirpath}" "${output_dirpath}" "${@}"
}


# ------------------------------------------ Shared Code-----------------------------------------------
# Schema of the "object" that's the value of this map:
# relativeOutputDirpath|findSelectorMatchingGeneratedFiles|bindingGenerationFunc
# NOTE: the binding-generating function signature is as follows: input_dirpath output_dirpath input_filepath1 [input_filepath2...]
declare -A generators
generators["${GOLANG_DIRNAME}"]="${GO_RELATIVE_OUTPUT_DIRPATH}|-name '*.go'|generate_golang_bindings"
generators["${RUST_DIRNAME}"]="${RUST_RELATIVE_OUTPUT_DIRPATH}|-name '*.rs' ! -name 'mod.rs'|generate_rust_bindings"

input_dirpath="${root_dirpath}/${INPUT_RELATIVE_DIRPATH}"
for lang in "${!generators[@]}"; do
    lang_config_str="${generators["${lang}"]}"
    IFS='|' read -r -a lang_config_arr < <(echo "${lang_config_str}")

    rel_output_dirpath="${lang_config_arr[0]}"
    generated_files_selectors="${lang_config_arr[1]}"
    bindings_gen_func="${lang_config_arr[2]}"

    abs_output_dirpath="${root_dirpath}/${lang}/${rel_output_dirpath}"

    if [ "${abs_output_dirpath}/" != "/" ]; then
        if ! find "${abs_output_dirpath}" ${generated_files_selectors} -delete; then
            echo "Error: An error occurred removing the existing protobuf-generated code" >&2
            exit 1
        fi
    else
        echo "Error: output dirpath must not be empty!" >&2
        exit 1
    fi

    # NOTE: When multiple people start developing on this, we won't be able to rely on using the user's local protoc because they might differ. We'll need to standardize by:
    #  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
    #  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
    #  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
    #  4) Adding the locally-generated Go output files to .gitignore
    #  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
    input_filepaths="$(find "${input_dirpath}" -name "*.proto")"
    if ! "${bindings_gen_func}" "${input_dirpath}" "${abs_output_dirpath}" ${input_filepaths}; then
        echo "Error: An error occurred generating ${lang} bindings in directory '${abs_output_dirpath}' for files:" >&2
        for input_filepath in ${input_filepaths}; do
            echo " - ${input_filepath}" >&2
        done
        exit 1
    fi
done
