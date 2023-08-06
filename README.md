## Install
```
apt install jq -y
```
## Download
```
cd ~ && wget -q -O auto-sync.sh https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/main/auto-sync.sh && chmod a+x auto-sync.sh
```
## Cron
```
*/2 * * * * /bin/bash /root/qBittorrent/auto-sync.sh >> /root/qBittorrent/auto-sync.sh 2>&1
```

