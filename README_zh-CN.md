# mihomo-update

一个用于更新 Clash/Mihomo 订阅配置的 Rust CLI 工具。该工具从订阅 URL 获取内容，将其与本地配置合并，并输出结果。

## 功能特性

- 从订阅 URL 获取 YAML 配置
- 缓存订阅内容以避免不必要的下载
- 使用多种策略合并远程和本地配置：
  - `Keep`：保留本地值，用远程列表替换
  - `KeepAll`：保留本地值，将列表追加到末尾
  - `Force`：用远程值覆盖本地值
- 支持自定义 User-Agent 头部
- 可配置的请求超时时间
- 国际化支持（英文和中文）

## 安装方法

```bash
# 从源码构建
cargo build --release
```

二进制文件将位于 `target/release/mihomo_update`。

## 使用方法

```bash
# 基本用法
cargo run --release -- \\\n  --url "https://example.com/sub" \\\n  --path /path/to/config

# 使用自定义选项
cargo run --release -- \\\n  --url "https://example.com/sub" \\\n  --path /path/to/config \\\n  --force \\\n  --timeout 60 \\\n  --user-agent "clash-verge/v2.4.6"
```

## 命令行选项

| 选项 | 说明 |
|--------|-------------|
| `-u, --url SUB` | 订阅地址 |
| `-p, --path PATH` | Mihomo 配置文件路径 |
| `-f, --force` | 即使存在缓存也强制更新 |
| `--timeout SECS` | 网络请求超时时间（秒，默认: 60）|
| `--user-agent UA` | 自定义 HTTP User-Agent （默认: clash-verge/v2.4.6）|
| `--lang LANG` | 指定 CLI 使用的语言 |

## 配置流程

1. 检查缓存是否需要更新（24 小时过期）
2. 如果强制更新或缓存已过期，则从 URL 获取订阅
3. 读取本地服务器配置
4. 使用 `Keep` 策略合并配置
5. 将合并后的配置写入输出文件

## 开发相关

```bash
# 运行测试
cargo test

# 格式检查
cargo fmt -- --check

# 代码检查
cargo clippy
```

[English Version](README.md)