// run.go - 主程序逻辑与并发处理
//
// 本文件实现了 mihomo-update 的核心业务逻辑，重点学习 Go 的并发编程模式。
// 对应 Rust 版本的 `src/run.rs`，但采用 Go 特有的并发原语和错误处理方式。
//
// 并发编程 (Concurrent Programming) 是 Go 的核心特性之一：
// 1. Goroutines: 轻量级线程，由 Go 运行时管理，开销极小（约 2KB 栈空间）
// 2. Channels: 用于 goroutines 之间的通信（CSP 模型）
// 3. sync 包: 提供传统的同步原语（Mutex, WaitGroup, Once 等）
// 4. errgroup 包: 管理一组 goroutines 的错误传播
//
// 本文件将展示三种并发模式：
// 1. 使用 sync.WaitGroup 等待多个 goroutines 完成
// 2. 使用 errgroup.Group 管理 goroutines 的错误传播
// 3. 使用 channel 传递结果和错误（可选）
//
// 与 Rust 对比：
// - Rust 使用 async/await 和 Future，基于轮询 (poll) 的协作式多任务
// - Go 使用 goroutines，基于调度的抢占式多任务
// - Rust 的 tokio::join! 类似 Go 的 sync.WaitGroup
// - Rust 的 Result<T, E> 错误处理 vs Go 的 (value, error) 返回模式
//
// 双语术语表：
// - concurrency: 并发，同时处理多个任务的能力
// - parallelism: 并行，同时执行多个任务（需要多核 CPU）
// - goroutine: Go 的轻量级线程，非操作系统线程
// - channel: 通道，goroutines 之间的通信管道
// - wait group: 等待组，用于等待一组 goroutines 完成
// - mutex: 互斥锁，保护共享资源的同步原语
// - race condition: 竞态条件，多个操作访问共享资源且结果取决于执行顺序
// - data race: 数据竞争，多个 goroutines 并发访问同一内存且至少一个是写入
// - context: 上下文，用于传递截止时间、取消信号和请求范围的值
// - errgroup: 错误组，管理一组相关 goroutines 的错误传播

package run

import (
	// 标准库导入
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	// 第三方库导入
	"golang.org/x/sync/errgroup"

	// 内部包导入
	"github.com/sylvan-lyon/mihomo-update/internal/args"
	"github.com/sylvan-lyon/mihomo-update/internal/errors"
	"github.com/sylvan-lyon/mihomo-update/internal/helper"
)

