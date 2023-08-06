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

## Friendly reminder
如果没有控制好下载速度，Download > Upload，可能会导致VPS内存爆满等其他问题😣

## Todo
- [ ] 消息通知
- [ ] 内存预警
- [ ] 自定义保种时间
- [ ] 多线程上传
- [ ] 空间不足时暂停等待
- [ ] ...

