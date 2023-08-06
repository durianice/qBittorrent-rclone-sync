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
    qb_cookie=$(curl -i --header "Referer: ${qb_host}" --data "username=${qb_username}&password=${qb_password}" "${qb_host}/api/v2/auth/login" | grep -P -o 'SID=\S{32}')
    if [ -n ${cookie} ]
    then
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
        finished_file=$(curl -s "${qb_host}/api/v2/torrents/files?hash=${hash}" --cookie "${qb_cookie}" | jq -c --arg dp "$download_path" --arg sp "$save_path"  '.[] | select(.progress == 1) | . + {download_path: $dp,save_path: $sp}')
        all+="${finished_file}"
        let COUNT++
    done
}

function lock(){
    $(touch auto_sync.lock)
}
function get_lock_status(){
    if [[ -f "auto_sync.lock" ]]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已有同步程序正在执行 "
        exit 0
    fi
}
function unlock(){
    $(rm -rf auto_sync.lock)
}

# 同步已下载文件
function main() {
    get_lock_status
    login
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
            # rm "${source_path}"
            let COUNT++
            continue
        fi
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始同步 ${COUNT} "
        echo "FROM:"
        echo "${source_path}"
        echo "TO:"
        echo "${target_path}"
        cmd=$(/usr/bin/rclone  -v -P moveto --transfers ${qb_transfers} --log-file "${log_file}" "${source_path}" "${target_path}")
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步完成 ${COUNT} "
        let COUNT++
    done
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步脚本执行结束 "
    unlock
    exit 0
}

main
