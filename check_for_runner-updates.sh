#!/bin/bash

# Script Name: check_for_runner-updates.sh 
# Purpose: Check for updates in a GitHub repository and sync the local copy

# Configuration
github_repo_url="https://github.com/willKman718/command-runner.git"
github_api_url="https://api.github.com/repos/willKman718/command-runner/commits?per_page=1"
local_repo_path="/opt/perforce/command-runner" 
ConfigFile=".update_config"

function msg() { echo -e "$*"; }
function bail() { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }

# Check for dependencies
for dep in curl git jq go; do
    which $dep > /dev/null || bail "Failed to find $dep in PATH."
done

# Ensure the directory exists, or create it
if [ ! -d "$local_repo_path" ]; then
    mkdir -p "$local_repo_path" || bail "Failed to create directory $local_repo_path"
fi

# Check if the repository exists locally
if [ ! -d "$local_repo_path/.git" ]; then
    msg "Repository not found locally. Cloning..."
    git clone "$github_repo_url" "$local_repo_path" || bail "Failed to clone repository."
    cd "$local_repo_path" || bail "Can't cd to $local_repo_path"
else
    cd "$local_repo_path" || bail "Can't cd to $local_repo_path"
fi

last_github_sha=""
if [[ -e "$ConfigFile" ]]; then
    last_github_sha=$(grep last_github_sha "$ConfigFile" | cut -d= -f2)
fi

github_sha=$(curl -s "$github_api_url" | jq -r '.[] | .sha')

if [[ "$last_github_sha" != "$github_sha" ]]; then
    msg "Updating project"
    git pull origin master || bail "Failed to pull updates from repository."

    if [ -f Makefile ]; then
        make || bail "Failed to compile after pulling updates."
    else
        msg "Makefile not found. Compilation skipped."
    fi

    echo "last_github_sha=$github_sha" > "$ConfigFile"
    msg "Project updated"
else
    msg "Project is up-to-date - nothing to do"
fi
