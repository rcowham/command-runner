#!/bin/bash

# Variables
REPO_URL="https://github.com/willKman718/command-runner.git"
LOCAL_REPO_PATH="/opt/perforce/command-runner"
GO_VERSION="1.17"

# Ensure running with sudo
if [ "$EUID" -ne 0 ]; then
    echo "Please run this script with sudo."
    exit
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
else
    echo "Cannot detect OS type. Exiting..."
    exit 1
fi

echo "Detected OS: $ID_LIKE ($ID)"

# Determine user existence
if id "perforce" &>/dev/null; then
    USER_NAME="perforce"
elif id "command-runner" &>/dev/null; then
    USER_NAME="command-runner"
else
    useradd -m -s /bin/bash command-runner && USER_NAME="command-runner" || {
        echo "Failed to create user 'command-runner'. Exiting..."
        exit 1
    }
fi

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
            echo "Unsupported OS. Exiting..."
            exit 1
        fi

        command -v $utility &> /dev/null || {
            echo "Failed to install $utility. Exiting..."
            exit 1
        }
    fi
}
install_golang() {
    if ! command -v go &> /dev/null; then
        echo "Go (golang) not found, installing..."
        wget https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz
        tar -xvf go$GO_VERSION.linux-amd64.tar.gz
        mv go /usr/local
        echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
        source /etc/profile
    fi
}

install_utility git
install_utility curl
install_utility jq
install_utility make
install_golang #Currently always installing go #TODO FIX THIS

# Clone and compile
[ ! -d "$LOCAL_REPO_PATH" ] && git clone "$REPO_URL" "$LOCAL_REPO_PATH"
cd "$LOCAL_REPO_PATH" && {
    [ -f Makefile ] && make || echo "Makefile not found. Skipping compilation."
} || {
    echo "Failed to navigate to $LOCAL_REPO_PATH. Skipping compilation."
}

# Set permissions
chown -R $USER_NAME:$USER_NAME "$LOCAL_REPO_PATH"
chmod +x "$LOCAL_REPO_PATH/check_for_runner-updates.sh" "$LOCAL_REPO_PATH/report_instance_data.sh"

# Setup cron jobs
CRON_JOB_CONTENT="# Command-Runner check for updates and report instance data
0 2 * * * $LOCAL_REPO_PATH/check_for_runner-updates.sh >> /var/log/command-runner.log 2>&1
10 0 * * * $LOCAL_REPO_PATH/report_instance_data.sh >> /var/log/report-instance-data.log 2>&1"

# Save the current cron jobs, then append the new jobs (if they don't already exist), and reload them
(crontab -u $USER_NAME -l 2>/dev/null | grep -v -E "check_for_runner-updates.sh|report_instance_data.sh"; echo "$CRON_JOB_CONTENT") | crontab -u $USER_NAME -

echo "Installation complete!"
