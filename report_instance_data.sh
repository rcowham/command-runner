#!/bin/bash
cmdrnrscriptDir=$(dirname "$0")
commandRunnerPath="$cmdrnrscriptDir/command-runner"
echo "Command Runner Path: $commandRunnerPath"
TempLog="/tmp/out.json"
rm -f $TempLog
# report_instance_data.sh
#
# Collects basic instance metadata about a customer environment (for AWS and Azure and ultimately other cloud envs)
#
# If used, put this job into perforce user crontab:
#
#   10 0 * * * /p4/common/site/bin/report_instance_data.sh -c /p4/common/config/.push_metrics.cfg > /dev/null 2>&1 ||:
#
# You can specify a config file as above, with expected format the same as for push_metrics.sh
#
# Uses AWS metadata URLs as defined: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
#
# Please note you need to make sure that the specified directory below (which may be linked)
# can be read by the node_exporter user (and is setup via --collector.textfile.directory parameter)
#
#TODO Better logging
#TODO NEEDS AZURE testing
# ============================================================
# Configuration section
# Find out if we're in AWS, GCP, or AZURE..

declare -i autoCloud=0
declare -i p4dInstalled=0
declare -i p4dRunning=0
declare -i swarmRunning=0
declare -i hasRunning=0

#This scripts default config file location
declare ConfigFile="/p4/common/config/.push_metrics.cfg"

## example .push_metrics.cfg
# ----------------------
# metrics_host=http://some.ip.or.host:9091
# metrics_customer=Customer-Name
# metrics_instance=server-name
# metrics_user=username-for-pushgateway
# metrics_passwd=password-for-pushgateway
# report_instance_logfile=/log/file/location
# metrics_cloudtype=AWS,GCP,AZure
# ----------------------

# May be overwritten in the config file.
#declare report_instance_logfile="/p4/1/logs/report_instance_data.log"
declare report_instance_logfile="/opt/perforce/command-runner/logs/command-runner.log"

### Auto Cloud Configs
### Timeout in seconds until we're done attempting to contact internal cloud information
autoCloudTimeout=5

