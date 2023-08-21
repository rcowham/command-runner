#!/bin/bash

DEFAULT_DIR="/p4/common/config"
DEFAULT_CFG_FILE=".push_metrics.cfg"
DEFAULT_CFG_PATH="$DEFAULT_DIR/$DEFAULT_CFG_FILE"
DEFAULT_LOG_PATH="/p4/1/logs/report_instance_data.log"
CONFIG_CHANGED=0
LOG_CHANGED=0

# Check if the default configuration file exists
if [ -f "$DEFAULT_CFG_PATH" ]; then
    echo "$DEFAULT_CFG_PATH already exists. Exiting..."
    exit 0
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
    fi
    CONFIG_CHANGED=1
fi

# Create configuration based on user input
echo "Setting up configuration in $TARGET_CFG_PATH..."

read -p "Enter metrics host (e.g., http://some.ip.or.host:9091): " metrics_host
read -p "Enter metrics customer (e.g., Customer-Name): " metrics_customer
read -p "Enter metrics instance (e.g., server-name): " metrics_instance
read -p "Enter metrics user (e.g., username-for-pushgateway): " metrics_user
read -s -p "Enter metrics password: " metrics_passwd
echo
read -p "Enter report instance logfile path [default: $DEFAULT_LOG_PATH]: " report_instance_logfile

# Set the default for the logfile path if not provided
if [ -z "$report_instance_logfile" ]; then
    report_instance_logfile="$DEFAULT_LOG_PATH"
else
    LOG_CHANGED=1

fi

# Ensure the directory exists
CONFIG_DIR="$(dirname "$TARGET_CFG_PATH")"
mkdir -p "$CONFIG_DIR" || {
    echo "Failed to create directory $CONFIG_DIR. Exiting..."
    exit 1
}

# Write the configuration to the file
cat << EOF > "$TARGET_CFG_PATH"
metrics_host=$metrics_host
metrics_customer=$metrics_customer
metrics_instance=$metrics_instance
metrics_user=$metrics_user
metrics_passwd=$metrics_passwd
report_instance_logfile=$report_instance_logfile
EOF

echo "Configuration saved to $TARGET_CFG_PATH."

if [ "$CONFIG_CHANGED" -eq 1 ]; then
    echo "Updating config path in report_instance_data.sh"
    # Update the report_instance_data.sh script with the new config path
    sed -i "s|ConfigFile=\"/p4/common/config/.push_metrics.cfg\"|ConfigFile=\"$TARGET_CFG_PATH\"|" ./report_instance_data.sh
fi

if [ "$LOG_CHANGED" -eq 1 ]; then
    echo "Updating log path in report_instance_data.sh"
    # Update the report_instance_data.sh script with the new log path
    sed -i "s|declare report_instance_logfile=\"/p4/1/logs/report_instance_data.log\"|declare report_instance_logfile=\"$report_instance_logfile\"|" ./report_instance_data.sh
else
    echo "Log paths remain unchanged. No need to update report_instance_data.sh."
fi
