// http.go - HTTP 客户端辅助函数
//
// 本文件提供了 HTTP 请求相关的辅助函数，用于从远程 URL 获取 YAML 配置。
// 在 Go 中，HTTP 客户端通过 `net/http` 标准库提供，这是一个功能完整且广泛使用的库。
//
// Go 的 HTTP 客户端特点：
// 1. 简单直接：默认客户端适用于大多数场景
// 2. 高度可配置：支持超时、代理、TLS 配置等
// 3. 并发安全：http.Client 可以在多个 goroutine 中安全使用
// 4. 自动连接池：默认启用持久连接（HTTP keep-alive）
//
// 与 Rust 对比：
// - Rust 使用 reqwest 库，基于 tokio 异步运行时
// - Go 的 net/http 是同步 API，但 goroutine 使其天然适合并发
// - 两者都支持超时、重定向、cookie 管理等常见功能
//
// 术语表（中英对照）：
// - HTTP client: HTTP 客户端，用于发送 HTTP 请求
// - request: 请求，包含 URL、方法、头部等信息
// - response: 响应，包含状态码、头部、响应体
// - timeout: 超时，请求的最长等待时间
// - User-Agent: 用户代理，标识客户端应用程序的字符串
// - status code: 状态码，表示请求结果的数字代码
// - redirect: 重定向，服务器指示客户端访问另一个 URL
// - keep-alive: 持久连接，复用 TCP 连接以提高性能
//
// 本文件函数设计原则：
// 1. 提供简单易用的高级 API
// 2. 支持常见的 HTTP 客户端配置
// 3. 提供清晰的错误信息和上下文
// 4. 遵循 Go 的惯用错误处理模式

package helper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sylvan-lyon/mihomo-update/internal/errors"
)

// DefaultTimeout 是 HTTP 请求的默认超时时间（秒）
// 在 Go 中，time.Duration 类型表示时间段，单位纳秒
// 使用 time.Second * 60 表示 60 秒，比直接使用数字更清晰
const DefaultTimeout = 60 * time.Second

// DefaultUserAgent 是默认的 User-Agent 头
// 遵循 RFC 7231 规范，格式为 "产品名/版本"
const DefaultUserAgent = "clash-verge/v2.4.6"

// FetchYAMLFromURL 从指定 URL 获取 YAML 内容
//
// 功能：向远程服务器发送 HTTP GET 请求，获取 YAML 配置数据。
// 参数：
//   - url: 目标 URL
//   - timeout: 请求超时时间（秒），0 表示使用默认值
//   - userAgent: User-Agent 头，空字符串表示使用默认值
//
// 返回值：
//   - []byte: 响应体内容
//   - error: 请求失败时返回错误
//
// 实现步骤：
// 1. 创建配置好的 HTTP 客户端
// 2. 构建 HTTP 请求，设置 User-Agent 头
// 3. 发送请求并检查响应状态码
// 4. 读取响应体内容
// 5. 关闭响应体（重要！避免资源泄漏）
//
// 注意：
// - 必须始终关闭响应体，即使不读取其内容
// - 应该检查状态码，非 2xx 状态码视为错误
// - 超时包括连接建立、请求发送和响应读取的全部时间
//
// 可能返回的错误：
//   - 网络错误：DNS 解析失败、连接拒绝、连接超时等
//   - HTTP 错误：状态码 4xx 或 5xx
//   - IO 错误：读取响应体失败
//   - 超时错误：在指定时间内未完成请求
func FetchYAMLFromURL(url string, timeout time.Duration, userAgent string) ([]byte, error) {
	// 实现步骤：
	// 1. 如果 timeout 为 0，使用 DefaultTimeout
	// 2. 如果 userAgent 为空，使用 DefaultUserAgent
	// 3. 使用 http.NewRequestWithContext 创建请求
	// 4. 使用 createHTTPClient 创建配置好的客户端
	// 5. 检查响应状态码，非 2xx 返回错误
	// 6. 使用 io.ReadAll 读取响应体
	// 7. 确保使用 defer resp.Body.Close() 关闭响应体
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	client := createHTTPClient(timeout)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "创建请求失败")
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "发送请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.HTTPError("HTTP 响应码失败", resp, nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "响应体无效")
	}

	return body, nil
}

// createHTTPClient 创建配置好的 HTTP 客户端
//
// 功能：根据配置创建 http.Client 实例。
// 参数：
//   - timeout: 超时时间
//
// 返回值：
//   - *http.Client: 配置好的 HTTP 客户端
//
// 实现原理：
// 设置客户端的超时参数，包括：
// 1. Timeout: 整个请求的超时时间（从拨号到读取响应体）
// 2. 注意：Go 1.13+ 推荐使用 Timeout 字段而不是分开设置
//
// 超时策略：
// - 连接超时：建立 TCP 连接的最长时间
// - TLS 握手超时：TLS 握手的最长时间
// - 请求超时：从发送请求到接收响应头的最长时间
// - 响应体读取超时：读取响应体的最长时间
//
// 注意：http.Client 的 Timeout 字段涵盖了所有阶段，
// 比分别设置更简单且不易出错。
func createHTTPClient(timeout time.Duration) *http.Client {
	// 创建配置好的 HTTP 客户端
	// 使用 http.Client 的 Timeout 字段，它涵盖了连接、TLS握手、请求和响应的总超时时间
	return &http.Client{
		Timeout: timeout,
	}
}

