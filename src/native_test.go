package main

import (
	"strings"
	"testing"
)

func TestNativeInboundTagMatchesXUIFormat(t *testing.T) {
	// tag 格式必须和 3x-ui 一致，否则两种后端的绑定语义会对不上
	cases := []struct {
		ib   nativeInbound
		want string
	}{
		{nativeInbound{Port: 443, Network: "tcp"}, "in-443-tcp"},
		{nativeInbound{Port: 8080, Network: "ws"}, "in-8080-ws"},
		{nativeInbound{Port: 1234}, "in-1234-tcp"}, // 缺省按 tcp
	}
	for _, c := range cases {
		if got := c.ib.tag(); got != c.want {
			t.Errorf("tag() = %q, want %q", got, c.want)
		}
	}
}

func TestBuildSingBoxConfigBindsOnlyLiveTunnels(t *testing.T) {
	up := &Tunnel{Port: 1080, Status: "up", Node: Node{HostName: "jp1"}}
	down := &Tunnel{Port: 1081, Status: "failed", Node: Node{HostName: "jp2"}}
	inbounds := []*nativeInbound{
		{ID: 1, Port: 100, Protocol: "vless", Enable: true, BoundTo: "jp1"},
		{ID: 2, Port: 200, Protocol: "vless", Enable: true, BoundTo: "jp2"},
		{ID: 3, Port: 300, Protocol: "vless", Enable: true},
	}

	cfg, err := buildSingBoxConfig(inbounds, []*Tunnel{up, down})
	if err != nil {
		t.Fatal(err)
	}

	outs := map[string]bool{}
	for _, o := range cfg["outbounds"].([]any) {
		outs[o.(map[string]any)["tag"].(string)] = true
	}
	if !outs["miku-jp1"] {
		t.Error("已连通的隧道应当有对应出站")
	}
	if outs["miku-jp2"] {
		t.Error("未连通的隧道不该生成出站")
	}

	rules := cfg["route"].(map[string]any)["rules"].([]any)
	if len(rules) != 2 {
		t.Fatalf("连通出口路由与失联出口阻断规则均须保留，实际 %d 条", len(rules))
	}
	if got := rules[0].(map[string]any)["outbound"]; got != "miku-jp1" {
		t.Errorf("outbound = %v, want miku-jp1", got)
	}
	if got := rules[1].(map[string]any)["action"]; got != "reject" {
		t.Errorf("失联出口不能改走直连: %v", got)
	}
}

