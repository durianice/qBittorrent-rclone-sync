#!/bin/bash

if [[ ! -f "/etc/systemd/system/qbrs.service" ]]; then

    cd ./go
    if [[ ! -f "config.env" ]]; then
        echo "请先运行以下命令"
        echo "cd go && cp config.example config.env && vim config.env"
        exit 1
    fi

    go get
    go build -o qbrs -v main.go

    mv qbrs /usr/local/bin/
    cp config.env /usr/local/bin/
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
