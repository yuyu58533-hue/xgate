# 自用版指向说明

本包已去掉 VMess、加上 Hysteria2 / TUIC，并且不再默认指向
上游 vioIetMiku/miku-gate，改成指向自用仓库：

**https://github.com/yuyu58533-hue/xgate**

安装命令：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/yuyu58533-hue/xgate/main/mikugate.sh)
```

不想用默认仓库时，也可以运行时传环境变量，环境变量优先：

```bash
MIKU_GATE_REPO=你的用户名/你的仓库名 bash <(curl -fsSL https://raw.githubusercontent.com/yuyu58533-hue/xgate/main/mikugate.sh)
```

## 这次改了哪些指向

- `mikugate.sh`：下载 `install-online.sh` 的地址，改成
  `https://github.com/yuyu58533-hue/xgate/releases/latest/download/install-online.sh`
  （可用 `MIKU_GATE_REPO` 覆盖）。
- `install-online.sh`：GitHub API 地址改成
  `https://api.github.com/repos/yuyu58533-hue/xgate`，同样可用环境变量覆盖。
- `src/update.go`：更新检查仓库改成 `updateRepoValue()`，
  默认 `yuyu58533-hue/xgate`，优先读 `MIKU_GATE_REPO` 环境变量；
  原来这里写的是上游 fanout 仓库，也一并改掉了。
- `src/main.go`：更新检查返回的仓库链接改成跟上面同一处。
- `README.md`、`scripts/release-notes.md`：安装命令和 Release 链接
  换成 `yuyu58533-hue/xgate` 版。

注意：要让上面的安装命令真正能跑通，`yuyu58533-hue/xgate` 仓库里
需要有 `main` 分支的 `mikugate.sh`，并且 Release 里有
`install-online.sh` 这个附件，否则下载那一步会 404。

## 故意没改的

- VPN Gate 节点列表镜像（GeorgeXie2333 等）：那是节点数据来源，
  不是 miku-gate 上游，改了会断节点来源。
- sing-box 核心下载（CORE_ASSETS.json 里的 SagerNet）：
  那是协议核心本身，不是 miku-gate 上游。
- README 里的赞助商、原项目致谢段落：属于署名，不是安装指向，
  留着原样；你要去掉赞助商段落时自行删那一段即可。
- 文件名、目录名：全部保持原样，没有重命名。

## 配套改动

- Xray 已移除：自建核心只剩 sing-box，一键配置直接走 sing-box；
  XHTTP 入站不再支持，升级前请先删掉或改成 ws/grpc/tcp。
- VMess 已移除：原生入站不再接受 VMess，界面下拉里没有 VMess，
  `vmess://` 导入和订阅转换都会报错；升级后残留的 VMess 入站
  不再生成运行配置、不再输出分享链接，需要在面板里删掉。
- Hysteria2 / TUIC：见 `docs/RELAY_HY2_TUIC.md`，自建核心只有 sing-box，
  未绑出口默认走 `127.0.0.1:8500`，绑了出口走该隧道。
