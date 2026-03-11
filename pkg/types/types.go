package types

// 此包包含可能被其他包使用的公共类型定义。
// 如果不需要导出任何类型，可以删除此文件。

// ErrorCode 表示应用程序特定的错误代码。
type ErrorCode int

const (
	// ErrConfigInvalid 表示配置验证失败。
	ErrConfigInvalid ErrorCode = iota + 1000

	// ErrNetworkFailure 表示网络请求失败。
	ErrNetworkFailure

	// ErrYAMLParse 表示YAML解析失败。
	ErrYAMLParse

	// ErrMergeConflict 表示无法解决的合并冲突。
	ErrMergeConflict

	// ErrCacheInvalid 表示缓存验证失败。
	ErrCacheInvalid
)

// String 返回错误代码的人类可读描述。
func (ec ErrorCode) String() string {
	switch ec {
	case ErrConfigInvalid:
		return "配置无效"
	case ErrNetworkFailure:
		return "网络故障"
	case ErrYAMLParse:
		return "YAML解析错误"
	case ErrMergeConflict:
		return "合并冲突"
	case ErrCacheInvalid:
		return "缓存无效"
	default:
		return "未知错误"
	}
}

// AppError 表示带有代码和上下文的应用程序错误。
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error // 包装的错误（如果有）
}

// Error 实现error接口。
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap 返回包装的错误以便进行错误链检查。
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建新的AppError。
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// MergeResult 表示合并操作的结果。
type MergeResult struct {
	// 用于合并的策略
	Strategy string

	// 合并过程中遇到的冲突（如果有）
	Conflicts []Conflict

	// 合并操作的统计信息
	Stats MergeStats
}

// Conflict 表示无法自动解决的合并冲突。
type Conflict struct {
	Path     string      // YAML路径（例如："servers[0].name"）
	OldValue interface{} // 旧值
	NewValue interface{} // 新值
	Reason   string      // 冲突原因
}

// MergeStats 包含合并操作的统计信息。
type MergeStats struct {
	TotalKeys     int // 处理的总键数
	KeysAdded     int // 从新配置添加的键数
	KeysModified  int // 修改的键数
	KeysDeleted   int // 删除的键数
	KeysUnchanged int // 未更改的键数
}

// 展示的最佳实践:
// 1. 带有错误代码枚举的自定义错误类型
// 2. 错误包装用于错误链检查
// 3. 操作结果的结构化结果类型
// 4. 使用iota生成错误代码常量
// 5. 值类型上的方法（在适当的地方不使用指针接收器）

// 何时导出类型:
// 1. 模块内多个包使用的类型
// 2. 构成公共API契约的类型
// 3. 库使用者需要理解的类型
// 4. 实现公共接口的类型

// 何时保持类型内部化:
// 1. 仅在单个包内使用的类型
// 2. 可能更改的实现细节
// 3. 不必要暴露内部状态的类型
// 4. 可能混淆库用户的类型

// 使用示例:
// func Process() error {
//     result, err := doMerge()
//     if err != nil {
//         return NewAppError(ErrMergeConflict, "合并配置失败", err)
//     }
//     logStats(result.Stats)
//     return nil
// }
