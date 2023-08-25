#!/bin/bash

# Script to retrieve static system information

echo "---------------------"
echo "System Information"
echo "---------------------"

# Operating System Information
echo "OS Information:"
echo $(lsb_release -d | cut -f2-)

# Kernel Version
echo "Kernel Version:"
echo $(uname -r)

# Architecture
echo "Architecture:"
echo $(uname -m)

# CPU Information
echo "CPU Information:"
cpu_info=$(lscpu | grep "Model name:")
echo ${cpu_info#"Model name:"}

# Memory Information
echo "Total Memory:"
total_memory=$(free -h | grep Mem: | awk '{print $2}')
echo $total_memory

# Disk Size
echo "Disk Size:"
disk_size=$(df -h --total | grep total | awk '{print $2}')
echo $disk_size

# Graphics Card Information (requires lspci tool)
if command -v lspci &> /dev/null; then
    echo "Graphics Card Information:"
    echo $(lspci | grep VGA | cut -d: -f3-)
fi

echo "---------------------"
echo "End of System Information"
echo "---------------------"