// Run 是应用程序的主入口点，执行配置更新流程
//
// 功能：
// 1. 根据 force 标志决定从缓存读取还是重新获取远程配置
// 2. 并发获取远程配置和读取本地配置
// 3. 使用指定策略合并配置
// 4. 将合并后的配置写入文件
//
// 参数：
//   - ctx: context.Context，用于传递截止时间和取消信号
//   - args: *args.Args，命令行参数
//
// 返回值：
//   - error: 如果流程中任何步骤失败，返回错误；否则返回 nil
//
// 并发模式选择：
// 1. 如果两个任务相互独立且都需要错误处理，使用 errgroup
// 2. 如果只需等待任务完成而不关心错误，使用 sync.WaitGroup
// 3. 如果任务需要复杂的结果传递，使用 channel
//
// 本例中，远程获取和本地读取都是可能失败的 I/O 操作，因此使用 errgroup
// 同时，我们还需要将两个结果都传递出来，所以使用闭包捕获结果
func Run(ctx context.Context, args *args.Args) error {
	// TODO: 实现路径基础目录
	// 提示: Rust 版本使用 PathBuf::from(&path)，Go 可直接使用 args.Path
	baseDir := args.Path

	// 创建 errgroup，它会自动传播 goroutines 中的错误
	// WithContext 返回一个 Group 和一个派生 Context
	// 当任何一个 goroutine 返回错误时，Group 的 Context 会被取消
	group, ctx := errgroup.WithContext(ctx)

	// 使用闭包捕获两个任务的返回结果
	// 在 Go 中，闭包可以访问外部变量，这是常见的 goroutine 参数传递方式
	var (
		remoteYAML any   // 远程配置数据
		localYAML  any   // 本地配置数据
		remoteErr  error // 远程获取错误
		localErr   error // 本地读取错误
	)

	// TODO: 实现远程配置获取（goroutine 1）
	// 功能: 根据 force 标志决定从缓存读取还是重新获取
	// 参考 Rust 的 fetch_and_cache 和 try_read_from_cache 函数
	// 注意: 在 goroutine 中修改变量需要同步，这里使用闭包直接赋值是安全的
	// 因为每个变量只被一个 goroutine 写入，且 errgroup 保证在 Wait 后读取
	group.Go(func() error {
		remoteYAML, remoteErr = fetchRemoteYAML(ctx, baseDir, args.URL, args.Force, args.Timeout, args.UserAgent)
		// TODO: 实现远程配置获取逻辑
		// 如果 args.Force 为 true，强制重新获取
		// 否则，检查缓存是否有效，有效则读取缓存，无效则重新获取
		// 使用 helper 包中的缓存相关函数
		// 将结果赋值给 remoteYAML，错误赋值给 remoteErr
		// 注意: 返回的错误会被 errgroup 收集
		return remoteErr
	})

	// TODO: 实现本地配置读取（goroutine 2）
	// 功能: 读取本地服务器配置文件
	// 参考 Rust 的 read_yaml(server_file(&base))
	// 使用 helper 包中的文件路径函数和 YAML 读取函数
	group.Go(func() error {
		// TODO: 实现本地配置读取逻辑
		// 构建服务器配置文件路径
		// 读取 YAML 文件
		// 将结果赋值给 localYAML，错误赋值给 localErr
		localYAML, localErr = readLocalYAML(ctx, baseDir)

		return localErr
	})

	// 等待所有 goroutines 完成
	// Wait 会阻塞直到所有 goroutines 完成或某个返回错误
	// 如果所有 goroutines 都成功，返回 nil
	// 如果某个 goroutine 返回错误，返回第一个错误
	if err := group.Wait(); err != nil {
		// TODO: 错误处理
		// 检查错误是否为可跳过的错误（使用 errors.Is）
		// COMMENT 此函数为调度函数，不会再收到 Skippable 的错误
		// 如果不是可跳过错误，返回包装后的错误

		// 我们使用 Join 来收集所有的错误，这应该挺合理吧？
		// return err
	}

	// 检查各个任务的错误（虽然 errgroup 已处理，但我们需要具体错误信息）
	// 注意: errgroup 返回后，我们可以安全地读取各个错误变量
	// 因为所有 goroutines 都已结束
	if remoteErr != nil {
		// TODO: 处理远程获取错误（如果是可跳过错误，可以继续执行）
		// COMMENT 同其他 COMMENT
		// 使用 errors.Is 检查是否为可跳过错误
	}

	if localErr != nil {
		// TODO: 处理本地读取错误
	}

	errors.Join(remoteErr, localErr)

	// TODO: 合并配置
	// 使用 helper.MergeYAML 函数合并 localYAML 和 remoteYAML
	// 传递 args.MergeStrategy 作为合并策略
	mergedYAML, err := helper.MergeYAML(localYAML, remoteYAML, args.MergeStrategy)
	if err != nil {
		// TODO: 错误处理，返回包装后的错误
		return err
	}

	// TODO: 写入合并后的配置
	// 使用 helper 包中的文件写入函数
	// 构建配置文件路径（如 config.yaml）
	// 写入 YAML 文件
	writeMergedYAML(ctx, baseDir, mergedYAML)

	return nil
}

