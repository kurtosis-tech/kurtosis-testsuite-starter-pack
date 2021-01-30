TESTSUITE_IMPL_DIRNAME="testsuite"
README_FILENAME="README.md"
DOCKERIGNORE_FILENAME=".dockerignore"

# Constants 
GO_MOD_FILENAME="go.mod"
GO_MOD_MODULE_KEYWORD="module "  # The key we'll look for when replacing the module name in go.mod
BUILDSCRIPT_FILENAME="build_and_run.sh"
DOCKER_IMAGE_VAR_KEYWORD="SUITE_IMAGE=" # The variable we'll look for in the Docker file for replacing the Docker image name
IS_KURTOSIS_CORE_DEV_MODE_VAR_KEYWORD="IS_KURTOSIS_CORE_DEV_MODE=" # The variable we'll look for when setting whether Kurtosis core dev mode is enabled

# Response required from the user to do the bootstrapping
BOOTSTRAP_VERIFICATION_RESPONSE="create new repo"

set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
buildscript_filepath="${root_dirpath}/scripts/${BUILDSCRIPT_FILENAME}"
go_mod_filepath="${root_dirpath}/${GO_MOD_FILENAME}"

# ============== Validation =================================================================
# Validation, to save us in case someone changes stuff in the future
if [ "$(grep "${GO_MOD_MODULE_KEYWORD}" "${go_mod_filepath}" | wc -l)" -ne 1 ]; then
    echo "Validation failed: Could not find exactly one line in ${GO_MOD_FILENAME} with keyword '${GO_MOD_MODULE_KEYWORD}' for use when replacing with the user's module name" >&2
    exit 1
fi
if [ "$(grep "^${DOCKER_IMAGE_VAR_KEYWORD}" "${buildscript_filepath}" | wc -l)" -ne 1 ]; then
    echo "Validation failed: Could not find exactly one line in ${buildscript_filepath} starting with keyword '${DOCKER_IMAGE_VAR_KEYWORD}' for use when replacing with the user's Docker image name" >&2
    exit 1
fi
if [ "$(grep "^${IS_KURTOSIS_CORE_DEV_MODE_VAR_KEYWORD}" "${buildscript_filepath}" | wc -l)" -ne 1 ]; then
    echo "Validation failed: Could not find exactly one line in ${buildscript_filepath} starting with keyword '${IS_KURTOSIS_CORE_DEV_MODE_VAR_KEYWORD}' for use when setting the Kurtosis Core dev mode to false" >&2
    exit 1
fi

# ============== Inputs & Verification =================================================================
prompt_response=""
while [ "${prompt_response}" != "${BOOTSTRAP_VERIFICATION_RESPONSE}" ]; do
    read -p "This script should only be run if you're trying to create a new testsuite repo! To verify this is what you want, enter '${BOOTSTRAP_VERIFICATION_RESPONSE}': " prompt_response
done
new_module_name=""
while [ -z "${new_module_name}" ]; do
    read -p "Name for the Go module that will contain your testsuite project (e.g. github.com/my-org/my-repo): " new_module_name
done
docker_image_name=""
while [ -z "${docker_image_name}" ]; do
    echo "Name for the Docker image that this repo will build, which must conform to the Docker image naming rules:"
    echo "  https://docs.docker.com/engine/reference/commandline/tag/#extended-description"
    read -p "Image name (e.g. my-dockerhub-org/my-image-name): " docker_image_name
done


# ============== Main Code =================================================================
find "${root_dirpath}" \
    ! -name bootstrap \
    ! -name "${TESTSUITE_IMPL_DIRNAME}" \
    ! -name "${GO_MOD_FILENAME}" \
    ! -name go.sum \
    ! -name scripts \
    -mindepth 1 \
    -maxdepth 1 \
    -exec rm -rf {} \;

cp "${script_dirpath}/${README_FILENAME}" "${root_dirpath}/"
cp "${script_dirpath}/${DOCKERIGNORE_FILENAME}" "${root_dirpath}/"   # build_and_run requires a .dockerignore file, for best practice

# Replace module names in code (we need the "-i '' " argument because Mac sed requires it)
existing_module_name="$(grep "module" "${go_mod_filepath}" | awk '{print $2}')"
sed -i '' "s,${existing_module_name},${new_module_name},g" ${go_mod_filepath}
# We search for old_module_name/testsuite because we don't want the old_module_name/lib entries to get renamed
sed -i '' "s,${existing_module_name}/${TESTSUITE_IMPL_DIRNAME},${new_module_name}/${TESTSUITE_IMPL_DIRNAME},g" $(find "${root_dirpath}" -type f)

# Replace Docker image name in buildscript
sed -i '' "s,^${DOCKER_IMAGE_VAR_KEYWORD}.*,${DOCKER_IMAGE_VAR_KEYWORD}\"${docker_image_name}\"," "${buildscript_filepath}"

# Set Kurtosis Core dev mode to false in buildscript
sed -i '' "s,^${IS_KURTOSIS_CORE_DEV_MODE_VAR_KEYWORD}.*,${IS_KURTOSIS_CORE_DEV_MODE_VAR_KEYWORD}\"false\"," "${buildscript_filepath}"

rm -rf "${script_dirpath}"
echo "Bootstrap complete; view the README.md in ${root_dirpath} for next steps"
