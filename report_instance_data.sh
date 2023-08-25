#!/bin/bash

cmdrnrscriptDir=$(dirname "$0")
commandRunnerPath="$cmdrnrscriptDir/command-runner"
echo "Command Runner Path: $commandRunnerPath"
TempLog="/tmp/out.json"
rm -f $TempLog

declare ConfigFile="/p4/common/config/.push_metrics.cfg"
declare report_instance_logfile="/opt/perforce/command-runner/logs/command-runner.log"

function msg () { echo -e "$*"; }
function log () { dt=$(date '+%Y-%m-%d %H:%M:%S'); echo -e "$dt: $*" >> "$report_instance_logfile"; msg "$dt: $*"; }
function bail () { msg "\nError: ${1:-Unknown Error}\n"; exit ${2:-1}; }

# Default value
declare useAutobots=true   # Change this to false if you don't want to use the --autobots flag

run_command_runner() {
    local command="$commandRunnerPath --server"
    if $useAutobots; then
        command+=" --autobots"
    fi
    
    log "Executing command: $command"
    $command 2>> "$report_instance_logfile"  # Redirecting stderr to the logfile
}

run_command_runner
