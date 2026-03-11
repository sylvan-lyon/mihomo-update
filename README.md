# mihomo-update

A Rust CLI tool for updating Clash/Mihomo subscription configurations. This tool fetches subscription content from a URL, merges it with your local configuration using various strategies, and outputs the result.

[中文版本](README_zh-CN.md)

## Features

- Fetch YAML configuration from subscription URLs
- Cache subscription content to avoid unnecessary downloads
- Merge remote and local configurations with different strategies:
  - `Keep`: Preserve local values, replace lists with remote ones
  - `KeepAll`: Preserve local values, append to lists
  - `Force`: Override local values with remote ones
- Support for custom User-Agent headers
- Configurable request timeout
- Internationalization support (English and Chinese)

## Installation

```bash
# Build from source
cargo build --release
```

The binary will be located at `target/release/mihomo_update`.

## Usage

```bash
# Basic usage
cargo run --release -- \\n  --url "https://example.com/sub" \\n  --path /path/to/config

# With custom options
cargo run --release -- \\n  --url "https://example.com/sub" \\n  --path /path/to/config \\n  --force \\n  --timeout 60 \\n  --user-agent "clash-verge/v2.4.6"
```

## Command Line Options

| Option | Description |
|--------|-------------|
| `-u, --url SUB` | Subscription URL to fetch configuration from |
| `-p, --path PATH` | Path to Mihomo configuration file |
| `-f, --force` | Force update even if cache exists |
| `--timeout SECS` | Network request timeout in seconds (default: 60) |
| `--user-agent UA` | Custom HTTP User-Agent (default: clash-verge/v2.4.6) |
| `--lang LANG` | Override language used for CLI messages |

## Configuration Flow

1. Check if cache should be updated (expires after 24 hours)
2. If forced or cache expired, fetch subscription from URL
3. Read local server configuration
4. Merge configurations using the `Keep` strategy
5. Write merged configuration to output file

## Development

```bash
# Run tests
cargo test

# Run with formatting check
cargo fmt -- --check

# Run with linter
cargo clippy
```
