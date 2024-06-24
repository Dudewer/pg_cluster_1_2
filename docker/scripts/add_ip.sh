#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <ip>"
    exit 1
fi

ip=$1
interface=$(ip route | awk '/default/ {print $5}')
def_ip=$(ip addr show dev $interface | awk '/inet / {print $2}')

if ! [[ $ip =~  $def_ip ]]; then
    echo "IP address not set on default interface. Setting IP address..."
    ip addr add $ip/24 dev $interface
    echo "IP address set to $ip interface $interface"
    exit 0
else
    echo "IP address $def_ip already set on interface $interface"
    exit 1
fi
