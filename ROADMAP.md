# Go 移植学习路线图 (Roadmap for Go Porting)

基于 Rust 版本的 mihomo-update 项目，设计一个 **10阶段** 的平滑学习曲线，从 Go 基础到完整项目移植。每个阶段都聚焦特定 Go 概念，并创建对应的骨架文件（stub files）附带详细中文教学注释。

## 项目概览

Rust 原项目是一个 **Mihomo/Clash 订阅更新工具**，核心功能包括：
- 从 URL 获取 YAML 配置
- 与本地配置合并（三种策略：Keep、KeepAll、Force）
- 24小时缓存机制
- 完整的错误处理和国际化（中英文）

## 阶段概览

| 阶段 | 核心学习目标 | 对应 Rust 模块 | 关键 Go 概念 | 输出文件 |
|------|-------------|----------------|-------------|----------|
| **0** | 项目初始化与模块管理 | Cargo.toml | `go.mod`, 项目结构, 导入路径 | `go.mod`, `cmd/mihomo-update/main.go` |
| **1** | 命令行参数解析 | `args.rs` | `flag` 包, 结构体标签, `cobra` 库 | `internal/args/args.go` |
| **2** | 错误处理基础 | `errors.rs` | `error` 接口, 错误包装, 自定义错误 | `internal/errors/errors.go` |
| **3** | 文件操作与 YAML | `helper.rs` (部分) | `os`, `io`, `yaml.v3`, 序列化 | `internal/helper/io.go`, `internal/helper/yaml.go` |
| **4** | HTTP 客户端 | `helper.rs` (fetch) | `net/http`, 超时, User-Agent | `internal/helper/http.go` |
| **5** | 合并策略算法 | `helper.rs` (merge) | 递归, 类型断言, `interface{}` | `internal/helper/merge.go` |
| **6** | 缓存与时间处理 | `run.rs` (缓存) | `time` 包, 文件时间戳 | `internal/helper/cache.go` |
| **7** | 并发与并行 | `run.rs` (并行) | goroutine, `sync.WaitGroup`, `errgroup` | `internal/run/run.go` |
| **8** | 国际化 (i18n) | `locales/` | 文本本地化, 模板替换 | `internal/i18n/i18n.go`, `locales/` |
| **9** | 测试与文档 | `tests.rs` | `testing` 框架, 表格驱动测试 | `internal/helper/merge_test.go` |

## 详细阶段说明

### 阶段 0: 环境设置与项目初始化
**学习目标**: Go 模块系统、项目布局约定、导入路径  
**教学重点**: 
- `go mod init` 与模块版本管理
- Go 标准项目结构 (`cmd/`, `internal/`, `pkg/`)
- 包声明与导入路径规则
- `main` 包与可执行文件入口

**输出文件**:
- `go.mod` - Go 模块定义文件
- `cmd/mihomo-update/main.go` - 程序入口点骨架

### 阶段 1: 命令行参数解析
**学习目标**: 标准库 `flag` 与高级库 `cobra` 的使用  
**教学重点**:
- `flag` 包基础用法 (位置参数、标志)
- 结构体标签 `flag:"..."` 的映射
- `cobra` 库的层级命令与自动帮助生成
- 枚举类型 (`MergeStrategy`) 的实现方式 (使用 `iota`)

### 阶段 2: 错误处理基础
**学习目标**: Go 错误处理哲学与最佳实践  
**教学重点**:
- `error` 接口与自定义错误类型
- 错误链: `fmt.Errorf` 与 `%w` 包装符
- 错误检查惯用法 (`if err != nil`)
- 可跳过错误 (`skippable`) 的模式实现
- 错误上下文信息携带

### 阶段 3: 文件操作与 YAML 处理
**学习目标**: 文件 I/O 与 YAML 序列化库  
**教学重点**:
- `os.Open`/`os.Create` 与 `io.ReadAll`/`io.WriteString`
- `gopkg.in/yaml.v3` 的 `Marshal`/`Unmarshal`
- YAML 节点类型 (`yaml.Node`) 与递归遍历
- 文件路径操作 (`path/filepath`)

### 阶段 4: HTTP 客户端
**学习目标**: `net/http` 标准库与客户端配置  
**教学重点**:
- `http.Client` 配置 (超时、代理、TLS)
- 请求构建: `http.NewRequest` 与 Header 设置
- 响应处理: 状态码检查、Body 读取与关闭
- User-Agent 自定义与超时控制

