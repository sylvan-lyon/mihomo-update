package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	// 内部包 - 这些包只能在此模块内访问
	"github.com/sylvan/mihomo-update/internal/cli"
	"github.com/sylvan/mihomo-update/internal/config"
	"github.com/sylvan/mihomo-update/internal/http"
	"github.com/sylvan/mihomo-update/internal/merge"
)

// main 是应用程序的入口点。
// 它设置信号处理、初始化依赖项并执行CLI命令。
func main() {
	// 创建可取消的上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理操作系统信号以实现优雅关闭
	setupSignalHandling(cancel)

	// 初始化应用程序配置
	// TODO: 从文件、环境变量或CLI标志加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 使用配置的超时和用户代理初始化HTTP客户端
	// TODO: 创建具有重试逻辑、超时和自定义标头的HTTP客户端
	httpClient := http.NewClient(cfg.HTTPTimeout, cfg.UserAgent)

	// 使用指定策略初始化YAML合并器
	// TODO: 根据配置选择合并策略（Keep、KeepAll、Force）
	merger := merge.NewMerger(cfg.MergeStrategy)

	// 创建并执行CLI命令
	// TODO: 使用Cobra定义命令、标志和帮助文本
	cmd := cli.NewRootCommand(cfg, httpClient, merger)
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "命令执行失败: %v\n", err)
		os.Exit(1)
	}
}

// setupSignalHandling 配置信号处理以实现优雅关闭。
// 当收到SIGINT或SIGTERM信号时，将调用取消函数。
func setupSignalHandling(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("\n收到信号: %v. 正在关闭...\n", sig)
		cancel()
	}()
}

// 展示的最佳实践:
// 1. 使用Context进行取消和超时控制
// 2. 通过信号处理实现优雅关闭
// 3. 依赖注入（配置、HTTP客户端、合并器传递给CLI）
// 4. 提供信息性消息的错误处理
// 5. 关注点分离（main处理设置，CLI处理执行）
// 6. 使用接口提高可测试性（http.Client、merge.Merger）

// 后续步骤:
// 1. 实现config.Load()从YAML、环境变量和标志读取配置
// 2. 实现http.NewClient()，包含适当的超时和重试配置
// 3. 使用策略模式实现merge.NewMerger()
// 4. 使用Cobra库实现cli.NewRootCommand()
// 5. 添加结构化日志记录（使用logrus或zap库）
// 6. 如果需要，添加指标和遥测功能