func TestBuildSingBoxConfigForcesIPv4OnDirect(t *testing.T) {
	// 隧道内没有 IPv6，direct 走 IPv6 会暴露母机真实地址
	cfg, err := buildSingBoxConfig(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// domain_resolver 需要一个 local DNS 服务器
	dns, ok := cfg["dns"].(map[string]any)
	if !ok {
		t.Fatal("缺少 dns 配置")
	}
	servers, ok := dns["servers"].([]any)
	if !ok || len(servers) != 1 {
		t.Fatalf("应有一个 local DNS 服务器，实际 %v", dns["servers"])
	}
	server := servers[0].(map[string]any)
	if server["type"] != "local" || server["tag"] != "local" {
		t.Errorf("local DNS 服务器不对: %v", server)
	}
	for _, o := range cfg["outbounds"].([]any) {
		m := o.(map[string]any)
		if m["tag"] != "direct" {
			continue
		}
		resolver, ok := m["domain_resolver"].(map[string]any)
		if !ok || resolver["server"] != "local" || resolver["strategy"] != "ipv4_only" {
			t.Errorf("direct 出站应经 local DNS 强制 IPv4，实际 %v", m["domain_resolver"])
		}
		return
	}
	t.Fatal("没有找到 direct 出站")
}

func TestShareLinkPerProtocol(t *testing.T) {
	c := nativeClient{ID: "uuid-1", Password: "pw-1", Email: "e", Enable: true}

	vless := shareLink(&nativeInbound{Port: 100, Protocol: "vless", Remark: "r"}, c, "1.2.3.4")
	if !strings.HasPrefix(vless, "vless://uuid-1@1.2.3.4:100?") {
		t.Errorf("vless 链接格式不对: %s", vless)
	}
	if !strings.Contains(vless, "encryption=none") {
		t.Errorf("vless 需要 encryption=none: %s", vless)
	}

	tro := shareLink(&nativeInbound{Port: 200, Protocol: "trojan", Network: "ws", Path: "/p"}, c, "h")
	if !strings.HasPrefix(tro, "trojan://pw-1@h:200?") {
		t.Errorf("trojan 应当用密码而不是 UUID: %s", tro)
	}
	if !strings.Contains(tro, "path=%2Fp") {
		t.Errorf("ws 链接要带 path: %s", tro)
	}
}

// 复制出来的入站直接叫出口的名字，不再拼模板备注。
// 以前拼接会在"拿复制品当模板再复制"时叠成 miku-JP-243-VN-165 这种。
func TestUniqueRemark(t *testing.T) {
	taken := map[string]bool{}
	if got := uniqueRemark("🇯🇵 日本 54", taken); got != "🇯🇵 日本 54" {
		t.Errorf("没撞名时应原样返回，实际 %q", got)
	}
	taken["🇯🇵 日本 54"] = true
	// 同一条出口挂第二个节点时要加序号，mihomo 那边名字重了会丢节点
	if got := uniqueRemark("🇯🇵 日本 54", taken); got != "🇯🇵 日本 54 2" {
		t.Errorf("撞名应加序号，实际 %q", got)
	}
	taken["🇯🇵 日本 54 2"] = true
	if got := uniqueRemark("🇯🇵 日本 54", taken); got != "🇯🇵 日本 54 3" {
		t.Errorf("第三个应是 3，实际 %q", got)
	}
}

// 换节点只改 fanout 自己起的名字，用户手工改过的不碰。
func TestRenameExitLabel(t *testing.T) {
	if got := renameExitLabel("🇯🇵 日本 54", "🇰🇷 韩国 15"); got != "🇰🇷 韩国 15" {
		t.Errorf("自动名应整体换掉，实际 %q", got)
	}
	for _, mine := range []string{"线路A", "给老王的", "JP-244", ""} {
		if got := renameExitLabel(mine, "🇰🇷 韩国 15"); got != mine {
			t.Errorf("手工备注 %q 不该被改成 %q", mine, got)
		}
	}
	// 新名字算不出来时别把原来的抹掉
	if got := renameExitLabel("🇯🇵 日本 54", ""); got != "🇯🇵 日本 54" {
		t.Errorf("新标签为空时应保持原样，实际 %q", got)
	}
}

func TestIsGeneratedLabel(t *testing.T) {
	for _, s := range []string{"🇯🇵 日本 54", "🇺🇸 美国 1", " 🇰🇷 韩国 15"} {
		if !isGeneratedLabel(s) {
			t.Errorf("%q 是自动生成的名字", s)
		}
	}
	for _, s := range []string{"", "线路A", "JP-244", "miku-JP-243", "日本 54"} {
		if isGeneratedLabel(s) {
			t.Errorf("%q 不是自动生成的名字", s)
		}
	}
}

func TestVisionCapable(t *testing.T) {
	// Vision 只在 VLESS + 裸 TCP + TLS/REALITY 下有效，其他组合 Xray 会拒绝启动
	if !visionCapable("vless", "tcp", "reality") {
		t.Error("vless/tcp/reality 应当支持 vision")
	}
	if !visionCapable("vless", "tcp", "tls") {
		t.Error("vless/tcp/tls 应当支持 vision")
	}
	if visionCapable("vless", "ws", "tls") {
		t.Error("ws 不该支持 vision")
	}
	if visionCapable("vless", "tcp", "none") {
		t.Error("没有安全层时不该支持 vision")
	}
	if visionCapable("trojan", "tcp", "tls") {
		t.Error("vision 是 VLESS 专属")
	}
}

func TestStreamSettingsPerNetwork(t *testing.T) {
	cases := []struct {
		ib      nativeInbound
		key     string
		wantKey string
		want    any
	}{
		{nativeInbound{Network: "ws", Path: "/p"}, "wsSettings", "path", "/p"},
		{nativeInbound{Network: "httpupgrade", Path: "/h"}, "httpupgradeSettings", "path", "/h"},
		{nativeInbound{Network: "xhttp", Path: "/x"}, "xhttpSettings", "path", "/x"},
		// gRPC 没有 path，Path 字段复用为 serviceName，且不带前导斜杠
		{nativeInbound{Network: "grpc", Path: "/svc"}, "grpcSettings", "serviceName", "svc"},
	}
	for _, c := range cases {
		st := streamSettingsJSON(&c.ib)
		sub, ok := st[c.key].(map[string]any)
		if !ok {
			t.Errorf("%s 缺少 %s", c.ib.Network, c.key)
			continue
		}
		if got := sub[c.wantKey]; got != c.want {
			t.Errorf("%s 的 %s = %v, want %v", c.ib.Network, c.wantKey, got, c.want)
		}
	}
}

func TestStreamSettingsReality(t *testing.T) {
	ib := nativeInbound{
		Network: "tcp", Security: "reality",
		Reality: &realityConfig{
			Dest: "www.cloudflare.com:443", ServerNames: []string{"www.cloudflare.com"},
			PrivateKey: "priv", PublicKey: "pub", ShortIDs: []string{"abcd1234"},
		},
	}
	st := streamSettingsJSON(&ib)
	if st["security"] != "reality" {
		t.Fatalf("security = %v", st["security"])
	}
	r, ok := st["realitySettings"].(map[string]any)
	if !ok {
		t.Fatal("缺少 realitySettings")
	}
	if r["privateKey"] != "priv" {
		t.Errorf("服务端要写私钥，实际 %v", r["privateKey"])
	}
	// 公钥只有客户端用，写进服务端配置会被 Xray 拒绝
	if _, leaked := r["publicKey"]; leaked {
		t.Error("服务端配置不该出现 publicKey")
	}
}

func TestShareLinkCarriesSecurityParams(t *testing.T) {
	c := nativeClient{ID: "uuid-1", Enable: true, Flow: "xtls-rprx-vision"}

	re := shareLink(&nativeInbound{
		Port: 100, Protocol: "vless", Network: "tcp", Security: "reality", Remark: "r",
		Reality: &realityConfig{
			ServerNames: []string{"www.cloudflare.com"}, PublicKey: "PBK",
			ShortIDs: []string{"sid1"}, Fingerprint: "chrome",
		},
	}, c, "h")
	for _, want := range []string{"pbk=PBK", "sid=sid1", "fp=chrome",
		"sni=www.cloudflare.com", "flow=xtls-rprx-vision"} {
		if !strings.Contains(re, want) {
			t.Errorf("REALITY 链接缺少 %s: %s", want, re)
		}
	}

	// 自签证书验不过 CA，链接必须带指纹，否则客户端连不上
	tl := shareLink(&nativeInbound{
		Port: 200, Protocol: "vless", Network: "tcp", Security: "tls", Remark: "t",
		TLS: &tlsConfig{ServerName: "demo.local", SelfSigned: true, CertSha256: "AABB"},
	}, nativeClient{ID: "u", Enable: true}, "h")
	if !strings.Contains(tl, "pinSHA256=AABB") {
		t.Errorf("自签 TLS 链接要带证书指纹: %s", tl)
	}
}

// Hy2 链接要和 v2ray-agent 等生成的一致：显式 alpn=h3，
// 自签时 sni + insecure=1 + 证书指纹；正式证书只带 sni + alpn。
func TestShareLinkHysteria2(t *testing.T) {
	c := nativeClient{ID: "uuid-1", Password: "pw-1", Email: "e", Enable: true}

	hy := shareLink(&nativeInbound{
		Port: 24560, Protocol: "hysteria2", Remark: "h2",
		TLS: &tlsConfig{ServerName: "localhost", SelfSigned: true, CertSha256: "AABB"},
	}, c, "1.2.3.4")
	if !strings.HasPrefix(hy, "hysteria2://pw-1@1.2.3.4:24560?") {
		t.Errorf("hysteria2 链接格式不对: %s", hy)
	}
	for _, want := range []string{"sni=localhost", "insecure=1", "pinSHA256=AABB", "alpn=h3"} {
		if !strings.Contains(hy, want) {
			t.Errorf("自签 Hy2 链接缺少 %s: %s", want, hy)
		}
	}

	hy2 := shareLink(&nativeInbound{
		Port: 24560, Protocol: "hysteria2", Remark: "h2",
		TLS: &tlsConfig{ServerName: "example.com", SelfSigned: false},
	}, c, "1.2.3.4")
	for _, want := range []string{"sni=example.com", "alpn=h3"} {
		if !strings.Contains(hy2, want) {
			t.Errorf("正式证书 Hy2 链接缺少 %s: %s", want, hy2)
		}
	}
	for _, notWant := range []string{"insecure=1", "pinSHA256="} {
		if strings.Contains(hy2, notWant) {
			t.Errorf("正式证书 Hy2 链接不应带 %s: %s", notWant, hy2)
		}
	}
}
