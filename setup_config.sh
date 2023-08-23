#!/bin/bash

DEFAULT_DIR="/p4/common/config"
DEFAULT_CFG_FILE=".push_metrics.cfg"
DEFAULT_CFG_PATH="$DEFAULT_DIR/$DEFAULT_CFG_FILE"
DEFAULT_LOG_PATH="/opt/perforce/command-runner/logs/command-runner.log"
#CONFIG_CHANGED=0
#LOG_CHANGED=0
SETUP_LOG="/tmp/command-runner_setup.log"

function msg () { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$SETUP_LOG"; msg "$dt: $*"; }
function bail () { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }

# Check if the default configuration file exists
if [ -f "$DEFAULT_CFG_PATH" ]; then
    read -p "$DEFAULT_CFG_PATH already exists. Do you want to modify it? [Y/n]: " modify_choice
    if [[ "$modify_choice" =~ ^[Nn]$ ]]; then
        exit 0
    fi
    # Backup Existing Config
    BACKUP_PATH="${DEFAULT_CFG_PATH}_backup_$(date +%Y%m%d%H%M%S)"
    cp "$DEFAULT_CFG_PATH" "$BACKUP_PATH" || {
        log "Failed to create backup. Exiting..."
        exit 1
    }
    log "Backup of existing configuration saved to $BACKUP_PATH"
fi

# If the default directory exists, but the file doesn't
if [ -d "$DEFAULT_DIR" ]; then
    read -p "$DEFAULT_DIR exists but $DEFAULT_CFG_FILE does not. Would you like to use this directory? [Y/n]: " use_default
    
    # If user presses Enter without giving an input or provides 'Y'/'y', use the default path
    if [[ -z "$use_default" || "$use_default" =~ ^[Yy]$ ]]; then
        TARGET_CFG_PATH="$DEFAULT_CFG_PATH"
    else
        read -p "Full path to $DEFAULT_CFG_FILE config file: [default: $DEFAULT_CFG_PATH]:" custom_dir
        TARGET_CFG_PATH="$custom_dir/$DEFAULT_CFG_FILE"
        CONFIG_CHANGED=1
    fi
else
    read -p "$DEFAULT_DIR does not exist. Full path to $DEFAULT_CFG_FILE config file: [default: $DEFAULT_CFG_PATH]:" custom_dir
    if [[ -z "$custom_dir" ]]; then
        TARGET_CFG_PATH="$DEFAULT_CFG_PATH"  # Use default if no input provided
    else
        TARGET_CFG_PATH="$custom_dir/$DEFAULT_CFG_FILE"
        CONFIG_CHANGED=1
    fi
fi

# Create configuration based on user input
echo "Setting up configuration in $TARGET_CFG_PATH..."

while true; do
    read -p "Enter metrics host (e.g., http://some.ip.or.host:9091): " metrics_host
    if [[ "$metrics_host" =~ ^http(s)?:// ]]; then
        break
    else
        msg "Please enter a valid URL."
    fi
done

while true; do
    read -p "Enter metrics customer (e.g., Customer_Name or Customer-Name): " metrics_customer

    if [[ "$metrics_customer" =~ [^a-zA-Z0-9_-] ]]; then
        echo "Error: Only alphanumeric characters, hyphens, and underscores are allowed. Try again."
    else
        break
    fi
done


read -p "Enter metrics instance (default is hostname: $(hostname)): " metrics_instance

if [ -z "$metrics_instance" ]; then
    metrics_instance=$(hostname)
fi

read -p "Enter metrics user (e.g., username-for-pushgateway): " metrics_user

while true; do
    read -s -p "Enter metrics password: " metrics_passwd
    echo
    read -s -p "Re-enter password for confirmation: " metrics_passwd_confirm
    echo
    if [ "$metrics_passwd" == "$metrics_passwd_confirm" ]; then
        break
    else
        msg "Passwords do not match. Try again."
    fi
done

read -p "Enter report instance logfile path [default: $DEFAULT_LOG_PATH]: " report_instance_logfile

# Set the default for the logfile path if not provided
#if [ -z "$report_instance_logfile" ]; then
#    report_instance_logfile="$DEFAULT_LOG_PATH"
#else
#    LOG_CHANGED=1
#fi

# Ensure the directory exists
CONFIG_DIR="$(dirname "$TARGET_CFG_PATH")"
mkdir -p "$CONFIG_DIR" || {
    log "Failed to create directory $CONFIG_DIR. Exiting..."
    exit 1
}

# Write the configuration to the file
{
    cat << EOF > "$TARGET_CFG_PATH"
metrics_host=$metrics_host
metrics_customer=$metrics_customer
metrics_instance=$metrics_instance
metrics_user=$metrics_user
metrics_passwd=$metrics_passwd
report_instance_logfile=$report_instance_logfile
EOF
} || {
    log "Failed to write to $TARGET_CFG_PATH. Exiting..."
    exit 1
}

chmod 600 "$TARGET_CFG_PATH"

msg "Configuration setup completed!"
log "Configuration saved to $TARGET_CFG_PATH."

# The commented-out sections remain in case you want to integrate them later
#if [ "$CONFIG_CHANGED" -eq 1 ]; then
#    log "Updating config path in report_instance_data.sh"
#    sed -i "s|ConfigFile=\"/p4/common/config/.push_metrics.cfg\"|ConfigFile=\"$TARGET_CFG_PATH\"|" ./report_instance_data.sh
#fi

#if [ "$LOG_CHANGED" -eq 1 ]; then
#
