// Package errors 提供应用程序的错误处理基础设施，遵循 Go 最佳实践。
// Package errors provides error handling infrastructure for the application, following Go best practices.
//
// 本模块对应 Rust 版本的 `src/errors.rs` 文件，但采用 Go 地道的错误处理模式。
// 关键学习点:
// 1. Go 的 error 接口与简单错误值
// 2. 错误包装 (error wrapping) 使用 fmt.Errorf 和 %w 动词
// 3. 自定义错误类型与错误检查 (errors.Is, errors.As)
// 4. 可跳过错误的实现模式
//
// 双语术语:
// - error interface: error 接口 - Go 内置的错误类型接口
// - error wrapping: 错误包装 - 使用 `%w` 动词封装底层错误
// - error unwrapping: 错误解包 - 使用 errors.Unwrap 获取底层错误
// - sentinel error: 哨兵错误 - 预定义的错误值，用于特定错误条件
package errors

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrSkippable 是一个哨兵错误值，用于标记可跳过的错误。
// ErrSkippable is a sentinel error value for marking skippable errors.
//
// 使用 errors.Is(err, ErrSkippable) 检查错误是否可跳过。
var ErrSkippable = errors.New("skippable error")

// 文件系统相关哨兵错误
// File system related sentinel errors
var (
	// ErrConfigNotFound 表示配置文件未找到
	// ErrConfigNotFound indicates configuration file not found
	ErrConfigNotFound = errors.New("configuration file not found")

	// ErrConfigPermission 表示配置文件权限不足
	// ErrConfigPermission indicates insufficient permissions for configuration file
	ErrConfigPermission = errors.New("insufficient file permissions")

	// ErrConfigInvalidPath 表示配置文件路径无效
	// ErrConfigInvalidPath indicates invalid configuration file path
	ErrConfigInvalidPath = errors.New("invalid configuration file path")

	// ErrConfigReadFailed 表示读取配置文件失败
	// ErrConfigReadFailed indicates failed to read configuration file
	ErrConfigReadFailed = errors.New("failed to read configuration file")

	// ErrConfigWriteFailed 表示写入配置文件失败
	// ErrConfigWriteFailed indicates failed to write configuration file
	ErrConfigWriteFailed = errors.New("failed to write configuration file")

	// ErrDirCreationFailed 表示创建目录失败
	// ErrDirCreationFailed indicates failed to create directory
	ErrDirCreationFailed = errors.New("failed to create directory")
)

// YAML 处理相关哨兵错误
// YAML processing related sentinel errors
var (
	// ErrYAMLParse 表示 YAML 解析失败
	// ErrYAMLParse indicates YAML parsing failed
	ErrYAMLParse = errors.New("YAML parsing failed")

	// ErrYAMLFormat 表示 YAML 格式无效
	// ErrYAMLFormat indicates invalid YAML format
	ErrYAMLFormat = errors.New("invalid YAML format")

	// ErrYAMLType 表示 YAML 类型不匹配
	// ErrYAMLType indicates YAML type mismatch
	ErrYAMLType = errors.New("YAML type mismatch")

	// ErrYAMLSerialize 表示 YAML 序列化失败
	// ErrYAMLSerialize indicates YAML serialization failed
	ErrYAMLSerialize = errors.New("YAML serialization failed")
)

// AppError 是应用程序的自定义错误类型，包含原始错误和上下文信息。
// AppError is the custom error type for the application, containing the original error and context.
//
// 遵循 Go 最佳实践:
// 1. 错误类型名以 "Error" 结尾
// 2. 实现 error 接口 (Error() string 方法)
// 3. 实现 Unwrap() error 方法以支持错误链
// 4. 提供构造函数 NewAppError
type AppError struct {
	// Op 描述引发错误的具体操作，如 "读取配置文件"、"发送 HTTP 请求"
	// Op describes the operation that caused the error, e.g., "read config file", "send HTTP request"
	Op string

	// Err 是底层错误，可以是任何实现了 error 接口的类型
	// Err is the underlying error, can be any type implementing the error interface
	Err error

	// 注意: Go 中不推荐在错误类型中直接包含 "skippable" 标志，
	// 而是使用 ErrSkippable 哨兵错误或单独的检查机制。
}

// Error 返回错误的字符串表示，遵循 Go 错误格式惯例。
// Error returns the string representation of the error, following Go error formatting conventions.
func (e *AppError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap 返回底层错误，支持 errors.Unwrap 和 errors.Is/As。
// Unwrap returns the underlying error, supporting errors.Unwrap and errors.Is/As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建一个新的 AppError。
// NewAppError creates a new AppError.
func NewAppError(op string, err error) *AppError {
	return &AppError{Op: op, Err: err}
}

// Wrap 是一个辅助函数，用操作上下文包装错误。
// Wrap is a helper function that wraps an error with operation context.
//
// 这是 Go 中最常见的错误上下文添加方式。
// 示例:
//
//	data, err := os.ReadFile("config.yaml")
//	if err != nil {
//		return errors.Wrap(err, "读取配置文件")
//	}
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &AppError{Op: op, Err: err}
}