declare ThisScript=${0##*/}

function msg () { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$report_instance_logfile"; msg "$dt: $*"; }
function bail () { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }
#function upcfg () { msg "metrics_cloudtype=$1" >> "$ConfigFile"; } #TODO This could be way more elegant IE error checking the config file but it works

upcfg() {
    local metrics_cloudtype="$1"

    # Log the action
    log "Updating ConfigFile with metrics_cloudtype=$metrics_cloudtype"

    # Attempt to append to the ConfigFile and log any potential errors
    echo "metrics_cloudtype=$metrics_cloudtype" >> "$ConfigFile" 2>> "$report_instance_logfile"
    if [ $? -ne 0 ]; then
        log "Error updating $ConfigFile with metrics_cloudtype=$metrics_cloudtype"
    fi
}

#run command-runner run!
run_command-runner() {
    local instance_arg="$1"
    shift

    if [[ "$instance_arg" == "--instance="* ]]; then
        local instance_value=${instance_arg#--instance=}
        local command="$commandRunnerPath $@ --output=$TempLog $instance_arg"
    else
        local command="$commandRunnerPath $instance_arg $@ --output=$TempLog"
    fi

    log "COMMAND-RUNNER: $command"   # Logging the command being run

    # Capturing command output and errors
    local cmd_output
    cmd_output=$(eval "$command" 2>&1) 
    local status=$?

    # Logging the command output
    log "  - COMMAND-RUNNER output: $cmd_output"
    while IFS= read -r line; do
        log "    $line"
    done <<< "$cmd_output"
    if [ $status -ne 0 ]; then
        log "Error executing command. Exit code: $status"  # Logging the error
        exit $status
    fi
}

# Work instances here
function work_instance () {
    local instance="$1"
    source /p4/common/bin/p4_vars $instance
    file_path="$P4CCFG/p4_$instance.vars"
    log "Working instance labeled as: $instance"
    # Your processing logic for each instance goes here
    {
        run_command-runner --instance=$instance
    }
}


# Instance Counter
# Thanks to ttyler below
function get_sdp_instances () {
    log "Finding p4d instances"
    SDPInstanceList=
    cd /p4 || bail "Could not cd to /p4."
    for e in *; do
        if [[ -r "/p4/$e/root/db.counters" ]]; then
            SDPInstanceList+=" $e"
        fi
    done

    # Trim leading space.
    # shellcheck disable=SC2116
    SDPInstanceList=$(echo "$SDPInstanceList")

    # Count instances
    instance_count=$(echo "$SDPInstanceList" | wc -w)

    # Loop through each instance and call the process_instance function
    for instance in $SDPInstanceList; do
        work_instance $instance
    done
}


findP4D() {
    # Check if p4d is installed
    if ! command -v p4d >/dev/null; then
        log "p4d is not installed."
        p4dInstalled=0
        return
    else
        log "p4d is installed."
        p4dInstalled=1
    fi
    # Function to check if a p4d process is running
    if pgrep -f "p4d_*" >/dev/null; then
        log "p4d service is running."
        p4dRunning=1
    else
        log "p4d service is not running."
    fi
}

function usage () {
    local style=${1:-"-h"}  # Default to "-h" if no style argument provided
    local errorMessage=${2:-"Unset"}
    if [[ "$errorMessage" != "Unset" ]]; then
        echo -e "\n\nUsage Error:\n\n$errorMessage\n\n" >&2
    fi

    echo "USAGE for $ThisScript:

$ThisScript -c <config_file> [-azure|-gcp]

    or

$ThisScript -h

    -azure      Specifies to collect Azure specific data
    -aws        Specifies to collect GCP specific data
    -gcp        Specifies to collect GCP specific data
    -acoff      Turns autoCloud off
    -acon       Turns autoCloud on
    -timeout    Sets timeout(In seconds) to wait for Cloud provider responds (default is $autoCloudTimeout seconds)

Collects metadata about the current instance and pushes the data centrally.

This is not normally required on customer machines. It assumes an SDP setup."
}

# Command Line Processing

declare -i shiftArgs=0

set +u
while [[ $# -gt 0 ]]; do
    case $1 in
        (-h) usage -h && exit 0;;
        # (-man) usage -man;;
        (-c) ConfigFile=$2; shiftArgs=1;;
        (-azure) IsAzure=1; IsGCP=0; IsAWS=0; autoCloud=0; echo "Forced GCP by -azure";;
        (-aws) IsAWS=1; IsGCP=0; IsAzure=0; autoCloud=0; echo "Forced GCP by -aws";;
        (-gcp) IsGCP=1; IsAWS=0; IsAzure=0; autoCloud=0; echo "Forced GCP by -gcp";;
        (-acoff) autoCloud=3; echo "AutoCloud turned OFF";;
        (-acon) autoCloud=1; echo "AutoCloud turned ON";;
        (-timeout) shift; autoCloudTimeout=$1; echo "Setting autoCloudTimeout to $autoCloudTimeout";;
        (-*) usage -h "Unknown command line option ($1)." && exit 1;;
    esac

    # Shift (modify $#) the appropriate number of times.
    shift; while [[ "$shiftArgs" -gt 0 ]]; do
        [[ $# -eq 0 ]] && usage -h "Incorrect number of arguments."
        shiftArgs=$shiftArgs-1
        shift
    done
done
set -u

[[ -f "$ConfigFile" ]] || bail "Can't find config file: ${ConfigFile}!"

# Get config values from config file- format: key=value
metrics_host=$(grep metrics_host "$ConfigFile" | awk -F= '{print $2}')
metrics_customer=$(grep metrics_customer "$ConfigFile" | awk -F= '{print $2}')
metrics_instance=$(grep metrics_instance "$ConfigFile" | awk -F= '{print $2}')
metrics_user=$(grep metrics_user "$ConfigFile" | awk -F= '{print $2}')
metrics_passwd=$(grep metrics_passwd "$ConfigFile" | awk -F= '{print $2}')
metrics_logfile=$(grep metrics_logfile "$ConfigFile" | awk -F= '{print $2}')
report_instance_logfile=$(grep report_instance_logfile "$ConfigFile" | awk -F= '{print $2}')
metrics_cloudtype=$(grep metrics_cloudtype "$ConfigFile" | awk -F= '{print $2}')
# Set all thats not set to Unset
metrics_host=${metrics_host:-Unset}
metrics_customer=${metrics_customer:-Unset}
metrics_instance=${metrics_instance:-Unset}
metrics_user=${metrics_user:-Unset}
metrics_passwd=${metrics_passwd:-Unset}
report_instance_logfile=${report_instance_logfile:-/p4/1/logs/report_instance_data.log}
metrics_cloudtype=${metrics_cloudtype:-Unset}
if [[ $metrics_host == Unset || $metrics_user == Unset || $metrics_passwd == Unset || $metrics_customer == Unset || $metrics_instance == Unset ]]; then
    msg "\\nError: Required parameters not supplied.\\n"
    msg "You must set the variables metrics_host, metrics_user, metrics_passwd, metrics_customer, metrics_instance in $ConfigFile."
    exit 1
fi
log autocloud is set to $autoCloud
## Auto set cloudtype in config?
if [[ $metrics_cloudtype == Unset ]]; then
    log "No Instance Type Defined"
    if [[ $autoCloud != 3 ]]; then
        log "using autoCloud"
        autoCloud=1
    fi
fi
cloudtype="${metrics_cloudtype^^}"

# Convert host from 9091 -> 9092 (pushgateway -> datapushgateway default)
# TODO - make more configurable
metrics_host=${metrics_host/9091/9092}

# Collect various metrics into a tempt report file we post off

pushd $(dirname "$metrics_logfile")

if [ $autoCloud -eq 1 ]; then
{
    log "Using autoCloud"
    #==========================
    # Check if running on AZURE
    log "Checking for AZURE"
    curl -s --connect-timeout $autoCloudTimeout -s -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance?api-version=2021-02-01" | grep -q "location"
    if [ $? -eq 0 ]; then
        curl -s --connect-timeout $autoCloudTimeout -s -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance?api-version=2021-02-01" | grep "location"  | awk -F\" '{print $4}' >/dev/null
        log "You are on an AZURE machine."
        declare -i IsAzure=1
        upcfg "Azure"
    else
        log "You are not on an AZURE machine."
        declare -i IsAzure=0
    fi
    #==========================
    # Check if running on AWS
    log "Checking for AWS"
    #aws_region_check=$(curl --connect-timeout $autoCloudTimeout -s http://169.254.169.254/latest/dynamic/instance-identity/document | grep -q "region")
    curl -s --connect-timeout $autoCloudTimeout -s http://169.254.169.254/latest/dynamic/instance-identity/document | grep -q "region"
    if [ $? -eq 0 ]; then
        curl -s --connect-timeout $autoCloudTimeout -s http://169.254.169.254/latest/dynamic/instance-identity/document | grep "region"  | awk -F\" '{print $4}' >/dev/null
        log "You are on an AWS machine."
        declare -i IsAWS=1
        upcfg "AWS"
    else
        log "You are not on an AWS machine."
        declare -i IsAWS=0
    fi
    #==========================
    # Check if running on GCP
    log "Checking for GCP"
    curl -s --connect-timeout $autoCloudTimeout -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true" -s | grep -q "google"
    if [ $? -eq 0 ]; then
        log "You are on a GCP machine."
        declare -i IsGCP=1
        upcfg "GCP"
    else
        log "You are not on a GCP machine."
        declare -i IsGCP=0
    fi
    }
    if [[ $IsAWS -eq 0 && $IsAzure -eq 0 && $IsGCP -eq 0 ]]; then
        log "No cloud detected setting to OnPrem"
        upcfg "OnPrem"
        declare -i IsOnPrem=1
    fi

    else {
        log "Not using autoCloud"
        declare -i IsAWS=0
        declare -i IsAzure=0
        declare -i IsGCP=0
        declare -i IsOnPrem=0
    }
fi


if [[ $cloudtype == AZURE ]]; then
    log "Config says cloud type is: Azure"
    declare -i IsAzure=1
fi
if [[ $cloudtype == AWS ]]; then
    log "Config says cloud type is: AWS"
    declare -i IsAWS=1
fi
if [[ $cloudtype == GCP ]]; then
    log "Config says cloud type is: GCP"
    declare -i IsGCP=1
fi
if [[ $cloudtype == ONPREM ]]; then
    log "Config says cloud type is: OnPrem"
    declare -i IsOnPrem=1
fi

if [[ $IsAWS -eq 1 ]]; then
    log "Doing the AWS meta-pull"
    run_command-runner --server --cloud=aws
fi

if [[ $IsAzure -eq 1 ]]; then
    log "Doing the Azure meta-pull"
    # DO Azure command-runner stuff
    run_command-runner --server --cloud=azure
fi

if [[ $IsGCP -eq 1 ]]; then
    log "Doing the GCP meta-pull"
    # DO GCP command-runner stuff
    run_command-runner --server --cloud=gcp
fi

if [[ $IsOnPrem -eq 1 ]]; then
    log "Doing the OnPrem stuff"
    run_command-runner --server

fi

findP4D
if [[ $p4dInstalled -eq 1 ]]; then
    get_sdp_instances
fi

# Loop while pushing as there seem to be temporary password failures quite frequently
# TODO Look into this.. (Note: Looking at the go build it's potentially related datapushgate's go build) --- Regarding password authentication.
iterations=0
max_iterations=10
STATUS=1

#push to datapushgateway
while [ $STATUS -ne 0 ]; do
    sleep 1
    ((iterations=$iterations+1))
    log "Pushing Support data"
    result=$(curl -s --connect-timeout $autoCloudTimeout --retry 5 --user "$metrics_user:$metrics_passwd" --data-binary "@$TempLog" "$metrics_host/json/?customer=$metrics_customer&instance=$metrics_instance")
    STATUS=0
    log "Checking result: $result"
    if [[ "$result" = '{"message":"invalid username or password"}' ]]; then
        STATUS=1
        log "Retrying due to temporary password failure"
    fi
    if [ "$iterations" -ge "$max_iterations" ]; then
        log "Push loop iterations exceeded"
        exit 1
    fi
done
popd