// fetchRemoteYAML 获取远程 YAML 配置（带缓存逻辑）
//
// 这是一个辅助函数，封装远程配置获取的逻辑，可以在 goroutine 中调用
// 它实现了 Rust 版本中 try_read_from_cache 和 fetch_and_cache 的功能
//
// 参数：
//   - ctx: context.Context，用于传递取消信号
//   - baseDir: 基础目录路径
//   - url: 订阅地址
//   - force: 是否强制更新
//   - timeout: 超时时间（秒）
//   - userAgent: User-Agent 头
//
// 返回值：
//   - any: YAML 数据
//
// COMMENT 这里其实已经是主逻辑了，run 函数只是用来调度这两个函数的，
// 所以完全可以把这两个函数的可跳过错误原谅了
//   - error: 获取过程中的错误（不会是可跳过的错误）
func fetchRemoteYAML(ctx context.Context, baseDir, url string, force bool, timeout uint64, userAgent string) (any, error) {
	// TODO: 实现带缓存的远程配置获取
	// 1. 如果 force 为 true，跳过缓存直接获取
	// 2. 否则，检查缓存是否有效（使用 helper.IsCacheValid）
	// 3. 如果缓存有效，读取缓存（使用 helper.ReadCache）
	// 4. 如果缓存无效或读取失败，重新获取
	// 5. 获取远程配置（使用 helper.FetchYAMLWithRetry）
	// 6. 解析 YAML 内容（使用 helper.ParseYAML）
	// 7. 将结果写入缓存（使用 helper.WriteCache）
	// 8. 记录更新时间（可选）
	// 注意: 适当使用 context 处理取消信号

	// 如果收到了取消信号，那么就直接返回
	testContext := func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	var (
		yaml any
		err  error
	)
	cacheDir := helper.GetCacheDir(baseDir)

	if force {
		if err := testContext(); err != nil {
			return nil, err
		}
		yaml, err = helper.FetchYAMLFromURL(url, time.Duration(timeout), userAgent)
	} else {
		if err := testContext(); err != nil {
			return nil, err
		}
		// 获取缓存出错我们不认为是致命的
		yaml, err = helper.ReadCache(cacheDir, url)

		if err != nil {
			defer helper.ClearCache(cacheDir, url)

			// 跳过
			err = errors.MarkSkippable(err)
			fmt.Println(err)

			// 重新获取，此处再出错那就没办法了，必须终止
			if err := testContext(); err != nil {
				return nil, err
			}
			yaml, err = helper.FetchYAMLFromURL(url, time.Duration(timeout), userAgent)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "在获取订阅时")
	}

	// 写入缓存的时候如果出错了无所谓，通知一下用户
	if err := testContext(); err != nil {
		return nil, err
	}
	if err := helper.WriteCache(cacheDir, &yaml, url); err != nil {
		defer helper.ClearCache(cacheDir, url)

		err = errors.MarkSkippable(err)
		fmt.Println(err)
	}

	return yaml, nil
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
	// TODO: 实现本地配置读取
	// 1. 构建服务器配置文件路径（如 server.yaml）
	// 2. 检查文件是否存在（使用 helper.FileExists）
	// 3. 如果文件不存在，返回可跳过错误（errors.ErrConfigNotFound）
	// 4. 读取 YAML 文件（使用 helper.ReadYAMLFile）
	// 5. 处理可能的读取错误

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		yaml, err := helper.ReadYAMLFile(baseDir + "mihomo-server.yaml")
		if err != nil {
			return nil, errors.Wrap(err, "在读取本地配置文件时")
		} else {
			return yaml, nil
		}
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
	// TODO: 实现配置文件写入
	// 1. 构建配置文件路径（如 config.yaml）
	// 2. 确保目录存在（使用 helper.EnsureDir）
	// 3. 写入 YAML 文件（使用 helper.WriteYAMLFile）
	// 4. 处理写入错误

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		configPath := filepath.Join(baseDir, "config.yaml")

		if err := helper.WriteYAMLFile(configPath, data); err != nil {
			return errors.Wrap(err, "在写入最终配置文件时")
		} else {
			return nil
		}
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
