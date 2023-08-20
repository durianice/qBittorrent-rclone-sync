#!/bin/bash

if [[ -f "/etc/systemd/system/qbrs.service" ]]; then
    systemctl stop qbrs
    systemctl disable qbrs
    
    rm /usr/local/bin/qbrs
    rm /usr/local/bin/config.env

    rm /etc/systemd/system/qbrs.service
    systemctl daemon-reload
fi

if [[ ! -f "/etc/systemd/system/qbrs.service" ]]; then
    echo "已卸载qbrs"
    echo "https://github.com/CCCOrz/qBittorrent-rclone-sync/tree/release"
fi
