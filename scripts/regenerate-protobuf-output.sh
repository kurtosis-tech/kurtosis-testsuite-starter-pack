# This script regenerates Go bindings corresponding to the .proto files that define the API container's API
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
# Relative to THE ROOT OF THE ENTIRE REPO
INPUT_RELATIVE_DIRPATH="core-api"

# -------------------------- Golang ---------------------------
GOLANG_DIRNAME="golang"
GO_MOD_FILENAME="go.mod"
GO_MOD_FILE_MODULE_KEYWORD="module"
# Relative to the root of THE LANG DIR!!!!
GO_RELATIVE_OUTPUT_DIRPATH="lib/core_api_bindings"

# =============================== MAIN LOGIC =======================================================
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

generate_go_protoc_args() {
    input_filepath="${1}"
    output_dirpath="${2}"

    protobuf_filename="$(basename "${input_filepath}")"

    echo "--go_out=plugins=grpc:${output_dirpath}"

    # Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),
    #  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated
    # See also: https://github.com/golang/protobuf/issues/1272
    echo "--go_opt=M${protobuf_filename}=${go_bindings_pkg};$(basename "${go_bindings_pkg}")"
}

# Schema of the "object" that's the value of this map:
# relativeOutputDirpath|patternMatchingGeneratedFiles|additionalProtocArgsGeneratingFunc
# NOTE: the protoc args-generating function takes in two args: 1) the input filepath and 2) output dirpath
declare -A generators
generators["${GOLANG_DIRNAME}"]="${GO_RELATIVE_OUTPUT_DIRPATH}|*.go|generate_go_protoc_args"


input_dirpath="${root_dirpath}/${INPUT_RELATIVE_DIRPATH}"
for lang in "${!generators[@]}"; do
    lang_config_str="${generators["${lang}"]}"
    IFS='|' read -r -a lang_config_arr < <(echo "${lang_config_str}")

    rel_output_dirpath="${lang_config_arr[0]}"
    generated_files_pattern="${lang_config_arr[1]}"
    protoc_args_gen_func="${lang_config_arr[2]}"

    abs_output_dirpath="${root_dirpath}/${lang}/${rel_output_dirpath}"

    if [ "${abs_output_dirpath}/" != "/" ]; then
        if ! find "${abs_output_dirpath}" -name "${generated_files_pattern}" -delete; then
            echo "Error: An error occurred removing the existing protobuf-generated code" >&2
            exit 1
        fi
    else
        echo "Error: output dirpath must not be empty!" >&2
        exit 1
    fi

    for protobuf_filepath in $(find "${input_dirpath}" -name "*.proto"); do

        additional_protoc_args="$(eval "${protoc_args_gen_func} ${protobuf_filepath} ${abs_output_dirpath}")"

        # NOTE: When multiple people start developing on this, we won't be able to rely on using the user's local protoc because they might differ. We'll need to standardize by:
        #  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
        #  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
        #  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
        #  4) Adding the locally-generated Go output files to .gitignore
        #  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
        if ! protoc -I="${root_dirpath}/${INPUT_RELATIVE_DIRPATH}" ${additional_protoc_args} "${protobuf_filepath}"; then
            echo "Error: An error occurred generating ${lang} bindings from protobuf file: ${protobuf_filepath}" >&2
            exit 1
        fi
    done
done
