#### 如在使用中有任何问题，请先使用一键更新命令升级到最新版本~

## 功能
- 动态启动停止(硬盘使用xx时停止下载、占用小于xx时恢复下载)
- 可选保种选项
- 多线程上传到远程盘
- Telegram机器人通知
- 检查新版本

## 开始
客户端 qBittorrent v4.x.x


网盘挂载 rclone


远端存储 Google Drive / One Drive


本机 Ubuntu 20.04 / 2CPU / 1GB RAM / 硬盘40GB


支持的平台：[见发行版](https://github.com/durianice/qBittorrent-rclone-sync/releases)


## 安装/更新
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/durianice/qBittorrent-rclone-sync/release/install-qbrs.sh)"
```

## 卸载
```
sudo bash -c "$(curl -sL https://raw.githubusercontent.com/durianice/qBittorrent-rclone-sync/release/uninstall-qbrs.sh)"
```

## 配置文件
[config.example](https://github.com/durianice/qBittorrent-rclone-sync/blob/release/app/config.example)
 
### 分类
- 启动程序会自动创建 "_电影"、"_电视节目" 这两个分类
- 传入自定义分类：如“合集”，对应的保存目录则为 `env.RCLONE_LOCAL_DIR/env.RCLONE_REMOTE_DIR/合集/`


**注意：请在新增下载时选择分类之一，否则不会自动同步~**

### 标签
想保留本地资源用于做种，给下载任务添加**保种**标签

## 本地开发&手动编译
```
git clone -b release https://github.com/durianice/qBittorrent-rclone-sync.git
sudo bash go-build.sh
```


## Todo
- [ ] qBittorrent自动打标签
- [x] 按qBittorrent分类来分目录上传保存路径
- [ ] 更多的自定义配置
- [ ] ...

