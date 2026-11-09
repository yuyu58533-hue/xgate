## Miku Gate v0.3.1 — 修复 sing-box 配置校验失败

v0.3.0 给 direct 出站加的 `domain_strategy: ipv4_only` 是旧写法，sing-box 1.12 已废弃、1.14 已移除，导致 `sing-box check` 直接报错、配置无法应用。已按官方迁移方案改为 `domain_resolver` + local DNS 服务器，direct 出站照常强制 IPv4；已用真实 sing-box 1.14.3 校验通过。

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/yuyu58533-hue/xgate/main/mikugate.sh)
```

- 自建核心只剩 sing-box；XHTTP 入站不支持，升级前请先删掉或改成 ws/grpc/tcp。
- REALITY 密钥由程序内生成；旧版 panel_mode 里写的 `native` 会自动按 sing-box 处理。
- 3x-ui / S-UI 接管模式与 xray-cf-lite 模式不受影响。
- Docker 默认两个独立 Compose 项目：非特权 Nginx 静态前端，以及负责认证 / API / VPN 的 Go 后端。原生和仅后端继续兼容内嵌页面。

准备 Linux Docker/Compose、Python 3.10+ 和 TUN；部署后用 `sudo bash install-docker.sh frontend info` 读取真实前端完整地址，并用后端口令登录。

- 两端端口占用自动避让，默认优先后端8899 / 前端8900，读取真实保存结果，不猜地址。
- frontend start/restart/stop/info/check 独立维护，停前端不停止后端；顶层 stop/down 操作两端。
- 同源登录/API/订阅代理，随机路径与 Origin 保护；远程 HTTPS 后端验证证书。
- 前端专用包可部署到另一台 Linux 服务器，无需后端 TUN、Go、程序或私钥。
- 保留缓存优化和 quick/sub；二进制全部从源码重建为 0.3.1。

附件为 `miku`、完整 VPS 包、完整源码 ZIP、前端专用包、install-online.sh、SHA256SUMS，共六项。

详细安装、跨服务器、HTTPS、端口和升级说明见 README 与 docs/FRONTEND.md。旧原生维护工具有文件数限制，升级推荐解压新包后 install.sh --update；Docker 先备份再更新原项目目录，保留数据与 .env。旧标签和附件不覆盖。

后端仍需特权、host网络和TUN；前端权限不同。Docker 不运行原生 systemd 续期 timer，前端公网 TLS 由外层反代维护。真实隧道/公网证书由部署者验收。
