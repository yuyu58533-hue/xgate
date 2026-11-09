#!/usr/bin/env bash
# Miku Gate short installer entry point (self-use version).
#
# 自用版指向：不要再用上游 vioIetMiku/miku-gate，默认指向
# yuyu58533-hue/xgate。运行时可用环境变量覆盖，环境变量优先：
#   MIKU_GATE_REPO=你的用户名/你的仓库名 bash <(curl -fsSL https://raw.githubusercontent.com/yuyu58533-hue/xgate/main/mikugate.sh)
set -euo pipefail
repo="${MIKU_GATE_REPO:-yuyu58533-hue/xgate}"
url="https://github.com/${repo}/releases/latest/download/install-online.sh"
tmp=$(mktemp -d "${TMPDIR:-/tmp}/mikugate.XXXXXXXX")
trap 'rm -rf -- "$tmp"' EXIT
curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$tmp/install-online.sh"
bash "$tmp/install-online.sh" "$@"
