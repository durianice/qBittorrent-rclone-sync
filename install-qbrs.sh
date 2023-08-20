#!/bin/bash

if [[ ! -f "/etc/systemd/system/qbrs.service" ]]; then
    if ! command -v wget &>/dev/null; then
        echo "请先安装 wget"
        exit 1
    fi
    if ! command -v vim &>/dev/null; then
        echo "请先安装 vim"
        exit 1
    fi
    cd ~
    wget "https://github.com/CCCOrz/qBittorrent-rclone-sync/releases/download/v1.0.0/qbrs"
    wget -O config.env https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/go/config.example
    vim config.env

    if [[ ! -f "config.env" ]]; then
        echo "配置文件config.env 不存在"
        exit 1
    fi

    mv qbrs /usr/local/bin/
    mv config.env /usr/local/bin/

    echo "[Unit]
    Description=qBittorrent-rclone-sync
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/qbrs
    WorkingDirectory=/usr/local/bin/
    Restart=always

    [Install]
    WantedBy=default.target" > /etc/systemd/system/qbrs.service

    systemctl daemon-reload
    systemctl start qbrs
    systemctl enable qbrs
fi

echo "启动 systemctl start qbrs"
echo ""
echo "停止 systemctl stop qbrs"
echo ""
echo "状态 systemctl status qbrs"
echo ""
echo "开机自启 systemctl enable qbrs"
echo ""
echo "禁用自启 systemctl disable qbrs"
echo ""