### 阶段 5: 合并策略算法
**学习目标**: 递归算法与动态类型处理  
**教学重点**:
- `interface{}` 类型断言 (`value.(type)`)
- 递归函数设计模式
- 映射 (`map[interface{}]interface{}`) 与切片 (`[]interface{}`) 操作
- 三种策略 (`Keep`, `KeepAll`, `Force`) 的算法差异

### 阶段 6: 缓存与时间处理
**学习目标**: `time` 包与文件系统状态检查  
**教学重点**:
- `time.Time` 与持续时间 (`time.Duration`)
- `os.Stat` 获取文件修改时间
- 24小时缓存过期逻辑实现
- JSON 序列化用于简单结构化存储

### 阶段 7: 并发与并行
**学习目标**: goroutine 与同步原语  
**教学重点**:
- goroutine 启动与等待 (`sync.WaitGroup`)
- 错误组模式 (`errgroup.Group`)
- 并行获取与读取的协调
- 竞态条件避免与数据同步

### 阶段 8: 国际化 (i18n)
**学习目标**: 文本本地化系统设计  
**教学重点**:
- 键值对翻译存储 (JSON/YAML)
- 模板变量替换 (`strings.ReplaceAll`)
- 语言环境检测与回退机制
- 与命令行参数的集成

### 阶段 9: 测试与文档
**学习目标**: Go 测试框架与文档生成  
**教学重点**:
- 表格驱动测试 (`[]struct` 测试用例)
- 测试辅助函数与子测试 (`t.Run`)
- `go test` 命令与覆盖率 (`-cover`)
- GoDoc 注释规范与文档生成

## 双语术语表

| 英文术语 | 中文翻译 (客观) | 备注 |
|----------|----------------|------|
| package | 包 | 标准翻译，Go 代码组织的基本单位 |
| import | 导入 | 标准翻译，引入外部包 |
| struct | 结构体 | 标准翻译，复合数据类型 |
| interface | 接口 | 标准翻译，定义方法集合的类型 |
| goroutine | goroutine | 保留英文，Go 的轻量级线程，无广泛认可中文翻译 |
| channel | 通道 | 常用翻译，goroutine 间的通信机制 |
| flag | 标志 (命令行) | 根据上下文，命令行参数的一种 |
| cobra | cobra (库名) | 保留英文，流行的命令行库 |
| YAML | YAML | 保留英文，数据序列化格式 |
| marshaling | 序列化 | 标准翻译，将数据结构转换为字节流 |
| unmarshaling | 反序列化 | 标准翻译，将字节流转换为数据结构 |
| recursive | 递归 | 标准翻译，函数调用自身的编程技巧 |
| concurrency | 并发 | 标准翻译，同时处理多个任务的能力 |
| parallelism | 并行 | 标准翻译，同时执行多个任务，与并发有细微区别 |
| iota | iota | 保留英文，Go 的常量计数器 |
| interface{} | 空接口 | 标准翻译，可以保存任何类型的值 |
| slice | 切片 | 标准翻译，动态数组的视图 |
| map | 映射 | 标准翻译，键值对集合 |

## 学习建议

1. **循序渐进**: 按阶段顺序学习，每个阶段确保理解核心概念
2. **实践为主**: 在骨架文件中添加实现，理解注释中的教学要点
3. **对比学习**: 对照 Rust 原版代码，理解相同功能在不同语言中的实现差异
4. **测试驱动**: 阶段 9 的测试可以提前编写，帮助验证实现正确性
5. **查阅文档**: 善用 `go doc` 命令和官方文档 ([golang.org](https://golang.org/doc/))

## 扩展资源

- [Go 官方教程](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go 标准库文档](https://pkg.go.dev/std)
- [Go by Example](https://gobyexample.com/)
- [Cobra 库文档](https://github.com/spf13/cobra)

## 项目结构预览

```
mihomo-update/
├── ROADMAP.md          # 本文件
├── go.mod              # Go 模块定义 (阶段 0)
├── cmd/
│   └── mihomo-update/
│       └── main.go     # 程序入口点 (阶段 0)
├── internal/
│   ├── args/           # 命令行参数解析 (阶段 1)
│   ├── errors/         # 错误处理 (阶段 2)
│   ├── helper/         # 辅助函数 (阶段 3-6)
│   ├── run/            # 主业务逻辑 (阶段 7)
│   └── i18n/           # 国际化 (阶段 8)
├── locales/            # 翻译文件目录 (阶段 8)
└── [测试文件]          # 测试文件 (阶段 9)
```

---
*最后更新: 2026年3月12日*  
*基于 Rust 分支: `rust`*  
*目标 Go 版本: 1.21+*