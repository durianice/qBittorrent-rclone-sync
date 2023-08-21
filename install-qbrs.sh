#!/bin/bash

get_platform() {
    ARCH=$(uname -m)
    result=""
        case "$ARCH" in
            x86_64)
                result="amd64"
                ;;
            aarch64)
                result="arm64"
                ;;
            armv7l)
                result="arm"
                ;;
            ppc64le)
                result="ppc64le"
                ;;
            ppc64)
                result="ppc64"
                ;;
            s390x)
                result="s390x"
                ;;
            *)
                result=""
                ;;
        esac
    if [[ $result == "" ]]; then
        echo "暂不支持该平台: $ARCH，请手动编译"
        exit 1
    fi
    echo "$result"
}
if [[ ! -f "/etc/systemd/system/qbrs.service" ]]; then
    if ! command -v wget &>/dev/null; then
        echo "请先安装 wget"
        exit 1
    fi
    if ! command -v vim &>/dev/null; then
        echo "请先安装 vim"
        exit 1
    fi
    type=$(get_platform)
    cd ~
    wget "https://github.com/CCCOrz/qBittorrent-rclone-sync/releases/download/v1.0.0/qbrs_${type}"
    wget -O config.env https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/go/config.example
    vim config.env

    if [[ ! -f "config.env" ]]; then
        echo "配置文件config.env 不存在"
        exit 1
    fi

    mv qbrs /usr/local/bin/
    mv config.env /usr/local/bin/
    chmod +x /usr/local/bin/qbrs

    echo "[Unit]
    Description=qBittorrent-rclone-sync
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/qbrs
    WorkingDirectory=/usr/local/bin/
    Restart=on-abnormal

    [Install]
    WantedBy=default.target" > /etc/systemd/system/qbrs.service

    systemctl daemon-reload
    systemctl start qbrs
    systemctl enable qbrs
    systemctl status qbrs
fi

echo "======== QBRS ========"
echo "启动 systemctl start qbrs"
echo "停止 systemctl stop qbrs"
echo "重启 systemctl restart qbrs"
echo "状态 systemctl status qbrs"
echo "配置文件 /usr/local/bin/config.env"
echo "开机自启 systemctl enable qbrs"
echo "禁用自启 systemctl disable qbrs"
echo "更多https://github.com/CCCOrz/qBittorrent-rclone-sync"
echo "======== QBRS ========"
