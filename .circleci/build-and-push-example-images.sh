set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==========================================================================================
#                                         Constants
# ==========================================================================================
KURTOSIS_DOCKERHUB_ORG="kurtosistech"
LANG_SCRIPTS_DIRNAME="scripts"
BUILD_AND_RUN_FILENAME="build-and-run.sh"

# ==========================================================================================
#                                        Arg-parsing
# ==========================================================================================
docker_username="${1:-}"
docker_password_DO_NOT_LOG="${2:-}" # WARNING: DO NOT EVER LOG THIS!!
circleci_git_tag="${3:-}"   # This should be mutually exclusive with the CircleCI Git branch
circleci_git_branch="${4:-}" # This should be mutually exclusive with the CircleCI Git tag

# ==========================================================================================
#                                        Arg validation
# ==========================================================================================
if [ -z "${docker_username}" ]; then
    echo "Error: Docker username cannot be empty" >&2
    exit 1
fi
if [ -z "${docker_password_DO_NOT_LOG}" ]; then
    echo "Error: Docker password cannot be empty" >&2
    exit 1
fi
if [ -z "${circleci_git_tag}" ] && [ -z "${circleci_git_branch}" ]; then
    echo "Error: Both CircleCI Git tag & branch were empty; either one or the other should be specified" >&2
    exit 1
fi
if [ -n "${circleci_git_tag}" ] && [ -n "${circleci_git_branch}" ]; then
    echo "Error: Both CircleCI Git tag & branch were specified; either one or the other should be specified" >&2
    exit 1
fi

# ==========================================================================================
#                                           Main code
# ==========================================================================================
# Docker is restricting anonymous image pulls, so we log in before we do any pulling
if ! docker login -u "${docker_username}" -p "${docker_password_DO_NOT_LOG}"; then
    echo "Error: Logging in to Docker failed" >&2
    exit 1
fi

if [ -n "${circleci_git_tag}" ]; then
    output_docker_tag_name="${circleci_git_tag}"
elif [ -n "${circleci_git_branch}" ]; then
    output_docker_tag_name="${circleci_git_branch}"
else
    echo "Error: Both CircleCI Git tag & branch were empty; this should never happen!" >&2
    exit 1
fi

echo "Pushing example testsuite Docker images to Dockerhub..."
supported_langs_filepath="${root_dirpath}/supported-languages.txt"
for lang in $(cat "${supported_langs_filepath}"); do
    echo "Building ${lang} Docker image..."
    buildscript_filepath="${root_dirpath}/${lang}/${LANG_SCRIPTS_DIRNAME}/${BUILD_AND_RUN_FILENAME}"
    if ! bash "${buildscript_filepath}" build; then
        echo "Error: Building example ${lang} image failed" >&2
        exit 1
    fi
    echo "Successfully built ${lang} Docker image"

    image_name="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-${lang}-example"
    full_image="${image_name}/${output_docker_tag_name}"
    if [ -n "${circleci_git_tag}" ]; then
        # When we run as the result of a tag build, the built image gets a tag called "HEAD" so we have to re-tag it to the expected tag name
        head_image="${image_name}:HEAD"
        if ! docker tag "${head_image}" "${full_image}"; then
            echo "Error: Could not re-tag Docker image ${head_image} -> ${full_image}" >&2
            exit 1
        fi
    fi

    echo "Pushing ${full_image} to Dockerhub..."
    if ! docker push ${full_image}; then
        echo "Error: Could not push Docker image '${full_image}' to Dockerhub" >&2
        exit 1
    fi
    echo "Successfully pushed ${full_image} to Dockerhub"
done
echo "Successfully pushed all example testsuite Docker images to Dockerhub"
