set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==========================================================================================
#                                            Constants
# ==========================================================================================
DEVELOP_BRANCH="develop"
GITFLOW_PP_FILENAME="gitflow-pp.sh"
CHANGELOG_FILENAME="CHANGELOG.md"
CHANGELOG_TBD_LINE="# TBD"
CHANGELOG_TBD_LINE_PATTERN="^${CHANGELOG_TBD_LINE}$"
EXPECTED_NUM_VERSION_FRAGMENTS=3   # We expected X.Y.Z versions

# Rust
RUST_LANG_DIRNAME="rust"
RUST_LIB_PACKAGE_DIRNAME="lib"
RUST_TESTSUITE_PACKAGE_DIRNAME="testsuite"
CARGO_TOML_FILENAME="Cargo.toml"
CARGO_TOML_VERSION_LINE_PATTERN='^version = ".*$'



# ==========================================================================================
#                                    Pre-flight validation
# ==========================================================================================
# TODO: It really sucks that we require users to have:
# a) git-flow installed
# b) the gitflow-pp.sh script installed
# c) gitflow-pp on the PATH
# The longterm fix is to have a Kurtosis-internal tool that handles all our releases across all our repos, and which could have post-release processing hooks
if ! type -P "${GITFLOW_PP_FILENAME}" > /dev/null; then
    echo "Error: This release script requires gitflow-pp.sh to be installed *and* available on the PATH" >&2
    echo "To install gitflow-pp.sh, follow the instructions at the top of the file here: https://github.com/mieubrisse/dotfiles/blob/master/bash/utils/gitflow-pp.sh" >&2
    echo "You'll also need to install it on your PATH after installation" >&2
    exit 1
fi
gitflow_pp_filepath="$(type -P "${GITFLOW_PP_FILENAME}")"
changelog_filepath="${root_dirpath}/${CHANGELOG_FILENAME}"
if ! [ -f "${changelog_filepath}" ]; then
    echo "Error: No changelog file found at '${changelog_filepath}'" >&2
    exit 1
fi
num_tbd_lines="$(grep -c "${CHANGELOG_TBD_LINE_PATTERN}" "${changelog_filepath}")"
if [ "${num_tbd_lines}" -eq 0 ] || [ "${num_tbd_lines}" -gt 1 ]; then
    echo "Error: Expected exactly one line matching pattern '${CHANGELOG_TBD_LINE_PATTERN}' in '${changelog_filepath}' but found ${num_tbd_lines}" >&2
    exit 1
fi

# -------------------------------------------- Rust ----------------------------------------
# Verify that the Cargo.toml files have the line we expect
rust_lang_root_dirpath="${root_dirpath}/${RUST_LANG_DIRNAME}"
rust_cargo_toml_filepaths=(
    "${rust_lang_root_dirpath}/${RUST_LIB_PACKAGE_DIRNAME}/${CARGO_TOML_FILENAME}"
    "${rust_lang_root_dirpath}/${RUST_TESTSUITE_PACKAGE_DIRNAME}/${CARGO_TOML_FILENAME}"
)
for filepath in "${rust_cargo_toml_filepaths[@]}"; do
    if ! [ -f "${filepath}" ]; then
        echo "Error: Missing expected ${CARGO_TOML_FILENAME} at '${lib_cargo_toml_filepath}'" >&2
        exit 1
    fi
    num_version_lines="$(grep -c "${CARGO_TOML_VERSION_LINE_PATTERN}" "${filepath}")"
    if [ "${num_version_lines}" -eq 0 ] || [ "${num_version_lines}" -gt 1 ]; then
        echo "Error: Expected exactly one line matching pattern '${CARGO_TOML_VERSION_LINE_PATTERN}' in '${filepath}' but found ${num_version_lines}" >&2
        exit 1
    fi
done



# ==========================================================================================
#                                         Pre-release functions
# ==========================================================================================
function make_shared_pre_release_modifications() {
    new_version="${1}"

    # Update changelog
    new_version_line="# ${new_version}"
    if ! sed -i '' "s/${CHANGELOG_TBD_LINE_PATTERN}/${CHANGELOG_TBD_LINE}\n\n${new_version_line}/" "${changelog_filepath}"; then
        echo "Error: Could not sed TBD line '${CHANGELOG_TBD_LINE_PATTERN}' -> '${new_version_line}' in changelog file '${changelog_filepath}'" >&2
        return 1
    fi
}

