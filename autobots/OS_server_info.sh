#!/bin/bash

echo "----------------------"
echo "Detailed System Report"
echo "----------------------"

# Hostname
echo "Hostname:"
echo $(hostname)

# System Manufacturer and Product Name (typically for laptops/desktops)
if [ -f /sys/class/dmi/id/sys_vendor ] && [ -f /sys/class/dmi/id/product_name ]; then
    echo "System Manufacturer:"
    cat /sys/class/dmi/id/sys_vendor

    echo "Product Name:"
    cat /sys/class/dmi/id/product_name
fi

# Number of CPU Cores
echo "Number of CPU Cores:"
echo $(nproc)

# CPU details
echo "CPU Details:"
cat /proc/cpuinfo | grep "model name" | uniq

# Hard Drive Information
echo "Hard Drive Details:"
echo $(fdisk -l | grep Disk | grep "/dev/sd")

# Network Information
echo "Network Interface Details:"
echo $(ip -o -4 addr show | awk '{print $2 " " $4}')

# Installed Software Versions (some common ones)
echo "Installed Software Versions:"

# Check for Python Version
if command -v python3 &> /dev/null; then
    echo "Python 3 Version:"
    python3 --version
else
    echo "Python 3 is not installed."
fi

# Check for GCC Version
if command -v gcc &> /dev/null; then
    echo "GCC Version:"
    gcc --version | grep gcc
else
    echo "GCC is not installed."
fi

# Check for Java Version
if command -v java &> /dev/null; then
    echo "Java Version:"
    java -version 2>&1 | head -n 1
else
    echo "Java is not installed."
fi

echo "----------------------"
echo "End of Detailed Report"
echo "----------------------"
