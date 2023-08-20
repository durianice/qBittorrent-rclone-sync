## 开始
客户端 qBittorrent v4.5.2


网盘挂载 rclone


远端存储 Google Drive


本机 Ubuntu 20.04 / 2CPU / 1GB RAM / 硬盘40GB


## 安装运行
```
wget -N --no-check-certificate https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/main/install-qbrs.sh && bash install-qbrs.sh
```

## 卸载
```
wget -N --no-check-certificate https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/main/uninstall-qbrs.sh && bash uninstall-qbrs.sh
```

## 参考配置文件
[config.example](https://github.com/CCCOrz/qBittorrent-rclone-sync/blob/release/go/config.example)

## 功能
- 动态启动停止(硬盘使用90%时停止下载、占用小于45%时恢复下载)
- 可选保种选项
- 多线程上传到远程盘
- Telegram机器人通知

## 注意事项
目前版本添加tracker后需要手动添加并打上标签<脚本控制>

## Todo
- [ ] qBittorrent自动打标签
- [ ] GoogleDrive达到日流量时停止上传
- [ ] 按qBittorrent分类来分目录上传保存路径
- [ ] 更多的自定义配置
- [ ] ...

