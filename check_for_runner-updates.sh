#!/bin/bash

# Script Name: check_for_runner-updates.sh
# Purpose: Check for updates in a GitHub repository and sync the local copy
# Determine the home directory of the current user
export PATH=$PATH:/usr/local/go/bin

# Configuration
github_repo_url="https://github.com/willKman718/command-runner.git"
github_api_url="https://api.github.com/repos/willKman718/command-runner/commits?per_page=1"
local_repo_path="/opt/perforce/command-runner"
ConfigFile=".update_config"
LOG_FILE="/opt/perforce/command-runner/logs/command-runner.log"

function msg() { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$LOG_FILE"; msg "$dt: $*"; }
function bail() { msg "\nError: ${1:-Unknown Error}\n"; log "Error: ${1:-Unknown Error}\n"; exit ${2:-1}; }


# Check for dependencies
log "Checking for dependencies..."
for dep in curl git jq go; do
    which $dep > /dev/null || bail "Failed to find $dep in PATH."
done

# Ensure the directory exists, or create it
if [ ! -d "$local_repo_path" ]; then
    log "Creating directory $local_repo_path..."
    mkdir -p "$local_repo_path" || bail "Failed to create directory $local_repo_path"
fi

# Check if the repository exists locally
if [ ! -d "$local_repo_path/.git" ]; then
    log "Repository not found locally. Initiating cloning..."
    git clone "$github_repo_url" "$local_repo_path" || bail "Failed to clone repository."
    cd "$local_repo_path" || bail "Can't cd to $local_repo_path"
else
    cd "$local_repo_path" || bail "Can't cd to $local_repo_path"
fi


# Fetch the last known SHA from config if available
last_github_sha=""
if [[ -e "$ConfigFile" ]]; then
    last_github_sha=$(grep last_github_sha "$ConfigFile" | cut -d= -f2)
    log "Last GitHub SHA found in config: $last_github_sha"
else
    log "No last GitHub SHA found in config."
fi

github_sha=$(curl -s "$github_api_url" | jq -r '.[] | .sha')
[[ -z "$github_sha" ]] && bail "Failed to fetch latest SHA from GitHub."

# Handle the case of an untracked check_for_runner-updates.sh
if git ls-files --others --exclude-standard | grep -q "check_for_runner-updates.sh"; then
    mv check_for_runner-updates.sh "/tmp/check_for_runner-updates_$(date +%Y%m%d%H%M%S).sh"
    log "Moved local untracked check_for_runner-updates.sh to /tmp."
fi


if [[ "$last_github_sha" != "$github_sha" ]]; then
    log "Found updates on GitHub. Starting update process..."

    # Backup the current report_instance_data.sh
    log "Backing up report_instance_data.sh..."
    cp report_instance_data.sh report_instance_data.sh.bak || log "Warning: Failed to back up report_instance_data.sh. Continuing..."

    # Stashing local changes
    git stash || log "Warning: Failed to stash local changes. Continuing..."

    git pull origin master || bail "Failed to pull updates from repository."

    # Compile using Makefile if it exists
    if [ -f Makefile ]; then
        make || bail "Failed to compile after pulling updates."
    else
        log "Makefile not found. Compilation skipped."
    fi

    # Ensure the scripts are executable
    chmod +x check_for_runner-updates.sh
    chmod +x report_instance_data.sh

    echo "last_github_sha=$github_sha" > "$ConfigFile" || log "Warning: Failed to update config with the latest SHA. Continuing..."
    log "Updated config with latest GitHub SHA."
    log "Project updated"
    log "Reporting in"
    /opt/perforce/command-runner/report_instance_data.sh > /dev/null 2>&1 || log "Warning: report_instance_data.sh encountered an issue. Check report-instance-data.log for details."
    [[ -f "/tmp/out.json" ]] && rm /tmp/out.json || log "Warning: Failed to remove /tmp/out.json."

else
    log "Project is up-to-date - nothing to do"
fi
