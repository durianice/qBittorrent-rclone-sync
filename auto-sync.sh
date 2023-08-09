#!/bin/bash

############# Global Config #############

# qBittorrent Config
qb_host="http://127.0.0.1:8080"
qb_username="admin"
qb_password="adminadmin"
# rclone Config
rclone_name="gdrive:"
rclone_local_dir="/opt/GoogleDrive"
rclone_remote_dir="/media/tv/"
multi_thread_streams=16
# Custom Config

# 标签
SYNC_TAG="SYNC"
CTRL_TAG="CTRL"
STAY_TAG="STAY"

# 缓冲
MAX_DIST=9
MIN_DIST=3
# 上传线程数
THREAD=5

# 临时目录
TEMP_PATH="/opt/qBittorrent/temp_dir"
if [ ! -d "$TEMP_PATH" ]; then
    mkdir -p ${TEMP_PATH}
fi
# 日志文件
LOG_FILE=${TEMP_PATH}/qbit_sync_rclone.log

############# Statr #############

function login() {
    qb_cookie=$(curl -s -i --header "Referer: ${qb_host}" --data "username=${qb_username}&password=${qb_password}" "${qb_host}/api/v2/auth/login" | grep -P -o 'SID=\S{32}')
    if [ -n ${qb_cookie} ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 登录成功 "
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 登录失败 "
        exit 1
    fi
}

# 获取所有信息
function get_all_info() {
    download_data=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -c --arg tag "$SYNC_TAG" '.[] | select(.tags | contains($tag))')
    download_data=$(echo ${download_data} | sed 's/}/},/g')
    download_data=${download_data%,}
    download_data="[${download_data}]"

}

# 获取下载信息
function get_download_info() {
    all_done_list="[]"
    local data=${download_data}
    local len=$(echo "$data" | jq -c '.|length')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 下载列表: ${len} "
    if [ "$len" -eq 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 下载列表为空 "
        return 0
    fi
    local COUNT=0
    local temp=''
    while [[ $COUNT -lt $len ]]; do
        item=$(echo "$data" | jq ".[$COUNT]")
        name=$(echo "$item" | jq -r '.name')
        hash=$(echo "$item" | jq -r '.hash')
        tags=$(echo "$item" | jq -r '.tags')
        seq_dl=$(echo "$item" | jq -r '.seq_dl')
        progress=$(echo "$item" | jq -r '.progress')
        download_path=$(echo "$item" | jq -r '.download_path')
        save_path=$(echo "$item" | jq -r '.save_path')

        {
            # 按顺序下载
            if [[ ${seq_dl} == false ]]; then
                echo "[$(date '+%Y-%m-%d %H:%M:%S')] 非按顺序下载 强制顺序下载 $name"
                pause ${hash}
                resume ${hash} ${seq_dl}
            fi
        } &

        finished_file=$(curl -s "${qb_host}/api/v2/torrents/files?hash=${hash}" --cookie "${qb_cookie}" | jq -c --arg dp "$download_path" --arg sp "$save_path" --arg tags "$tags" '.[] | select(.progress == 1) | . + {download_path: $dp,save_path: $sp,tags: $tags}')
        temp+="${finished_file}"
        let COUNT++
    done
    all_done_list=$(echo ${temp} | sed 's/}/},/g')
    all_done_list=${all_done_list%,}
    all_done_list="[${all_done_list}]"
}

# 暂停
function pause() {
    local hash=$(echo "$1" | sed 's/"//g')
    curl -s "${qb_host}/api/v2/torrents/pause" \
        -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' \
        -H "Cookie: ${qb_cookie}" \
        --data-raw "hashes=$hash" \
        --compressed \
        --insecure
    local res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r --arg hash "$hash" '.[] | select(.hash == $hash) | [.name, .state] | @csv')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已暂停 $res "
}

