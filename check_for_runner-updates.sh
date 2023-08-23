#!/bin/bash

# Script Name: check_for_runner-updates.sh
# Purpose: Check for updates in a GitHub repository and sync the local copy

# Configuration
github_repo_url="https://github.com/willKman718/command-runner.git"
github_api_url="https://api.github.com/repos/willKman718/command-runner/commits?per_page=1"
local_repo_path="/opt/perforce/command-runner"
ConfigFile=".update_config"
LOG_FILE="/opt/perforce/command-runner/logs/command-runner.log"

function msg() { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$LOG_FILE"; msg "$dt: $*"; }
function bail() { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }

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

# Check if check_for_runner-updates.sh is untracked and if so, move it to /tmp
if git ls-files --others --exclude-standard | grep -q "check_for_runner-updates.sh"; then
    mv check_for_runner-updates.sh "/tmp/check_for_runner-updates_$(date +%Y%m%d%H%M%S).sh"
    log "Moved local untracked check_for_runner-updates.sh to /tmp."
fi

if [[ "$last_github_sha" != "$github_sha" ]]; then
    log "Found updates on GitHub. Starting update process..."

    # Backup the current report_instance_data.sh
    log "Backing up report_instance_data.sh..."
    cp report_instance_data.sh report_instance_data.sh.bak

    # Stash any local changes to allow the git pull to work without conflicts
    git stash

    git pull origin master || bail "Failed to pull updates from repository."

    if [ -f Makefile ]; then
        make || bail "Failed to compile after pulling updates."
    else
        log "Makefile not found. Compilation skipped."
    fi

    # Ensure the scripts are executable
    log "Setting execute permissions on scripts..."
    chmod +x check_for_runner-updates.sh || log "Warning: Failed to set execute permissions on check_for_runner-updates.sh. Continuing..."
    chmod +x report_instance_data.sh || log "Warning: Failed to set execute permissions on report_instance_data.sh. Continuing..."


    echo "last_github_sha=$github_sha" > "$ConfigFile"
    log "Updated config with latest GitHub SHA."
    log "Project updated"
    log "Reporting in"
    /opt/perforce/command-runner/report_instance_data.sh >> /opt/perforce/command-runner/logs/report-instance-data.log 2>&1
    rm /tmp/out.json
else
    log "Project is up-to-date - nothing to do"
fi
