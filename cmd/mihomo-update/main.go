package main

import (
	// 标准库导入
	"context"
	"fmt"
	"os"

	// 第三方库导入 (将在后续阶段添加)
	// "github.com/spf13/cobra"

	// 内部包导入 (将在后续阶段创建)
	"github.com/sylvan-lyon/mihomo-update/internal/args"
	"github.com/sylvan-lyon/mihomo-update/internal/run"
	// "github.com/sylvan-lyon/mihomo-update/internal/i18n"
)

func main() {
	cmdArgs, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数解析错误: %v\n", err)
		os.Exit(1)
	}

	err = run.Run(context.Background(), cmdArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "程序执行错误: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs() (*args.Args, error) {
	cmdArgs, err := args.Parse()
	if err != nil {
		return nil, err
	}
	return cmdArgs, nil
}
