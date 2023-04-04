#!/bin/sh

set -euxo pipefail

# Workaround for https://github.com/actions/runner-images/issues/6775
git config --global --add safe.directory "$GITHUB_WORKSPACE"

echo "Running owners"
owners github --owners_file_name="$INPUT_OWNERS_FILE_NAME"
