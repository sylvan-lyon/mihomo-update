package run

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sylvan-lyon/mihomo-update/internal/args"
	"github.com/sylvan-lyon/mihomo-update/internal/errors"
	"github.com/sylvan-lyon/mihomo-update/internal/helper"
	"golang.org/x/sync/errgroup"
)

// Run 执行配置更新流程
func Run(ctx context.Context, args *args.Args) error {
	baseDir := args.Path

	group, _ := errgroup.WithContext(ctx)

	var (
		remoteYAML any
		localYAML  any
		remoteErr  error
		localErr   error
	)

	// BEGIN sync
	// remoteYAML, remoteErr = fetchRemoteYAML(ctx, baseDir, args.URL, args.Force, args.Timeout, args.UserAgent)
	// localYAML, localErr = readLocalYAML(ctx, baseDir)
	//
	// if err := errors.Join(remoteErr, localErr); err != nil {
	// 	return err
	// }
	// END sync

	// BEGIN async
	group.Go(func() error {
		remoteYAML, remoteErr = fetchRemoteYAML(ctx, baseDir, args.URL, args.Force, args.Timeout, args.UserAgent)
		return remoteErr
	})

	group.Go(func() error {
		localYAML, localErr = readLocalYAML(ctx, baseDir)
		return localErr
	})

	if err := group.Wait(); err != nil {
		return err
	}
	// END async

	mergedYAML, err := helper.MergeYAML(localYAML, remoteYAML, args.MergeStrategy)
	if err != nil {
		return err
	}

	return writeMergedYAML(ctx, baseDir, mergedYAML)
}

// fetchRemoteYAML 获取远程 YAML 配置（带缓存逻辑）
func fetchRemoteYAML(ctx context.Context, baseDir, url string, force bool, timeout uint64, userAgent string) (any, error) {

	var yaml any
	cacheDir := helper.GetCacheDir(baseDir)

	if force {
		return fetchAndCache(ctx, cacheDir, url, timeout, userAgent)
	} else {
		entry, err := helper.ReadCache(cacheDir, url)
		if err != nil {
			return fetchAndCache(ctx, cacheDir, url, timeout, userAgent)
		}

		if ok, err := entry.IsValid(); !ok || err != nil {
			return fetchAndCache(ctx, cacheDir, url, timeout, userAgent)
		} else {
			yaml = entry.Data
		}
	}

	return yaml, nil
}

// 新增辅助函数
func fetchAndCache(ctx context.Context, cacheDir, url string, timeout uint64, userAgent string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "在试图获取订阅并缓存时，收到取消信号")
	default:
		yaml, err := helper.FetchRemoteYAML(url, time.Duration(timeout), userAgent)
		if err != nil {
			return nil, errors.Wrap(err, "获取订阅失败")
		}

		if err := helper.WriteCache(cacheDir, yaml, url); err != nil {
			skippableErr := errors.MarkSkippable(err)
			fmt.Printf("警告：缓存写入失败（可跳过）: %v\n", skippableErr)
		}

		return yaml, nil
	}
}

// readLocalYAML 读取本地服务器配置文件
//
// 这是一个辅助函数，封装本地配置读取的逻辑
//
// 参数：
//   - ctx: context.Context，用于传递取消信号
//   - baseDir: 基础目录路径
//
// 返回值：
//   - any: YAML 数据
//   - error: 读取过程中的错误（可能是可跳过错误）
func readLocalYAML(ctx context.Context, baseDir string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.Wrap( ctx.Err(), "读取本地服务器配置文件时被取消")
	default:
		yamlPath := filepath.Join(baseDir, "mihomo-server.yaml")
		if !helper.FileExists(yamlPath) {
			return nil, errors.MarkSkippable(errors.New("未发现本地配置文件"))
		}

		yaml, err := helper.ReadYAMLFile(yamlPath)
		if err != nil {
			return nil, errors.Wrap(err, "在读取本地配置文件时")
		}

		return yaml, nil
	}
}

// writeMergedYAML 写入合并后的配置文件
//
// 参数：
//   - ctx: context.Context，用于传递取消信号
//   - baseDir: 基础目录路径
//   - data: 要写入的 YAML 数据
//
// 返回值：
//   - error: 写入过程中的错误
func writeMergedYAML(ctx context.Context, baseDir string, data any) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		configPath := filepath.Join(baseDir, "config.yaml")

		if err := helper.EnsureDir(baseDir); err != nil {
			return errors.Wrap(err, "创建配置目录时")
		}

		if err := helper.WriteYAMLFile(configPath, data); err != nil {
			return errors.Wrap(err, "在写入最终配置文件时")
		}

		return nil
	}
}

// 以下是一个使用 sync.WaitGroup 的替代实现示例（供学习参考）
//
// 注意：在实际代码中，使用 errgroup 更符合需求，因为我们需要错误传播
// 这个示例仅用于演示 sync.WaitGroup 的用法
func runWithWaitGroup(ctx context.Context, args *args.Args) error {
	var wg sync.WaitGroup
	var remoteYAML, localYAML any
	var remoteErr, localErr error

	// 启动远程获取 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 这里调用 fetchRemoteYAML
		remoteYAML, remoteErr = fetchRemoteYAML(ctx, args.Path, args.URL, args.Force, args.Timeout, args.UserAgent)
	}()

	// 启动本地读取 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 这里调用 readLocalYAML
		localYAML, localErr = readLocalYAML(ctx, args.Path)
	}()

	// 等待两个 goroutines 完成
	wg.Wait()

	// 手动检查错误
	if remoteErr != nil {
		return remoteErr
	}
	if localErr != nil {
		return localErr
	}

	// 继续合并和写入...
	_ = remoteYAML
	_ = localYAML
	return nil
}

// 以下是一个使用 channel 传递结果的示例（供学习参考）
//
// 注意：当需要更复杂的结果传递或错误处理时，channel 是更好的选择
// 特别是当 goroutines 数量动态变化时
func runWithChannels(ctx context.Context, args *args.Args) error {
	// 创建带缓冲的 channel，避免 goroutines 阻塞
	remoteCh := make(chan struct {
		data any
		err  error
	}, 1)
	localCh := make(chan struct {
		data any
		err  error
	}, 1)

	// 启动远程获取 goroutine
	go func() {
		data, err := fetchRemoteYAML(ctx, args.Path, args.URL, args.Force, args.Timeout, args.UserAgent)
		remoteCh <- struct {
			data any
			err  error
		}{data, err}
	}()

	// 启动本地读取 goroutine
	go func() {
		data, err := readLocalYAML(ctx, args.Path)
		localCh <- struct {
			data any
			err  error
		}{data, err}
	}()

	// 从 channels 接收结果
	remoteResult := <-remoteCh
	localResult := <-localCh

	// 检查错误
	if remoteResult.err != nil {
		return remoteResult.err
	}
	if localResult.err != nil {
		return localResult.err
	}

	// 继续合并和写入...
	_ = remoteResult.data
	_ = localResult.data
	return nil
}