// Wrapf 是一个辅助函数，用格式化的操作上下文包装错误。
// Wrapf is a helper function that wraps an error with formatted operation context.
//
// 示例:
//
//	err := errors.Wrapf(ioErr, "处理文件 %s", filename)
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &AppError{Op: fmt.Sprintf(format, args...), Err: err}
}

// MarkSkippable 将错误标记为可跳过。
// MarkSkippable marks an error as skippable.
//
// 如果错误已经是 AppError，在其错误链中添加 ErrSkippable。
// 否则，创建一个新的 AppError 包装原始错误并添加 ErrSkippable。
func MarkSkippable(err error) error {
	if err == nil {
		return nil
	}

	// 检查错误是否已经是 AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		// 创建一个新的错误链: ErrSkippable -> 原始错误
		return &AppError{
			Op:  appErr.Op,
			Err: fmt.Errorf("%w: %w", ErrSkippable, appErr.Err),
		}
	}

	// 对于非 AppError 错误，用 ErrSkippable 包装
	return &AppError{
		Op:  "",
		Err: fmt.Errorf("%w: %v", ErrSkippable, err),
	}
}

// IsSkippable 检查错误是否可跳过。
// IsSkippable checks if an error is skippable.
//
// 使用 errors.Is 检查错误链中是否包含 ErrSkippable。
func IsSkippable(err error) bool {
	return errors.Is(err, ErrSkippable)
}

// Celebrate 在操作成功时打印庆祝消息。
// Celebrate prints a celebration message when an operation succeeds.
//
// 如果 err == nil，打印成功消息。
// 返回原始错误（无论是否成功）。
func Celebrate(err error, msg string) error {
	if err == nil && msg != "" {
		fmt.Println(msg)
	}
	return err
}

// OrSkipPrint 将错误标记为可跳过并打印。
// OrSkipPrint marks an error as skippable and prints it.
//
// 如果 err != nil，将其标记为可跳过并使用 fmt.Println 打印。
// 返回标记后的错误（或 nil）。
func OrSkipPrint(err error) error {
	if err == nil {
		return nil
	}

	skippableErr := MarkSkippable(err)
	fmt.Println(skippableErr)
	return skippableErr
}

// IOError 从 io 错误创建 AppError。
// IOError creates an AppError from an io error.
func IOError(op string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(err, op)
}

// HTTPError 从 HTTP 错误创建 AppError。
// HTTPError creates an AppError from an HTTP error.
func HTTPError(op string, resp *http.Response, err error) error {
	if err != nil {
		return Wrap(err, op)
	}

	if resp != nil && resp.StatusCode >= 400 {
		return NewAppError(op, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	return nil
}

// YAMLError 从 YAML 解析错误创建 AppError。
// YAMLError creates an AppError from a YAML parsing error.
//
// 注意: 实际的 YAML 库导入将在阶段 3 添加。
func YAMLError(op string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(err, op)
}

// 以下是 Go 错误处理最佳实践示例:

// ExampleSentinelErrors 展示哨兵错误的使用。
// ExampleSentinelErrors demonstrates sentinel error usage.
func ExampleSentinelErrors() error {
	// 使用预定义的哨兵错误
	// Using predefined sentinel errors
	_ = ErrConfigNotFound
	_ = ErrConfigPermission
	_ = ErrYAMLParse
	_ = ErrYAMLSerialize

	// 使用 errors.Is 检查哨兵错误
	// Using errors.Is to check for sentinel errors
	// if errors.Is(err, ErrConfigNotFound) {
	//     // 处理配置文件未找到的情况
	//     // Handle configuration file not found case
	// }
	//
	// if errors.Is(err, ErrYAMLParse) {
	//     // 处理 YAML 解析失败
	//     // Handle YAML parsing failure
	// }

	return nil
}

// ExampleErrorInspection 展示错误检查和类型断言。
// ExampleErrorInspection demonstrates error inspection and type assertion.
func ExampleErrorInspection(err error) {
	// 使用 errors.As 检查错误类型
	var appErr *AppError
	if errors.As(err, &appErr) {
		fmt.Printf("AppError operation: %s\n", appErr.Op)
	}

	// 使用 errors.Is 检查错误值
	if errors.Is(err, ErrSkippable) {
		fmt.Println("Error is skippable")
	}

	// 错误解包
	// for e := err; e != nil; e = errors.Unwrap(e) {
	//     fmt.Printf("Error in chain: %v\n", e)
	// }
}

// ExampleErrorChaining 展示错误链的创建。
// ExampleErrorChaining demonstrates error chain creation.
func ExampleErrorChaining() error {
	// 模拟一个底层错误
	baseErr := io.ErrUnexpectedEOF

	// 使用 %w 包装错误
	wrappedErr := fmt.Errorf("read data: %w", baseErr)

	// 进一步添加上下文
	finalErr := Wrap(wrappedErr, "process file")

	return finalErr
}
