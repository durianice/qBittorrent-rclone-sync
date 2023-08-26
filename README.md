## 功能
- 动态启动停止(硬盘使用90%时停止下载、占用小于45%时恢复下载)
- 可选保种选项
- 多线程上传到远程盘
- Telegram机器人通知

## 开始
客户端 qBittorrent v4.x.x


网盘挂载 rclone


远端存储 Google Drive


本机 Ubuntu 20.04 / 2CPU / 1GB RAM / 硬盘40GB


## 安装
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/install-qbrs.sh)"
```

## 卸载
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/uninstall-qbrs.sh)"
```

## 参考配置文件
[config.example](https://github.com/CCCOrz/qBittorrent-rclone-sync/blob/release/go/config.example)

脚本控制：添加这个标签才会受脚本控制，按顺序下载，自动启动/停止

保种：添加这个标签不会删除本地资源，用于刷上传量（不想保留了移除该标签会自动删除本地资源）


![image](https://github.com/CCCOrz/qBittorrent-rclone-sync/assets/135111234/53a64c12-8610-4ffc-ad88-3c90c078ada0)

## 本地开发&手动编译
```
git clone -b release https://github.com/CCCOrz/qBittorrent-rclone-sync.git
sudo bash go-build.sh
```

## 注意事项
- 启用脚本控制后会自动勾选<按顺序下载>保证磁盘不被未完成资源占坑
- 目前版本添加tracker后需要手动添加并打上标签<脚本控制>

## Todo
- [ ] qBittorrent自动打标签
- [ ] GoogleDrive达到日流量时停止上传
- [ ] 按qBittorrent分类来分目录上传保存路径
- [ ] 更多的自定义配置
- [ ] ...

