#!/bin/bash

if [[ -f "/etc/systemd/system/qbrs.service" ]]; then
    systemctl stop qbrs
    systemctl disable qbrs
    
    rm /usr/local/bin/qbrs
    rm /usr/local/bin/config.env

    rm /etc/systemd/system/qbrs.service
    systemctl daemon-reload
fi
