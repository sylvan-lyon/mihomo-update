// io.go - 文件操作辅助函数
//
// 本文件提供了与文件系统交互的辅助函数，包括文件存在检查、目录创建等。
// 在 Go 中，文件操作主要通过 `os` 和 `io` 包完成，与 Rust 的 `std::fs` 和 `std::io` 类似，
// 但在错误处理和资源管理上有不同的哲学。
//
// Go 文件操作特点：
// 1. 显式错误检查：每个可能失败的操作都返回 error，必须显式检查
// 2. 延迟关闭：使用 defer 确保文件描述符被正确关闭
// 3. 简单直接：API 设计简单，通常一行代码完成一个操作
//
// 与 Rust 对比：
// - Rust 使用 Result<T, E> 类型和 ? 操作符进行错误传播
// - Go 使用多返回值 (value, error) 和 if err != nil 模式
// - Rust 有所有权系统确保资源安全，Go 依赖 defer 和程序员自觉
//
// 术语表（中英对照）：
// - file descriptor: 文件描述符，操作系统对打开文件的引用
// - defer: 延迟执行，函数返回前执行的语句
// - close: 关闭，释放文件描述符等系统资源
// - read: 读取，从文件获取数据
// - write: 写入，向文件输出数据
// - path: 路径，文件在文件系统中的位置
// - permission: 权限，文件访问控制规则
//
// 本文件函数设计原则：
// 1. 每个函数完成单一职责
// 2. 使用 Go 的惯用错误处理模式
// 3. 提供清晰的错误上下文信息

package helper

import (
	"os"
	// "path/filepath" // 暂时注释，将在需要时使用

	"github.com/sylvan-lyon/mihomo-update/internal/errors"
)

// FileExists 检查文件是否存在
//
// 功能：检查指定路径的文件或目录是否存在。
// 参数：
//   - path: 文件路径
//
// 返回值：
//   - bool: 文件存在返回 true，否则返回 false
//
// 实现原理：
// 使用 os.Stat 获取文件信息，如果返回错误且错误为 os.ErrNotExist，
// 则文件不存在。其他错误（如权限错误）也视为文件不存在。
//
// 注意：此函数不区分文件和目录，两者都视为"存在"。
// 如果需要区分，请使用 IsFile 或 IsDir 函数。
func FileExists(path string) bool {
	// TODO: 实现文件存在检查
	// 提示：
	// 1. 使用 os.Stat 获取文件信息
	// 2. 检查返回的错误是否为 os.ErrNotExist
	// 3. 返回适当的布尔值

	// 示例代码（取消注释并修改）：
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// 其他错误（如权限错误）也视为文件不存在
		return false
	}
	return true

	// return false // 待实现
}

// EnsureDir 确保目录存在
//
// 功能：检查目录是否存在，如果不存在则创建它（包括所有父目录）。
// 参数：
//   - path: 目录路径
//
// 返回值：
//   - error: 创建目录失败时返回错误
//
// 实现原理：
// 使用 os.MkdirAll 创建目录，该函数会创建所有必要的父目录。
// 权限默认为 0755（用户可读写执行，组和其他用户可读执行）。
//
// 注意：如果目录已存在，os.MkdirAll 不会返回错误。
//
// 可能返回的错误：
//   - os.ErrPermission: 权限不足，无法创建目录
//   - os.ErrInvalid: 路径无效
//   - ErrDirCreationFailed: 创建目录失败（包装底层错误）
func EnsureDir(path string) error {
	// TODO: 确保目录存在
	// 提示：使用 os.MkdirAll 创建目录

	// 示例代码（取消注释并修改）：
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return errors.Wrapf(err, "创建目录 %s 失败", path)
	}
	return nil

	// return nil // 待实现
}

// ReadFile 读取文件内容
//
// 功能：读取指定文件的全部内容。
// 参数：
//   - path: 文件路径
//
// 返回值：
//   - []byte: 文件内容
//   - error: 读取失败时返回错误
//
// 实现原理：
// 使用 os.ReadFile 函数，该函数一次性读取整个文件。
// 对于大文件，应考虑使用带缓冲的逐行读取。
//
// 注意：此函数会读取整个文件到内存，不适合大文件。
//
// 可能返回的错误：
//   - os.ErrNotExist: 文件不存在
//   - os.ErrPermission: 权限不足，无法读取文件
//   - os.ErrInvalid: 路径无效
//   - ErrConfigReadFailed: 读取配置文件失败（包装底层错误）
func ReadFile(path string) ([]byte, error) {
	// TODO: 读取文件内容
	// 提示：使用 os.ReadFile

	// 示例代码（取消注释并修改）：
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "读取文件 %s 失败", path)
	}
	return data, nil

	// return nil, nil // 待实现
}

// WriteFile 写入文件内容
//
// 功能：将数据写入指定文件，如果文件不存在则创建。
// 参数：
//   - path: 文件路径
//   - data: 要写入的数据
//
// 返回值：
//   - error: 写入失败时返回错误
//
// 实现原理：
// 使用 os.WriteFile 函数，该函数会创建文件（如果不存在）并写入数据。
// 权限默认为 0644（用户可读写，组和其他用户可读）。
//
// 注意：此函数会覆盖已存在的文件。
//
// 可能返回的错误：
//   - os.ErrPermission: 权限不足，无法写入文件
//   - os.ErrInvalid: 路径无效
//   - os.ErrNoSpace: 磁盘空间不足
//   - ErrConfigWriteFailed: 写入配置文件失败（包装底层错误）
func WriteFile(path string, data []byte) error {
	// TODO: 写入文件内容
	// 提示：使用 os.WriteFile

	// 示例代码（取消注释并修改）：
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "写入文件 %s 失败", path)
	}
	return nil

	// return nil // 待实现
}
