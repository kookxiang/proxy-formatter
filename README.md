# Proxy Formatter

一个通过 HTTP 按需生成代理订阅内容的小工具。

它会读取本地配置文件，按顺序执行其中的指令，然后把结果输出成目标格式。目前内置了两类能力：

- 指令：抓取、过滤、改名、改请求头、解析 DNS、检查订阅流量等
- 格式化器：输出为 `clash`、`surge` 或 `loon` 订阅内容

本项目使用 `GPL-3.0` 许可证发布。

## 工作方式

每个配置文件都是一个按行执行的流水线。

- 一行一个指令
- 按出现顺序执行
- 空行会忽略
- 以 `#` 开头的行会被当作注释
- 参数支持双引号，适合包含空格的内容

例如：

```text
# 从远端获取原始订阅
fetch https://example.com/proxies.yaml

# 设置请求头
header User-Agent "Clash.Meta"

# 只保留日本节点
include /日本/

# 给节点名前面加一个标签
prefix Tokyo

# 输出为 Clash 格式
clash
```

## 常见用法

### 过滤并改名

```text
fetch https://example.com/proxies.yaml
include /香港|日本|新加坡/
exclude /剩余流量|即将到期/
emoji
replace IEPL IPLC
suffix Premium
clash
```

### 使用中转代理抓取订阅

```text
proxy http://127.0.0.1:7890
fetch https://example.com/proxies.yaml
proxy
clash
```

`proxy` 不带参数时会清空当前抓取代理设置。

### 自动处理过期或流量耗尽的订阅

```text
fetch https://example.com/proxies.yaml
check-traffic
clash
```

如果响应头中的 `Subscription-Userinfo` 表示订阅已经过期，或者总流量已经用尽，程序会清空原有节点，并注入一个提示原因的占位节点，避免部分客户端因为空列表而报错。

## 支持的指令

下面的指令都会按照配置中的顺序执行。

### fetch <url>

从指定 URL 获取订阅内容，并在本地缓存响应。

说明：

- 支持 YAML 订阅
- 如果 YAML 解析失败，会回退尝试解析为 Base64 订阅
- 当缓存需要刷新时，会重新拉取远端内容
- 如果之前设置过 `proxy` 或 `header`，抓取时会带上这些配置

示例：

```text
fetch https://example.com/proxies.yaml
```

### fetch-once <url>

与 `fetch` 类似，但只在本地没有缓存时拉取远端内容；一旦缓存存在，即使缓存可刷新，也优先使用缓存。

这个模式适合你希望手动控制刷新时机的场景。

### include <pattern>

只保留命中模式的节点名。

### exclude <pattern>

移除命中模式的节点名。

模式规则：

- 普通字符串：按包含关系匹配
- `/.../`：按正则表达式匹配

示例：

```text
include 日本
exclude /过期|失效|测试/
```

### emoji

移除节点名称中的国旗 emoji。

示例：

```text
emoji
```

### replace <from> <to>

把节点名称中的所有 `<from>` 替换成 `<to>`。

示例：

```text
replace IEPL IPLC
```

### prefix <content>

给节点名称前面添加前缀。实现上这是一个“切换”操作：

- 如果节点名还没有这个前缀，就加上
- 如果已经有这个前缀，就移除

程序会自动在前缀末尾补一个空格。

示例：

```text
prefix HK
```

### suffix <content>

给节点名称后面添加后缀，同样是“切换”操作：

- 如果节点名还没有这个后缀，就加上
- 如果已经有这个后缀，就移除

程序会自动在后缀前面补一个空格。

示例：

```text
suffix Premium
```

### freeze

把当前所有节点标记为冻结。

冻结节点不会再被后续过滤指令移除，但仍然会出现在最终输出中。

这个指令主要适合在同一个配置文件里串联多个订阅源使用：

- 先 `fetch` 第一个网址
- 对这一批节点做过滤、改名等处理
- 执行一次 `freeze`
- 再 `fetch` 下一个网址，并继续处理后面这批节点

这样前一批已经整理好的节点可以被“锁住”，不会被后续针对新订阅的过滤条件误伤。

示例：

```text
fetch https://example.com/a.yaml
include /香港|日本/
prefix A
freeze

fetch https://example.com/b.yaml
include /新加坡|美国/
prefix B
clash
```

### header <key> [value]

设置抓取订阅时使用的请求头。

规则：

- `header <key> <value>`：设置请求头
- `header <key>`：删除该请求头

