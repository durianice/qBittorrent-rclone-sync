## 开始
客户端 qBittorrent v4.5.2


网盘挂载 rclone


远端存储 Google Drive


## 安装依赖软件
```
apt install jq -y
```

## 下载脚本
```
cd ~ && wget -q -O auto-sync.sh https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/main/auto-sync.sh && chmod a+x auto-sync.sh
```
## 配置文件
待补充

## 定时执行
```
*/2 * * * * /bin/bash /root/qBittorrent/auto-sync.sh >> /root/qBittorrent/auto-sync.sh 2>&1
```

## 特别说明
~~如果没有控制好下载速度，Download > Upload，可能会导致VPS内存爆满等其他问题😣~~

已启用拥塞控制😈

## Todo
- [ ] 消息通知
- [ ] 内存预警
- [ ] 自定义保种时间
- [ ] 多线程上传
- [x] 空间不足时暂停下载
- [ ] 文件接力
- [ ] ...

