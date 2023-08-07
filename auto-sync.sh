#!/bin/bash

############# Global Config #############

# qBittorrent Config
qb_host="http://127.0.0.1:8080"
qb_username="admin"
qb_password="adminadmin"
qb_transfers=4
# rclone Config
rclone_name="gdrive:"
rclone_local_dir="/opt/GoogleDrive"
rclone_remote_dir="/media/tv/"
# Custom Config
CTRL_TAG="CTRL_BY_SCRIPT"
MAX_DIST=10
MIN_DIST=1
# log
log_path="/opt/qBittorrent/log"
if [ ! -d "$log_path" ]; then
    mkdir -p ${log_path}
fi
log_file=${log_path}/qbit_sync_rclone.log

############# Statr #############
all=''
if [ ! -d "$folder" ]; then
    mkdir -p ${log_path}
fi

function login() {
    qb_cookie=$(curl -s -i --header "Referer: ${qb_host}" --data "username=${qb_username}&password=${qb_password}" "${qb_host}/api/v2/auth/login" | grep -P -o 'SID=\S{32}')
    if [ -n ${cookie} ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 登录成功 "
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 登录失败 "
        exit 1
    fi
}

# 获取所有信息
function get_all_info() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始获取下载列表 "
    download_data=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}")
    download_len=$(echo "$download_data" | jq '.|length')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 下载列表: ${download_len} "
    if [ "$download_len" -eq 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 暂无任务脚本退出 "
        exit 0
    fi
}

# 获取下载信息
function get_download_info() {
    get_all_info
    COUNT=0
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始获取文件详情 "
    while [[ $COUNT -lt $download_len ]]; do
        item=$(echo "$download_data" | jq ".[$COUNT]")
        name=$(echo "$item" | jq -r '.name')
        hash=$(echo "$item" | jq -r '.hash')
        progress=$(echo "$item" | jq -r '.progress')
        download_path=$(echo "$item" | jq -r '.download_path')
        save_path=$(echo "$item" | jq -r '.save_path')
        finished_file=$(curl -s "${qb_host}/api/v2/torrents/files?hash=${hash}" --cookie "${qb_cookie}" | jq -c --arg dp "$download_path" --arg sp "$save_path" '.[] | select(.progress == 1) | . + {download_path: $dp,save_path: $sp}')
        all+="${finished_file}"
        let COUNT++
    done
}

# 暂停
function pause() {
    hash=$(echo "$1" | sed 's/"//g')
    curl -s "${qb_host}/api/v2/torrents/pause" \
        -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' \
        -H "Cookie: ${qb_cookie}" \
        --data-raw "hashes=$hash" \
        --compressed \
        --insecure
    res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r --arg hash "$hash" '.[] | select(.hash == $hash) | [.name, .state] | @csv')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已暂停 $res "
}

# 恢复下载
function resume() {
    hash=$(echo "$1" | sed 's/"//g')
    curl -s "${qb_host}/api/v2/torrents/resume" \
        -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' \
        -H "Cookie: ${qb_cookie}" \
        --data-raw "hashes=$hash" \
        --compressed \
        --insecure
    res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r --arg hash "$hash" '.[] | select(.hash == $hash) | [.name, .state] | @csv')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已恢复下载 $res "
}

function lock() {
    $(touch "${log_path}/auto_sync.lock")
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已上锁 "
}
function get_lock_status() {
    if [[ -f "${log_path}/auto_sync.lock" ]]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 锁定中 "
        exit 0
    fi
}
function unlock() {
    rm -rf "${log_path}/auto_sync.lock"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已解锁 "
}

function getDownloading() {
    res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r '.[] | select(.state == "downloading"), select(.state == "forcedDL") | [.name, .hash, tags] | @csv')
    if [ -z "$res" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 下载队列为空"
        exit 0
    fi
    local -A map
    local -A tags_map
    while IFS=',' read name hash; do
        map[$name]="$hash"
        tags_map[$name]=$(echo "$tags" | sed 's/"//g')
    done <<<"$res"

    for i in "${!map[@]}"; do
        if [ "${tags_map[$i]}" != "${CTRL_TAG}" ]; then
            continue
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] $i ${map[$i]} 准备挂起下载"
            pause ${map[$i]}
        fi
    done
}
function getPause() {
    res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r '.[] | select(.state == "pausedDL") | [.name, .hash, .tags] | @csv')
    if [ -z "$res" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 暂停队列为空"
        exit 0
    fi
    local -A map
    local -A tags_map
    while IFS=',' read name hash tags; do
        map[$name]="$hash"
        tags_map[$name]=$(echo "$tags" | sed 's/"//g')
    done <<<"$res"

    for i in "${!map[@]}"; do
        if [ "${tags_map[$i]}" != "${CTRL_TAG}" ]; then
            continue
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] $i ${map[$i]} 准备恢复下载"
            resume ${map[$i]}
        fi
    done
}

function getFreeDisk() {
    free_space_kb=$(df --output=avail / | tail -n 1)
    free_space_kb=$(expr $free_space_kb + 0)
    free_space_mb=$(expr $free_space_kb / 1024)
    free_space_gb=$(expr $free_space_mb / 1024)
    free_space_gb=$(echo ${free_space_gb%.*})
    if [ "$free_space_gb" -lt $MIN_DIST ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 磁盘空间不足"
        getDownloading
        exit 0
    fi
    if [ "$free_space_gb" -gt $MAX_DIST ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 磁盘空间充足"
        getPause
    fi
}

# 同步已下载文件
function main() {
    login
    getFreeDisk
    get_lock_status
    get_download_info
    data=$(echo ${all} | sed 's/}/},/g')
    # data=$(echo ${all} | sed 's/}/},/g' | sed 's/ //g')
    data=${data%,}
    data="[${data}]"
    len=$(echo "$data" | jq -c '.|length')
    COUNT=0
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已下载完成的文件数: ${len} "
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始执行同步脚本 "
    lock
    while [[ $COUNT -lt $len ]]; do
        item=$(echo "$data" | jq ".[$COUNT]")
        file_name=$(echo "$item" | jq -r '.name')
        download_path=$(echo "$item" | jq -r '.download_path')
        save_path=$(echo "$item" | jq -r '.save_path')
        file_temp_path="${download_path}/${file_name}"
        file_save_path="${save_path}/${file_name}"
        source_path="${file_temp_path}"
        if [[ ! -f "${file_temp_path}" ]]; then
            source_path="${file_save_path}"
        fi

        if [[ ! -f "${source_path}" ]]; then
            # echo "!!!!!!!!!!!!!!!文件不存在!!!!!!!!!!!!!!"
            # echo "${source_path}"
            let COUNT++
            continue
        fi
        target_path="${rclone_name}${rclone_remote_dir}${file_name}"
        local_target_path="${rclone_local_dir}${rclone_remote_dir}${file_name}"
        if [[ -f "${local_target_path}" ]]; then
            # echo "!!!!!!!!!!!!!!!文件已存在!!!!!!!!!!!!!!"
            # echo "${local_target_path}"
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] 移除已同步文件 ${COUNT} "
            rm "${source_path}"
            let COUNT++
            continue
        fi
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始同步 ${COUNT} "
        echo "FROM:"
        echo "${source_path}"
        echo "TO:"
        echo "${target_path}"
        cmd=$(/usr/bin/rclone -v -P moveto --transfers ${qb_transfers} --log-file "${log_file}" "${source_path}" "${target_path}")
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步完成 ${COUNT} "
        let COUNT++
    done
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步脚本执行结束 "
    unlock
    exit 0
}

main
