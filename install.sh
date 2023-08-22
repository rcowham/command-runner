#!/bin/bash

# Variables
REPO_URL="https://github.com/willKman718/command-runner.git"
LOCAL_REPO_PATH="/opt/perforce/command-runner"
GO_VERSION="1.17"
COMMAND_RUNNER_LOG_DIR="/opt/perforce/command-runner/logs"
COMMAND_RUNNER_LOG="$COMMAND_RUNNER_LOG_DIR/command-runner.log"


function msg() { echo -e "$*"; }
function bail() { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }

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

        if [ "$utility" == "golang" ]; then
            command -v go &> /dev/null || {
            echo "Go is not installed. Exiting..."
            exit 1
        }
        else
            command -v $utility &> /dev/null || {
            echo "Failed to install $utility. Exiting..."
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
        popd > /dev/null
        echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
        source /etc/profile
    fi
}


install_utility git
install_utility curl
install_utility jq
install_utility make
install_golang #Currently always installing go #TODO FIX THIS
#install_utility golang

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



# Set permissions
chown -R $USER_NAME:$USER_NAME "$LOCAL_REPO_PATH"
chmod +x "$LOCAL_REPO_PATH/check_for_runner-updates.sh" "$LOCAL_REPO_PATH/report_instance_data.sh"
chmod +x "$LOCAL_REPO_PATH/setup_config.sh"

# Call the setup_config.sh script
echo "Setting up configuration with setup_config.sh..."
bash -i "$LOCAL_REPO_PATH/setup_config.sh"


# Ensure log directory exists
if [ ! -d "$COMMAND_RUNNER_LOG_DIR" ]; then
    mkdir -p "$COMMAND_RUNNER_LOG_DIR"
    chown $USER_NAME:$USER_NAME "$COMMAND_RUNNER_LOG_DIR"
fi

# Check the current cron jobs for the specific user
current_cron=$(crontab -u $USER_NAME -l 2>/dev/null)

# Check if comment exists
if ! echo "$current_cron" | grep -q "# Command-Runner check for updates and report instance data"; then
    COMMENT="# Command-Runner check for updates and report instance data"
    (echo "$current_cron"; echo "$COMMENT") | crontab -u $USER_NAME -
fi

# Add the job if not found
if ! echo "$current_cron" | grep -q "check_for_runner-updates.sh"; then
    CRON_JOB_CONTENT="0 2 * * * $LOCAL_REPO_PATH/check_for_runner-updates.sh >> $COMMAND_RUNNER_LOG 2>&1"
    (echo "$current_cron"; echo "$CRON_JOB_CONTENT") | crontab -u $USER_NAME -
fi

if ! echo "$current_cron" | grep -q "report_instance_data.sh"; then
    CRON_JOB_CONTENT="10 0 * * * $LOCAL_REPO_PATH/report_instance_data.sh >> $COMMAND_RUNNER_LOG 2>&1"
    (echo "$current_cron"; echo "$CRON_JOB_CONTENT") | crontab -u $USER_NAME -
fi


# Save the current cron jobs, then append the new jobs (if they don't already exist), and reload them
(crontab -u $USER_NAME -l 2>/dev/null | grep -v -E "check_for_runner-updates.sh|report_instance_data.sh"; echo "$CRON_JOB_CONTENT") | crontab -u $USER_NAME -

msg "Reporting in"
/opt/perforce/command-runner/report_instance_data.sh >> $COMMAND_RUNNER_LOG 2>&1
echo "Installation complete!"
