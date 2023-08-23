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

# Check if check_for_runner-updates.sh is untracked and if so, move it to /tmp
if git ls-files --others --exclude-standard | grep -q "check_for_runner-updates.sh"; then
    mv check_for_runner-updates.sh "/tmp/check_for_runner-updates_$(date +%Y%m%d%H%M%S).sh"
    echo "Moved local untracked check_for_runner-updates.sh to /tmp."
fi

if [[ "$last_github_sha" != "$github_sha" ]]; then
    msg "Updating project"

    # Backup the current report_instance_data.sh
    cp report_instance_data.sh report_instance_data.sh.bak

    # Stash any local changes to allow the git pull to work without conflicts
    git stash

    git pull origin master || bail "Failed to pull updates from repository."

    if [ -f Makefile ]; then
        make || bail "Failed to compile after pulling updates."
    else
        msg "Makefile not found. Compilation skipped."
    fi

# Extract the entire lines from the backup script
#old_config_line=$(grep 'declare ConfigFile=' report_instance_data.sh.bak | head -n 1)
#old_log_line=$(grep 'declare report_instance_logfile=' report_instance_data.sh.bak | head -n 1)

# Replace the lines in the new script
#sed -i "s#declare ConfigFile=.*#$old_config_line#" report_instance_data.sh
#sed -i "s#declare report_instance_logfile=.*#$old_log_line#" report_instance_data.sh


    # Ensure the scripts are executable
    chmod +x check_for_runner-updates.sh
    chmod +x report_instance_data.sh

    echo "last_github_sha=$github_sha" > "$ConfigFile"
    msg "Project updated"
    msg "Reporting in"
    /opt/perforce/command-runner/report_instance_data.sh >> /opt/perforce/command-runner/logs/report-instance-data.log 2>&1
    rm /tmp/out.json
else
    msg "Project is up-to-date - nothing to do"
fi
