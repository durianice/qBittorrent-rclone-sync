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

install() {
    if ! command -v wget &>/dev/null; then
        echo "请先安装 wget"
        exit 1
    fi
    if ! command -v vim &> /dev/null && ! command -v nano &> /dev/null; then
        echo "请先安装 vim 或 nano"
        exit 1
    fi
    type=$(get_platform)
    filename="qbrs_${type}"
    cd ~
    REPO_URL="https://api.github.com/repos/CCCOrz/qBittorrent-rclone-sync/releases/latest"
    TAG=$(wget -qO- -t1 -T2 ${REPO_URL} | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')
    wget "https://github.com/CCCOrz/qBittorrent-rclone-sync/releases/download/$TAG/$filename"
    wget -O config.env https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/go/config.example
    vim config.env

    if [[ ! -f "config.env" ]]; then
        echo "配置文件config.env 不存在"
        exit 1
    fi

    if [[ -f "/usr/local/bin/config.env" ]]; then
        echo ">>>>>>>>>>>>>>>>"
        echo "旧的配置文件已备份，请重新编辑新配置文件并重启"
        echo "旧配置：/usr/local/bin/config.env.bak"
        echo "<<<<<<<<<<<<<<<<"
        cp /usr/local/bin/config.env /usr/local/bin/config.env.bak
    fi

    mv $filename /usr/local/bin/
    mv config.env /usr/local/bin/
    chmod +x "/usr/local/bin/$filename"

    echo "[Unit]
    Description=qBittorrent-rclone-sync
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/$filename
    WorkingDirectory=/usr/local/bin/
    Restart=on-abnormal

    [Install]
    WantedBy=default.target" > /etc/systemd/system/qbrs.service

    systemctl daemon-reload
    systemctl start qbrs
    systemctl enable qbrs
    systemctl status qbrs

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
}

uninstall() {
    sudo bash -c "$(curl -sL https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/uninstall-qbrs.sh)"
}

if [[ -f "/etc/systemd/system/qbrs.service" ]]; then
    uninstall
fi
install

