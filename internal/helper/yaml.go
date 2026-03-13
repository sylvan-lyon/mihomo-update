// yaml.go - YAML 处理辅助函数
//
// 本文件提供了 YAML 数据的序列化和反序列化函数。
// 使用 gopkg.in/yaml.v3 库，这是 Go 中最流行的 YAML 处理库之一。
//
// YAML 是一种人类可读的数据序列化格式，常用于配置文件。
// 与 JSON 相比，YAML 支持注释、多行字符串和更简洁的语法。
//
// Go 的 YAML 处理特点：
// 1. 基于反射：通过结构体标签 (yaml:"name") 控制字段映射
// 2. 松绑定：可以处理未知结构的 YAML（使用 interface{}）
// 3. 保持格式：v3 版本能较好保持 YAML 的格式和样式
//
// 与 Rust 对比：
// - Rust 使用 serde_yaml 库，基于 serde 框架
// - Go 使用反射，Rust 使用派生宏 (derive)
// - 两者都支持结构体映射和通用数据结构
//
// 术语表（中英对照）：
// - marshal: 序列化，将 Go 数据结构转换为 YAML 字节
// - unmarshal: 反序列化，将 YAML 字节解析为 Go 数据结构
// - reflection: 反射，运行时检查类型信息的能力
// - struct tag: 结构体标签，附加在结构体字段上的元数据
// - interface{}: 空接口，可以表示任何 Go 类型
// - yaml.Node: YAML 节点，用于低级 YAML 操作
//
// 本文件函数设计原则：
// 1. 提供高层简单 API 和底层灵活 API
// 2. 支持严格类型检查和松类型处理
// 3. 提供详细的错误信息

package helper

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	"github.com/sylvan-lyon/mihomo-update/internal/args"
)

// ReadYAMLFile 读取并解析 YAML 文件
//
// 功能：读取指定路径的 YAML 文件，解析为 Go 数据结构。
// 参数：
//   - path: 文件路径
//
// 返回值：
//   - interface{}: 解析后的数据（通常是 map[interface{}]interface{}）
//   - error: 读取或解析失败时返回错误
//
// 实现步骤：
// 1. 读取文件内容
// 2. 使用 yaml.Unmarshal 解析
// 3. 返回解析结果
//
// 注意：返回的是通用接口，适合处理未知结构的配置数据。
// 如果需要类型安全，请使用 ReadYAMLFileStrict。
func ReadYAMLFile(path string) (interface{}, error) {
	// TODO: 实现 YAML 文件读取
	// 提示：
	// 1. 使用 os.ReadFile 读取文件内容
	// 2. 使用 yaml.Unmarshal 解析内容
	// 3. 返回解析结果

	// 示例代码（取消注释并修改）：
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件 %s 失败: %w", path, err)
	}

	var data interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("解析 YAML 文件 %s 失败: %w", path, err)
	}

	return data, nil

	// return nil, fmt.Errorf("未实现: ReadYAMLFile")
}

// ReadYAMLFileStrict 严格读取 YAML 文件
//
// 功能：读取 YAML 文件并解析到指定的结构体。
// 参数：
//   - path: 文件路径
//   - out: 目标结构体的指针
//
// 返回值：
//   - error: 读取或解析失败时返回错误
//
// 实现原理：
// 使用 yaml.Unmarshal 将 YAML 解析到指定结构体。
// 如果 YAML 中有结构体中不存在的字段，会返回错误。
//
// 注意：结构体字段需要添加 yaml 标签，如 `yaml:"field_name"`。
func ReadYAMLFileStrict(path string, out interface{}) error {
	// TODO: 实现严格模式 YAML 读取
	// 提示：
	// 1. 读取文件内容
	// 2. 使用 yaml.Unmarshal 解析到 out 参数

	// 示例代码（取消注释并修改）：
	// content, err := os.ReadFile(path)
	// if err != nil {
	//     return fmt.Errorf("读取文件 %s 失败: %w", path, err)
	// }
	//
	// if err := yaml.Unmarshal(content, out); err != nil {
	//     return fmt.Errorf("解析 YAML 文件 %s 失败: %w", path, err)
	// }
	//
	// return nil

	return fmt.Errorf("未实现: ReadYAMLFileStrict")
}

// WriteYAMLFile 将数据写入 YAML 文件
//
// 功能：将 Go 数据结构序列化为 YAML 并写入文件。
// 参数：
//   - path: 文件路径
//   - data: 要写入的数据
//
// 返回值：
//   - error: 写入或序列化失败时返回错误
//
// 实现步骤：
// 1. 使用 yaml.Marshal 序列化数据
// 2. 使用 os.WriteFile 写入文件
//
// 注意：写入前会创建必要的目录结构。
func WriteYAMLFile(path string, data interface{}) error {
	// TODO: 实现 YAML 文件写入
	// 提示：
	// 1. 使用 yaml.Marshal 序列化数据
	// 2. 确保目录存在（使用 EnsureDir）
	// 3. 使用 os.WriteFile 写入文件

	// 示例代码（取消注释并修改）：
	content, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("写入文件 %s 失败: %w", path, err)
	}

	return nil

	// return fmt.Errorf("未实现: WriteYAMLFile")
}

// MergeYAML 合并两个 YAML 数据结构
//
// 功能：根据指定策略合并两个 YAML 数据。
// 参数：
//   - local: 本地 YAML 数据
//   - remote: 远程 YAML 数据
//   - strategy: 合并策略（Keep, KeepAll, Force）
//
// 返回值：
//   - interface{}: 合并后的数据
//   - error: 合并失败时返回错误
//
// 实现原理：
// 根据策略递归合并两个数据结构。
// - Keep: 保留本地标量值，用远程替换序列
// - KeepAll: 保留本地标量值，将远程序列追加到本地
// - Force: 用远程值覆盖本地值
//
// 注意：此函数将在阶段 5 详细实现，这里先提供框架。
func MergeYAML(local, remote interface{}, strategy args.MergeStrategy) (interface{}, error) {
	// TODO: 实现 YAML 合并（阶段 5）
	// 提示：此函数将在阶段 5 实现
	return nil, fmt.Errorf("未实现: MergeYAML")
}

// ParseYAML 解析 YAML 字符串
//
// 功能：将 YAML 字符串解析为 Go 数据结构。
// 参数：
//   - content: YAML 字符串
//
// 返回值：
//   - interface{}: 解析后的数据
//   - error: 解析失败时返回错误
//
// 注意：这是 ReadYAMLFile 的内存版本，用于测试和调试。
func ParseYAML(content string) (interface{}, error) {
	// TODO: 实现 YAML 字符串解析
	// 提示：使用 yaml.Unmarshal

	// 示例代码（取消注释并修改）：
	var data interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("解析 YAML 字符串失败: %w", err)
	}
	return data, nil

	// return nil, fmt.Errorf("未实现: ParseYAML")
}

// ToYAML 将数据转换为 YAML 字符串
//
// 功能：将 Go 数据结构序列化为 YAML 字符串。
// 参数：
//   - data: 要序列化的数据
//
// 返回值：
//   - string: YAML 字符串
//   - error: 序列化失败时返回错误
//
// 注意：这是 WriteYAMLFile 的内存版本，用于测试和调试。
func ToYAML(data interface{}) (string, error) {
	// TODO: 实现数据到 YAML 字符串的转换
	// 提示：使用 yaml.Marshal

	// 示例代码（取消注释并修改）：
	content, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %w", err)
	}
	return string(content), nil

	// return "", fmt.Errorf("未实现: ToYAML")
}
