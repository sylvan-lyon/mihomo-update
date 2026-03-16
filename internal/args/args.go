// args.go - 命令行参数解析
//
// 本文件使用 cobra 库实现命令行参数解析。cobra 是一个流行的 Go 命令行库，提供：
// 1. 子命令支持 (如 `git commit`, `git push`)
// 2. 自动生成帮助信息
// 3. 参数验证和类型转换
// 4. Bash 自动补全生成
//
// 与标准库 flag 包相比，cobra 更适合复杂的 CLI 应用。
// 基本概念：
// - Command: 命令，如 `mihomo-update`
// - Flags: 标志，如 `--url`, `--path`
// - Args: 位置参数（本项目未使用）
//
// 参考 Rust 版本: src/args.rs
// Rust 使用 clap 库，cobra 是 Go 中的类似选择。

package args

import (
	// 第三方库导入
	"fmt"

	"github.com/spf13/cobra"
)

// MergeStrategy 定义配置合并的三种策略。
// Go 没有枚举类型，通常使用 iota 常量模拟枚举。
// iota 是 Go 的常量计数器，在 const 块中从 0 开始自增。
// 注意：每个 const 块内 iota 独立计数，从 0 开始。
type MergeStrategy int

const (
	// Keep 策略：尽量保留本地值，仅在双方均为映射时递归合并
	// 对应 Rust 的 MergeStrategy::Keep
	Keep MergeStrategy = iota

	// MergeAll 策略：保留本地标量值，追加远程序列（融合所有选项）
	// 对应 Rust 的 MergeStrategy::MergeAll
	MergeAll

	// Force 策略：递归覆盖本地值（映射递归合并）
	// 对应 Rust 的 MergeStrategy::Force
	Force
)

// String 方法实现 fmt.Stringer 接口
// 使 MergeStrategy 可以转换为字符串显示
// 在 cobra 中，这允许枚举值在帮助信息中正确显示
func (m MergeStrategy) String() string {
	switch m {
	case Keep:
		return "keep"
	case MergeAll:
		return "keepall"
	case Force:
		return "force"
	default:
		return "unknown"
	}
}

// Set 方法实现 flag.Value 接口
// 允许 cobra 将字符串参数转换为 MergeStrategy 类型
// 此方法用于命令行参数解析时的类型转换
func (m *MergeStrategy) Set(s string) error {
	switch s {
	case "keep":
		*m = Keep
	case "keepall":
		*m = MergeAll
	case "force":
		*m = Force
	default:
		return fmt.Errorf("无效的合并策略: %s", s)
	}
	return nil
}

// Type 方法实现 flag.Value 接口
// 返回值的类型名称，用于帮助信息
func (m *MergeStrategy) Type() string {
	return "MergeStrategy"
}

// Args 结构体包含所有命令行参数
// 结构体字段对应命令行标志
// 字段标签（tag）`cobra:"..."` 仅用于文档说明，实际参数绑定在 Parse 函数中完成
type Args struct {
	// URL 是订阅地址，必须提供
	// 对应 Rust 的 `url: String`
	// 短标志 `-u`，长标志 `--url`
	URL string `cobra:"url,u,订阅地址"`

	// Path 是配置文件目录路径，必须提供
	// 对应 Rust 的 `path: String`
	// 短标志 `-p`，长标志 `--path`
	Path string `cobra:"path,p,配置文件路径"`

	// Force 强制更新，即使缓存有效
	// 对应 Rust 的 `force: bool`
	// 短标志 `-f`，长标志 `--force`
	Force bool `cobra:"force,f,强制更新"`

	// MergeStrategy 配置合并策略
	// 对应 Rust 的 `merge_strategy: MergeStrategy`
	// 短标志 `-M`，长标志 `--merge-strategy`
	// 默认值 "keep"
	MergeStrategy MergeStrategy `cobra:"merge-strategy,M,合并策略"`

	// Timeout 网络请求超时时间（秒）
	// 对应 Rust 的 `timeout: u64`
	// 长标志 `--timeout`
	// 默认值 60
	Timeout uint64 `cobra:"timeout,网络请求超时(秒)"`

	// UserAgent HTTP 请求的 User-Agent 头
	// 对应 Rust 的 `user_agent: String`
	// 长标志 `--user-agent`
	// 默认值 "clash-verge/v2.4.6"
	UserAgent string `cobra:"user-agent,User-Agent 头"`

	// Lang 覆盖语言设置（如 "zh-CN", "en"）
	// 对应 Rust 的 `lang: Option<String>`
	// 长标志 `--lang`
	// 可选参数，使用指针表示可选性
	Lang *string `cobra:"lang,覆盖语言设置"`
}

