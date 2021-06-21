#!/usr/bin/env bash

# This script regenerates bindings for the testsuite container API in the various languages that this repo supports

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
GENERATOR_SCRIPT_FILENAME="generate-protobuf-bindings.sh"  # Must be on the PATH
SUITE_API_DIRNAME="suite-api"
GOLANG_DIRNAME="golang"

# =============================== MAIN LOGIC =======================================================
input_dirpath="${root_dirpath}/${SUITE_API_DIRNAME}"

# Golang
go_output_dirpath="${root_dirpath}/${GOLANG_DIRNAME}/lib/rpc_api/bindings"
if ! GO_MOD_FILEPATH="${root_dirpath}/${GOLANG_DIRNAME}/go.mod" "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${go_output_dirpath}" golang; then
    echo "Error: An error occurred generating Go bindings in directory '${go_output_dirpath}'" >&2
    exit 1
fi
echo "Successfully generated Go bindings in directory '${go_output_dirpath}'"