示例：

```text
header User-Agent "Clash.Meta"
header Authorization "Bearer token"
header Authorization
```

### proxy [address]

设置抓取远程订阅时使用的 HTTP 代理。

规则：

- `proxy http://127.0.0.1:7890`：设置代理
- `proxy`：清空代理

### resolve [dns]

把节点服务器地址解析成 IP，并写回配置。

默认 DNS 参数为 `119.29.29.29`。

行为说明：

- 会优先返回 IPv4
- 对部分 TLS / Reality / WebSocket 相关协议，会在解析前保留原始域名到 `SNI` 或 `ServerName`
- 解析结果会缓存

示例：

```text
resolve
resolve 1.1.1.1
```

### check-traffic

检查远程响应头里的 `Subscription-Userinfo`。

当满足以下任一条件时：

- 订阅已过期
- 上传流量 + 下载流量大于等于总流量

程序会清空当前节点列表，并注入一个本地假的占位节点，节点名会显示为原因，例如“订阅已过期”或“流量已用尽”。

## 输出格式

### clash

把当前节点输出为 Clash / Mihomo 风格的 YAML：

```yaml
proxies:
  - name: Example
    type: ss
    server: 1.2.3.4
    port: 443
```

响应头会设置为：

```text
Content-Type: text/yaml; charset=utf-8
```

### surge [version]

把当前节点输出为 Surge `[Proxy]` 段格式。

当前实现只支持部分代理类型；如果遇到未实现的类型，会直接返回错误。

响应头会设置为：

```text
Content-Type: text/plain; charset=utf-8
```

### loon

把当前节点输出为 Loon 节点行格式。

当前实现只支持以下代理类型：

- `Shadowsocks`
- `Trojan`
- `VLESS`
- `VMess`

其中：

- `Shadowsocks` 支持基础格式，以及 `simple-obfs` / `shadow-tls` 常见字段映射
- `Trojan` 支持普通 TLS 与 `ws`
- `VLESS` 支持 `tcp`、`ws`、`http`，以及 `reality` 所需的 `public-key` / `short-id`
- `VMess` 支持 `tcp`、`ws`、`http`，以及 `h2 -> http` 的 Loon 映射

如果遇到其它代理类型或未覆盖的传输方式，会直接返回错误。

响应头会设置为：

```text
Content-Type: text/plain; charset=utf-8
```

## 部署

### 1. 构建

```bash
go build -trimpath -o proxy-formatter .
```

### 2. 编写配置文件

假设你在配置目录下放了一个 `config.txt`：

```text
fetch https://example.com/proxies.yaml
check-traffic
exclude /过期|失效/
include /香港|日本|新加坡/
emoji
prefix IPLC
clash
```

Loon 输出示例：

```text
fetch https://example.com/proxies.yaml
include /香港|日本/
loon
```

### 3. 启动服务

```bash
./proxy-formatter -dir /path/to/config -cache-dir /path/to/cache -port 15725
```

参数说明：

- `-dir`：配置文件所在目录，默认是当前工作目录
- `-cache-dir`：HTTP 抓取缓存目录，默认是系统临时目录下的 `cache`
- `-editor`：启用在线编辑器，注意需要配合 `-password` 使用
- `-password`：保存配置时需要提供的密码
- `-port`：HTTP 服务端口，默认是 `15725`

### 4. 访问生成结果

如果配置文件名是 `config.txt`，启动后可直接访问：

```text
http://localhost:15725/config.txt
```

程序会读取对应文件，逐行执行配置，并返回最终生成的订阅内容。

如果启用了 `-editor`，也可以直接在首页新建规则：

```text
http://localhost:15725/
```

如果需要重新编辑由编辑器创建的配置，可以访问：

```text
http://localhost:15725/?edit=<配置文件名>
```

保存时需要提供 `-password` 对应的密码。

## 注意事项

- 配置文件是按顺序执行的，指令顺序会直接影响结果
- 如果你需要处理带空格的参数，请使用双引号
- `fetch-once` 不会主动刷新已有缓存
- `surge` 当前不是全协议覆盖，复杂订阅建议优先用 `clash`

## 系统服务

仓库中附带了 `proxy-formatter.service`，可用于 systemd 部署。

## License

本项目基于 `GNU General Public License v3.0` 发布，详见 [LICENSE](/Users/kookxiang/workspace/proxy-formatter/LICENSE)。