// Parse 解析命令行参数并返回 Args 结构体
// 这是主要的导出函数，在 main.go 中调用
func Parse() (*Args, error) {
	// 创建 Args 结构体实例
	args := &Args{}

	// 创建 root command
	rootCmd := &cobra.Command{
		Use:   "mihomo-update",
		Short: "更新你的 Clash 订阅",
		Long: `更新你的 Clash 订阅
并将其与本地 Mihomo 配置合并`,
		// Run 函数将在参数解析成功后执行
		// 这里可以留空，或者设置一个默认行为
		Run: func(cmd *cobra.Command, args []string) {
			// 业务逻辑在 run 包中实现
		},
	}

	// 添加命令行标志
	// URL 参数：必须提供，短标志 -u，长标志 --url
	rootCmd.Flags().StringVarP(&args.URL, "url", "u", "", "订阅地址")
	rootCmd.MarkFlagRequired("url")

	// Path 参数：必须提供，短标志 -p，长标志 --path
	rootCmd.Flags().StringVarP(&args.Path, "path", "p", "", "配置文件目录路径")
	rootCmd.MarkFlagRequired("path")

	// Force 参数：布尔值，短标志 -f，长标志 --force，默认 false
	rootCmd.Flags().BoolVarP(&args.Force, "force", "f", false, "强制更新（忽略缓存）")

	// MergeStrategy 参数：自定义类型，短标志 -M，长标志 --merge-strategy
	// 使用 VarP 因为 MergeStrategy 实现了 flag.Value 接口
	// 默认值为 Keep（零值），对应字符串 "keep"
	rootCmd.Flags().VarP(&args.MergeStrategy, "merge-strategy", "M", "合并策略 (keep|keepall|force)")

	// Timeout 参数：无短标志，长标志 --timeout，默认 60
	rootCmd.Flags().Uint64Var(&args.Timeout, "timeout", 60, "网络请求超时时间（秒）")

	// UserAgent 参数：无短标志（-u 已被 URL 使用），长标志 --user-agent
	rootCmd.Flags().StringVar(&args.UserAgent, "user-agent", "clash-verge/v2.4.6", "HTTP User-Agent 头")

	// Lang 参数：可选参数，使用指针，无短标志，长标志 --lang
	// StringVar 第一个参数是 *string，需要非nil指针
	// 初始化 Lang 字段为一个指向空字符串的指针
	if args.Lang == nil {
		args.Lang = new(string)
	}
	rootCmd.Flags().StringVar(args.Lang, "lang", "", "覆盖语言设置（如 zh-CN, en）")

	// 执行命令解析
	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}

	// 校验参数合法性
	if err := validateArgs(args); err != nil {
		return nil, err
	}

	return args, nil
}

// validateArgs 验证参数的有效性
// 可以在 Parse 函数中调用，确保参数合法
func validateArgs(args *Args) error {
	if args.URL == "" {
		return fmt.Errorf("必须提供订阅地址 (--url)")
	}
	if args.Path == "" {
		return fmt.Errorf("必须提供配置文件路径 (--path)")
	}
	if args.Timeout == 0 {
		return fmt.Errorf("超时时间必须大于 0")
	}
	return nil
}

// 术语表（中英对照）：
// - cobra: cobra 库名，保留英文
// - Command: 命令，cobra 的核心概念
// - Flags: 标志，命令行参数的一种
// - iota: iota 常量计数器，保留英文
// - Stringer: Stringer 接口，定义 String() 方法
// - flag.Value: flag.Value 接口，用于自定义类型解析
// - struct tag: 结构体标签，用于元数据注解
// - pointer: 指针，表示可选字段（nil 表示不存在）

// cobra 常用 API 速查：
// - cobra.Command: 命令定义结构体
// - Flags(): 获取命令的标志集
// - StringVarP(): 绑定字符串参数（有短标志）
// - BoolVarP(): 绑定布尔参数
// - Uint64Var(): 绑定 uint64 参数（无短标志）
// - MarkFlagRequired(): 标记参数为必填
// - SetDefault(): 设置默认值（已废弃，直接在 Var 函数中设置）
// - Execute(): 执行命令解析

// 最佳实践提示：
// 1. 为重要参数提供短标志和长标志
// 2. 为所有参数提供清晰的帮助信息
// 3. 验证参数的合理范围（如超时时间 > 0）
// 4. 使用指针表示可选参数（nil 表示未提供）
// 5. 实现自定义类型的 String() 和 Set() 方法以支持枚举

// 扩展学习：
// 1. 子命令系统：如 `mihomo config`, `mihomo update`
// 2. 环境变量支持：cobra 支持从环境变量读取默认值
// 3. 配置文件支持：使用 viper 库与 cobra 集成
// 4. Bash 补全：`cobra init` 可以生成补全脚本