function make_rust_pre_release_modifications() {
    new_version="${1}"

    # Frustratingly, Rust ONLY allows you to specify a crate's version in the Cargo.toml which means we need to 'sed' that file on every release
    new_version_line="version = \"${new_version}\"  # Do not modify; gets automatically updated during release!"
    for filepath in "${rust_cargo_toml_filepaths[@]}"; do
        if ! sed -i '' "s/${CARGO_TOML_VERSION_LINE_PATTERN}/${new_version_line}/" "${filepath}"; then
            echo "Error: Could not sed '${CARGO_TOML_VERSION_LINE_PATTERN}' -> '${new_version_line}' in ${CARGO_TOML_FILENAME} file '${filepath}'" >&2
            exit 1
        fi
    done
}

# NOTE: Go doesn't have any pre-release steps
pre_release_functions=(
    "make_shared_pre_release_modifications"
    "make_rust_pre_release_modifications"
)


# ==========================================================================================
#                                           Main code
# ==========================================================================================
new_version=""
while [ -z "${new_version}" ]; do
    read -p "What version should we release with? (X.Y.Z) " new_version_candidate
    IFS='.' read -ra candidate_version_fragments < <(echo "${new_version_candidate}")
    num_candidate_version_fragments="${#candidate_version_fragments[@]}"

    # Validate X.Y.Z format
    if [ "${num_candidate_version_fragments}" -ne "${EXPECTED_NUM_VERSION_FRAGMENTS}" ]; then
        echo "Error: Version must be in X.Y.Z format" >&2
        continue
    fi
    # Validate all fragments are numeric
    last_idx="$((num_candidate_version_fragments - 1))"
    for i in $(seq 0 "${last_idx}"); do
        candidate_fragment="${candidate_version_fragments[${i}]}"
        # This is a neat cross-shell-compatible trick to verify that a variable is a number
        # See: https://stackoverflow.com/questions/806906/how-do-i-test-if-a-variable-is-a-number-in-bash/806923
        if ! [ "${candidate_fragment}" -eq "${candidate_fragment}" ] > /dev/null 2>&1; then
            echo "Error: Version fragment #${i} was not a number" >&2
            continue 2 # The "2" tells Bash to continue on the outer loop
        fi
    done
    # Validate the tag doesn't already exist
    if git rev-parse "${new_version_candidate}" &> /dev/null; then
        echo "Error: Tag '${new_version_candidate}' already exists" >&2
        continue
    fi

    new_version="${new_version_candidate}"
done

read -p "Verification: release new version '${new_version}'? (ENTER to continue, Ctrl-C to quit)"

if ! git checkout "${DEVELOP_BRANCH}"; then
    echo "Error: Could not check out branch '${DEVELOP_BRANCH}'" >&2
    exit 1
fi
if ! bash "${gitflow_pp_filepath}" release start; then
    echo "Error: Could not start release" >&2
    exit 1
fi

for pre_release_function in "${pre_release_functions[@]}"; do
    if ! "${pre_release_function}" "${new_version}"; then
        echo "Error: A failure occurred executing pre-release function '${pre_release_function}'" >&2
        exit 1
    fi
done

git add -u "${root_dirpath}"
git commit -m "Pre-release changes for version ${new_version}"

finish_release_cmd="bash ${gitflow_pp_filepath} release finish"
push_release_cmd="bash ${gitflow_pp_filepath} push"
if ! ${finish_release_cmd}; then
    echo "Error: Could not finish release; you'll need to manually run '${finish_release_cmd}' to finish the release followed by '${push_release_cmd}' to push it to remote" >&2
    exit 1
fi
# TODO DEBUGGING
if ! echo ${push_release_cmd}; then
    echo "Error: Could not push release; you'll need to manually run '${push_release_cmd}' to push it to remote" >&2
    exit 1
fi
