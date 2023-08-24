#!/bin/bash
#JACK KNIFE
#Inspired and big thanks to Geoff R.(Puppet) and Jack C.(Support) for the ztag.. Thank you
#
# monitor_tag Autobot scriptname
# ie monitor_tag: "Autobot jack_ztag.sh"

p4 -ztag servers | grep -E '^\.\.\. ExternalAddress' | awk '{print $3}' | while read -r address; do p4 -p "$address" -ztag info; done