# 恢复下载
function resume() {
    local hash=$(echo "$1" | sed 's/"//g')
    local seq_dl=$2
    curl -s "${qb_host}/api/v2/torrents/resume" \
        -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' \
        -H "Cookie: ${qb_cookie}" \
        --data-raw "hashes=$hash" \
        --compressed \
        --insecure
    # 按顺序下载
    if [[ ${seq_dl} == false ]]; then
        curl "${qb_host}/api/v2/torrents/toggleSequentialDownload" \
            -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' \
            -H "Cookie: ${qb_cookie}" \
            --data-raw "hashes=$hash" \
            --compressed \
            --insecure
    fi
    local res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r --arg hash "$hash" '.[] | select(.hash == $hash) | [.name, .state] | @csv')
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已恢复 $res "
}

function get_downloading_queue() {
    local res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r '.[] | select(.state == "downloading"), select(.state == "forcedDL"), select(.state == "queuedDL") | [.name, .hash, .tags] | @csv')
    if [ -z "$res" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前下载&等待下载队列为空"
        return
    fi
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前下载&等待下载队列 $res"
    local -A map
    local -A tags_map
    while IFS=',' read name hash tags; do
        map[$name]="$hash"
        tags_map[$name]=$(echo "$tags" | sed 's/"//g')
    done <<<"$res"

    for i in "${!map[@]}"; do
        if [ "${tags_map[$i]}" != "*${CTRL_TAG}*" ]; then
            pause ${map[$i]}
        else
            continue
        fi
    done
}
function get_paused_queue() {
    local res=$(curl -s "${qb_host}/api/v2/torrents/info" --cookie "${qb_cookie}" | jq -r '.[] | select(.state == "pausedDL") | [.name, .hash, .tags, .seq_dl] | @csv')
    if [ -z "$res" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前暂停队列为空"
        return
    fi
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前暂停队列 $res"
    local -A map
    local -A tags_map
    local -A seq_dl_map
    while IFS=',' read name hash tags seq_dl; do
        map[$name]="$hash"
        tags_map[$name]=$(echo "$tags" | sed 's/"//g')
        seq_dl_map[$name]="$seq_dl"
    done <<<"$res"

    for i in "${!map[@]}"; do
        if [ "${tags_map[$i]}" != "*${CTRL_TAG}*" ]; then
            continue
        else
            resume ${map[$i]} ${seq_dl_map[$i]}
        fi
    done
}

function get_free_disk() {
    local free_space_kb=$(df --output=avail / | tail -n 1)
    local free_space_kb=$(expr $free_space_kb + 0)
    local free_space_mb=$(expr $free_space_kb / 1024)
    local free_space_gb=$(expr $free_space_mb / 1024)
    local free_space_gb=$(echo ${free_space_gb%.*})
    # MIN_DIST 到 MAX_DIST 之间作为缓冲
    if [ "$free_space_gb" -le "$MIN_DIST" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前磁盘空间 ${free_space_gb}GB 磁盘空间不足"
        get_downloading_queue
    elif [ "$free_space_gb" -gt "$MAX_DIST" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前磁盘空间 ${free_space_gb}GB 磁盘空间充足"
        get_paused_queue
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前磁盘空间 ${free_space_gb}GB 缓冲区"
    fi
}

function lock() {
    $(install -Dm0644 /dev/null "${TEMP_PATH}/lockfile/_$1.lock")
    # echo "[$(date '+%Y-%m-%d %H:%M:%S')] ${1} 已上锁 "
}
function get_lock_status() {
    if [[ -f "${TEMP_PATH}/lockfile/_$1.lock" ]]; then
        # echo "[$(date '+%Y-%m-%d %H:%M:%S')] ${1} 锁定中 "
        locked="1"
    else 
        locked="0"
    fi
}
function unlock() {
    rm -rf "${TEMP_PATH}/lockfile/_$1.lock"
    # echo "[$(date '+%Y-%m-%d %H:%M:%S')] ${1} 已解锁 "
}

function unlock_all() {
    rm -rf "${TEMP_PATH}/lockfile/"
}

function generateFifoFile() {
    tmp_fifofile="${TEMP_PATH}/$$.fifo" 
    mkfifo $tmp_fifofile
    exec 6<>$tmp_fifofile 
    rm $tmp_fifofile
    for ((i=0;i<${THREAD};i++));do
        echo -ne "\n" 1>&6
    done
}

# 同步
function sync() {
    generateFifoFile
    list=${all_done_list}
    local len=$(echo "$list" | jq -c '.|length')
    local COUNT=0
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 已下载完成的文件数: ${len} "
    if [[ ${len} == 0 ]]; then
        return 0
    fi
    local start_time=$(date +%s)
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始执行同步任务 并发数 ${THREAD}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行中..."
    while [[ $COUNT -lt $len ]]; do
        read -u 6
        {   
            item=$(echo "$list" | jq ".[$COUNT]")
            file_name=$(echo "$item" | jq -r '.name')
            tags=$(echo "$item" | jq -r '.tags')
            download_path=$(echo "$item" | jq -r '.download_path')
            save_path=$(echo "$item" | jq -r '.save_path')
            file_temp_path="${download_path}/${file_name}"
            file_save_path="${save_path}/${file_name}"
            source_path="${file_temp_path}"

            if [[ ! -f "${file_temp_path}" ]]; then
                source_path="${file_save_path}"
            fi

            if [[ -f "${source_path}" ]]; then
                get_lock_status "${file_name}"
                if [[ ${locked} == "0" ]]; then
                    lock "${file_name}"
                    target_path="${rclone_name}${rclone_remote_dir}${file_name}"
                    local_target_path="${rclone_local_dir}${rclone_remote_dir}${file_name}"
                    if [[ ! -f "${local_target_path}" ]]; then
                        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始同步 ${COUNT} "
                        echo "FROM:"
                        echo "${source_path}"
                        echo "TO:"
                        echo "${target_path}"
                        if [[ ${tags} == *${STAY_TAG}* ]]; then
                            echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前标签 ${tags} 保种模式"
                            cmd=$(/usr/bin/rclone -v -P copyto --multi-thread-streams ${multi_thread_streams} --log-file "${LOG_FILE}" "${source_path}" "${target_path}")
                        else
                            echo "[$(date '+%Y-%m-%d %H:%M:%S')] 当前标签 ${tags} 不保种"
                            cmd=$(/usr/bin/rclone -v -P copyto --multi-thread-streams ${multi_thread_streams} --log-file "${LOG_FILE}" "${source_path}" "${target_path}")
                        fi
                        cmd=$(/usr/bin/rclone -v -P copyto --multi-thread-streams ${multi_thread_streams} --log-file "${LOG_FILE}" "${source_path}" "${target_path}")
                        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步完成 ${COUNT} "
                    else 
                        # echo "!!!!!!!!!!!!!!!文件已存在!!!!!!!!!!!!!!"
                        # echo "${local_target_path}"
                        if [[ ${tags} != *${STAY_TAG}* ]]; then
                            rm "${source_path}"
                            # echo "[$(date '+%Y-%m-%d %H:%M:%S')] 移除已同步文件 ${COUNT} "
                        fi
                    fi
                    unlock "${source_path}"
                fi
            fi
            echo -ne "\n" 1>&6
        } &
        let COUNT++
    done
    wait
    exec 6>&-
    local end_time=$(date +%s)
    local duration=$(( end_time - start_time ))
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 同步任务执行结束 耗时 ${duration} 秒"
}

function main() {
    login
    get_free_disk
    get_lock_status "sync_task"
    if [[ ${locked} == "1" ]]; then
        exit 0
    fi
    lock "sync_task"
    get_all_info
    get_download_info
    sync
    unlock "sync_task"
    unlock_all
    exit 0
}

echo "-----------"
main
echo "-----------"
