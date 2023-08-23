#!/bin/bash

# Variables
REPO_URL="https://github.com/willKman718/command-runner.git"
LOCAL_REPO_PATH="/opt/perforce/command-runner"
GO_VERSION="1.19"
COMMAND_RUNNER_LOG_DIR="/opt/perforce/command-runner/logs"
COMMAND_RUNNER_LOG="$COMMAND_RUNNER_LOG_DIR/command-runner.log"
SETUP_LOG="/tmp/command-runner_install.log"


function msg() { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$SETUP_LOG"; msg "$dt: $*"; }
function bail() { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }


# Ensure running with sudo
if [ "$EUID" -ne 0 ]; then
    log "Please run this script with sudo."
    exit
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
else
    log "Cannot detect OS type. Exiting..."
    exit 1
fi

log "Detected OS: $ID_LIKE ($ID)"


# Determine user existence
if id "perforce" &>/dev/null; then
    USER_NAME="perforce"
elif id "command-runner" &>/dev/null; then
    USER_NAME="command-runner"
else
    useradd -m -s /bin/bash command-runner && USER_NAME="command-runner" || {
        log "Failed to create user 'command-runner'. Exiting..."
        exit 1
    }
fi

chown $USER_NAME:$USER_NAME "$SETUP_LOG"

install_utility() {
    utility=$1

    if ! command -v $utility &> /dev/null; then
        echo "Installing $utility..."

        # Check if OS ID or any ID_LIKE value matches supported OSes
        if [[ " $ID $ID_LIKE " =~ " debian " || " $ID $ID_LIKE " =~ " ubuntu " ]]; then
            apt-get update
            apt-get install -y $utility
        elif [[ " $ID $ID_LIKE " =~ " centos " || " $ID $ID_LIKE " =~ " rhel " || " $ID $ID_LIKE " =~ " fedora " ]]; then
            yum install -y $utility
        else
            log "Unsupported OS. Exiting..."
            exit 1
        fi

        if [ "$utility" == "golang" ]; then
            command -v go &> /dev/null || {
            log "Go is not installed. Exiting..."
            exit 1
        }
        else
            command -v $utility &> /dev/null || {
            log "Failed to install $utility. Exiting..."
            exit 1
        }
        fi
    fi
}
install_golang() {
    current_version=$(go version 2>/dev/null | awk '{print $3}' | tr -d "go")

    if [ -z "$current_version" ] || [ "$(printf '%s\n' "$GO_VERSION" "$current_version" | sort -V | head -n1)" != "$GO_VERSION" ]; then

        # Detect and remove if Go was installed via package managers
        if [[ " $ID $ID_LIKE " =~ " debian " || " $ID $ID_LIKE " =~ " ubuntu " ]]; then
            apt-get purge -y golang*  # Using purge to remove configurations as well
        elif [[ " $ID $ID_LIKE " =~ " centos " || " $ID $ID_LIKE " =~ " rhel " || " $ID $ID_LIKE " =~ " fedora " ]]; then
            yum remove -y golang
        fi

        # The rest of the installation steps from the previous script snippet...
        echo "Installing Go version $GO_VERSION..."
        pushd /tmp > /dev/null
        wget https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz
        tar -xvf go$GO_VERSION.linux-amd64.tar.gz
        mv go /usr/local
        rm go$GO_VERSION.linux-amd64.tar.gz
        popd > /dev/null
        echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
        source /etc/profile
    fi
}

log "Starting utility installations."
install_utility git
install_utility curl
install_utility jq
install_utility make
install_golang #Currently always installing go #TODO FIX THIS
#install_utility golang

log "Utility installations completed."

# Clone and compile
log "Starting cloning and compilation process."
[ ! -d "$LOCAL_REPO_PATH" ] && git clone "$REPO_URL" "$LOCAL_REPO_PATH"
cd "$LOCAL_REPO_PATH" && {
    [ -f Makefile ] && make || log "Makefile not found. Skipping compilation."
} || {
    log "Failed to navigate to $LOCAL_REPO_PATH. Skipping compilation."
}

# Set permissions
log "Setting permissions for $LOCAL_REPO_PATH."
chown -R $USER_NAME:$USER_NAME "$LOCAL_REPO_PATH"
chmod +x "$LOCAL_REPO_PATH/check_for_runner-updates.sh" "$LOCAL_REPO_PATH/report_instance_data.sh" "$LOCAL_REPO_PATH/setup_config.sh"
# chmod +x "$LOCAL_REPO_PATH/setup_config.sh"

# Ensure log directory exists
if [ ! -d "$COMMAND_RUNNER_LOG_DIR" ]; then
    mkdir -p "$COMMAND_RUNNER_LOG_DIR"
    chown $USER_NAME:$USER_NAME "$COMMAND_RUNNER_LOG_DIR"
fi

# Call the setup_config.sh script
log "Setting up configuration with setup_config.sh..."
sudo -u $USER_NAME bash -i "$LOCAL_REPO_PATH/setup_config.sh"





# Function to check and update a cron job if needed
current_cron=$(crontab -u $USER_NAME -l 2>/dev/null || echo "")
update_cron() {
    local job="$1"
    local name="$2"
    if echo "$current_cron" | grep -qF "$name"; then
        if echo "$current_cron" | grep -qF "$job"; then
            log "$name job in crontab matches the desired configuration. No changes made."
        else
            log "$name job in crontab differs from the desired configuration. Updating..."
            # Remove the old job
            current_cron=$(echo -e "$current_cron" | grep -vF "$name")
            # Add the new job
            current_cron="$current_cron\n$job"
            log "Updated $name job in crontab."
        fi
    else
        current_cron="$current_cron\n$job"
        log "Added $name job to crontab."
    fi
}

# Ensure comment is in the crontab or add it if missing
COMMENT="# Command-Runner check for updates and report instance data"
if ! echo "$current_cron" | grep -qF "$COMMENT"; then
    current_cron="$current_cron\n$COMMENT"
    log "Added comment to crontab."
fi

# Check and potentially update the two cron jobs

# check_for_runner-updates.sh job
CRON_JOB_UPDATES="0 2 * * * $LOCAL_REPO_PATH/check_for_runner-updates.sh > /dev/null 2>&1 ||:"
update_cron "$CRON_JOB_UPDATES" "check_for_runner-updates.sh"

# report_instance_data.sh job
CRON_JOB_REPORT="10 0 * * * $LOCAL_REPO_PATH/report_instance_data.sh > /dev/null 2>&1 ||:"
update_cron "$CRON_JOB_REPORT" "report_instance_data.sh"

# Update the crontab with the potential changes
echo -e "$current_cron" | crontab -u $USER_NAME -
log "Crontab operations completed."



log "Reporting in"
sudo -u $USER_NAME bash -l -c /opt/perforce/command-runner/report_instance_data.sh > /dev/null 2>&1 ||:
rm /tmp/out.json
log "Installation complete!"

