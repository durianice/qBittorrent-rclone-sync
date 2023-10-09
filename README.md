## 功能
- 动态启动停止(硬盘使用xx时停止下载、占用小于xx时恢复下载)
- 可选保种选项
- 多线程上传到远程盘
- Telegram机器人通知

## 开始
客户端 qBittorrent v4.x.x


网盘挂载 rclone


远端存储 Google Drive / One Drive


本机 Ubuntu 20.04 / 2CPU / 1GB RAM / 硬盘40GB


支持的平台：[见发行版](https://github.com/CCCOrz/qBittorrent-rclone-sync/releases)


## 安装/更新
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/install-qbrs.sh)"
```

## 卸载
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/CCCOrz/qBittorrent-rclone-sync/release/uninstall-qbrs.sh)"
```

## 参考配置文件
[config.example](https://github.com/CCCOrz/qBittorrent-rclone-sync/blob/release/go/config.example)

### 标签
❗脚本控制：添加这个标签（或者添加自动创建的两个分类）会受💥**脚本控制**💥，按顺序下载，自动启动/停止
保种：添加这个标签不会删除本地资源，用于刷上传量（不想保留了移除该标签会自动删除本地资源）
### 分类
启动程序会自动创建 "_电影"、"_电视节目" 这两个分类，为资源添加了分类会自动按文件夹归类并且受💥**脚本控制**💥


![image](https://github.com/CCCOrz/qBittorrent-rclone-sync/assets/135111234/53a64c12-8610-4ffc-ad88-3c90c078ada0)

## 本地开发&手动编译
```
git clone -b release https://github.com/CCCOrz/qBittorrent-rclone-sync.git
sudo bash go-build.sh
```

## 注意事项
- 启用脚本控制后会自动勾选<按顺序下载>保证磁盘不被未完成资源占坑
- ❗目前版本添加tracker后需要手动添加并打上标签<脚本控制>
- Docker版下载保存路径需要注意与实际挂载目录一致

## Todo
- [ ] qBittorrent自动打标签
- [x] 按qBittorrent分类来分目录上传保存路径
- [ ] 更多的自定义配置
- [ ] ...

