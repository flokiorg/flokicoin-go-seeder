#!/bin/sh

LOGDIR="/var/log/dnsseed"

# Check if the first argument is provided
if [ -z "$1" ]; then
    echo "Error: Missing required argument for netfile."
    echo "Usage: $0 <netfile>"
    exit 1
fi

# Ensure the log directory exists
mkdir -p "${LOGDIR}"

# Compress old log files
gzip -q "${LOGDIR}"/*.log 2>/dev/null

echo
echo "======= Run the Flokicoin seeder ======="
echo

dnsseed -p 53 -v -w 8880 -netfile "$1" 2>&1 | tee "${LOGDIR}/$(date +%F-%s)-dnsseed.log"