// go.mod - Go 模块定义文件
// 
// 本文件定义了 Go 模块的基本信息，包括模块路径、Go 版本和依赖项。
// Go 模块是 Go 1.11 引入的官方依赖管理系统，替代了旧的 GOPATH 模式。
//
// 语法说明：
// 1. `module` 指令声明模块路径，用于导入路径前缀
// 2. `go` 指令指定最低兼容的 Go 版本
// 3. `require` 指令声明依赖项及其版本
// 4. `replace` 指令可用于本地开发时替换依赖
// 5. `exclude` 指令排除特定版本
//
// 常用命令：
// - `go mod init <module-path>`: 初始化新模块
// - `go mod tidy`: 添加缺失的依赖，移除未使用的依赖
// - `go mod download`: 下载模块到本地缓存
// - `go mod vendor`: 创建 vendor 目录复制依赖
// - `go list -m all`: 列出所有依赖项
//
// 模块路径约定：
// - 通常使用代码仓库的 URL，如 github.com/用户名/仓库名
// - 确保唯一性，避免与其他模块冲突
// - 导入时使用完整模块路径 + 包路径

module github.com/sylvan-lyon/mihomo-update

// Go 版本指令
// 指定项目所需的最低 Go 版本，这里使用 1.21 以支持较新的特性
// 注意：实际使用的 Go 版本可以高于此版本，但不能低于此版本
go 1.21

// 依赖项声明
// 以下依赖项将在后续阶段逐步添加
// 使用 `go get <package>` 命令添加依赖，`go mod tidy` 会自动更新此文件

// 命令行解析库 (阶段 1)
// require github.com/spf13/cobra v1.8.0

// YAML 处理库 (阶段 3)
// require gopkg.in/yaml.v3 v3.0.1

// HTTP 客户端增强 (阶段 4，可选)
// require github.com/go-resty/resty/v2 v2.11.0

// 错误组库 (阶段 7)
// require golang.org/x/sync v0.6.0

// 国际化库 (阶段 8，可选)
// require golang.org/x/text v0.14.0

// 术语表（中英对照）：
// - module: 模块，Go 代码的版本化单元
// - require: 依赖声明，指定所需的包及其版本
// - dependency: 依赖项，项目所依赖的外部包
// - vendor: 供应商目录，用于将依赖项复制到项目内
// - tidy: 整理，清理和同步依赖关系

// 注意：`// indirect` 注释表示间接依赖（被直接依赖的包所依赖）
// 间接依赖通常不需要显式声明，`go mod tidy` 会自动管理