// FetchYAMLWithRetry 带重试机制的 YAML 获取
//
// 功能：尝试从 URL 获取 YAML，失败时重试指定次数。
// 参数：
//   - url: 目标 URL
//   - timeout: 单次请求超时时间
//   - userAgent: User-Agent 头
//   - maxRetries: 最大重试次数（0 表示不重试）
//   - retryDelay: 重试延迟时间
//
// 返回值：
//   - []byte: 响应体内容
//   - error: 所有重试都失败时返回错误
//
// 实现原理：
// 实现简单的指数退避重试策略：
// 1. 首次请求立即执行
// 2. 每次重试前等待 retryDelay * 2^n 时间
// 3. 最大重试次数限制避免无限重试
//
// 注意：仅对临时性错误重试（如网络超时、5xx 状态码）。
// 对客户端错误（4xx）不应重试。
func FetchYAMLWithRetry(url string, timeout time.Duration, userAgent string, maxRetries int, retryDelay time.Duration) ([]byte, error) {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			delay := retryDelay * time.Duration(1<<uint(i-1)) // 指数退避
			time.Sleep(delay)
		}

		body, err := FetchYAMLFromURL(url, timeout, userAgent)
		if err == nil {
			return body, nil
		}

		lastErr = err
		// 检查是否为可重试错误
		if !shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("重试 %d 次后失败: %v", maxRetries, lastErr)
}

// shouldRetry 判断错误是否应该重试
//
// 功能：根据错误类型判断请求是否应该重试。
// 参数：
//   - err: 错误对象
//
// 返回值：
//   - bool: true 表示应该重试，false 表示不应重试
//
// 可重试的错误类型：
// 1. 网络超时错误
// 2. 临时性网络错误（如连接被重置）
// 3. 服务器错误（5xx 状态码）
// 4. 特定情况下可安全重试的错误
//
// 不应重试的错误类型：
// 1. 客户端错误（4xx 状态码，如 404、403）
// 2. 请求格式错误
// 3. 认证失败
// 4. 非临时性错误
func shouldRetry(err error) bool {
	// 检查 HTTP 状态码错误
	var httpStatusError *errors.HTTPStatusError
	if errors.As(err, &httpStatusError) {
		// 5xx 服务器错误可重试，4xx 客户端错误不重试
		return httpStatusError.StatusCode >= 500 && httpStatusError.StatusCode < 600
	}

	// 检查网络超时错误
	if isTimeoutError(err) {
		return true
	}

	// 检查 url.Error（网络层错误）
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// Timeout() 表示操作超时
		// Temporary() 是历史遗留方法，表示临时错误（可能可重试）
		// 注意：Temporary() 方法已不推荐使用
		if urlErr.Timeout() || urlErr.Temporary() {
			return true
		}
		// 检查底层是否为 context.DeadlineExceeded
		if errors.Is(urlErr, context.DeadlineExceeded) {
			return true
		}
	}

	// 其他错误不重试
	return false
}

// isTimeoutError 检查是否为超时错误
//
// 功能：判断错误是否为网络超时错误。
// 参数：
//   - err: 错误对象
//
// 返回值：
//   - bool: true 表示超时错误，false 表示其他错误
//
// 实现原理：
// 检查错误链中是否包含超时错误。
// 网络超时错误通常由 context.DeadlineExceeded 或 os.IsTimeout 指示。
func isTimeoutError(err error) bool {
	// 检查 context.DeadlineExceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// 检查实现了 Timeout() bool 方法的错误
	// os.IsTimeout 会检查错误或解包后的错误是否实现了 Timeout() 且返回 true
	// 注意：os.IsTimeout 在 Go 1.13+ 可用
	return os.IsTimeout(err)
}

// ExampleUsage 展示 HTTP 客户端的使用示例
//
// 功能：演示如何使用本文件提供的函数。
// 这是一个示例函数，不会在正式代码中调用。
func ExampleUsage() {
	// 基本用法
	body, err := FetchYAMLFromURL("https://example.com/config.yaml", 30*time.Second, "my-app/1.0")
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}
	fmt.Printf("获取到 %d 字节数据\n", len(body))

	// 带重试的用法
	body2, err := FetchYAMLWithRetry("https://example.com/config.yaml", 30*time.Second, "my-app/1.0", 3, 1*time.Second)
	if err != nil {
		fmt.Printf("带重试的获取也失败了: %v\n", err)
		return
	}
	fmt.Printf("带重试获取到 %d 字节数据\n", len(body2))

	// 自定义客户端
	client := createHTTPClient(60 * time.Second)
	fmt.Printf("客户端超时: %v\n", client.Timeout)
}

// 常见问题与解决方案：
//
// 1. 问题：响应体未关闭导致文件描述符泄漏
//    解决：始终使用 defer resp.Body.Close()
//
// 2. 问题：大响应体导致内存溢出
//    解决：使用 io.Copy 流式处理，或设置响应体大小限制
//
// 3. 问题：连接池耗尽导致性能下降
//    解决：合理设置 MaxIdleConns 和 IdleConnTimeout
//
// 4. 问题：DNS 缓存导致域名解析不及时
//    解决：自定义 DialContext 或使用自定义 Resolver
//
// 5. 问题：TLS 证书验证失败
//    解决：自定义 TLS 配置或使用 InsecureSkipVerify（仅测试环境）
//
// 性能优化建议：
// 1. 复用 http.Client 实例（全局变量或依赖注入）
// 2. 启用连接池（默认已启用）
// 3. 使用连接复用（HTTP/1.1 keep-alive 或 HTTP/2）
// 4. 合理设置超时，避免长时间等待
// 5. 考虑使用 http.Transport 的高级配置
//
// 安全注意事项：
// 1. 验证服务器证书，避免中间人攻击
// 2. 对用户提供的 URL 进行安全校验
// 3. 设置合理的请求大小限制
// 4. 避免将敏感信息记录在日志中
// 5. 使用 HTTPS 而非 HTTP（重要！）
