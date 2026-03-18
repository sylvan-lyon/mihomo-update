// cache.go - 缓存与时间处理
//
// 本文件实现了订阅配置的缓存机制，避免频繁下载相同配置。
// 核心学习目标：time 包、文件系统状态检查和 JSON 序列化。
//
// 缓存设计要点：
// 1. 24小时过期时间：使用 time.Duration 表示时间间隔
// 2. 文件修改时间检查：使用 os.Stat 获取文件状态
// 3. JSON 序列化：使用 encoding/json 存储结构化缓存数据
// 4. 缓存键管理：基于 URL 生成唯一的缓存文件名
//
// Go 时间处理特点：
// 1. time.Time: 表示时间点的结构体，支持时区
// 2. time.Duration: 表示时间间隔，如 24*time.Hour
// 3. time.Now(): 获取当前时间
// 4. time.Since(t): 计算从 t 到现在的时间间隔
//
// 文件状态检查：
// 1. os.Stat(path) 返回 FileInfo 接口
// 2. FileInfo.ModTime() 返回文件最后修改时间
// 3. os.IsNotExist(err) 检查文件是否存在
//
// JSON 序列化：
// 1. encoding/json 是 Go 标准库的 JSON 支持
// 2. json.Marshal(v) 将 Go 值序列化为 JSON 字节
// 3. json.Unmarshal(data, &v) 将 JSON 字节反序列化为 Go 值
// 4. 结构体字段标签 `json:"field_name"` 控制序列化
//
// 与 Rust 对比：
// - Rust: 使用 std::time::SystemTime 和 chrono 库处理时间
// - Rust: 使用 serde_json 进行 JSON 序列化
// - Go: 时间处理更简单直接，JSON 序列化基于反射
// - Go: 错误处理模式不同（多返回值 vs Result 类型）
//
// 术语表（中英对照）:
// - cache: 缓存，临时存储以加速后续访问
// - time.Time: 时间点，Go 的时间表示类型
// - time.Duration: 时间间隔，如 24*time.Hour
// - os.Stat: 获取文件状态（stat 系统调用）
// - FileInfo: 文件信息接口
// - ModTime: 修改时间（modification time）
// - JSON: JSON 数据格式，保留英文
// - marshal: 序列化，将数据结构转换为字节流
// - unmarshal: 反序列化，将字节流转换为数据结构
// - encoding/json: Go 的 JSON 编码/解码包
// - struct tag: 结构体标签，元数据注解
// - expiration: 过期，缓存失效的时间点
// - TTL: 生存时间（Time To Live），缓存有效时长
//
// 本文件设计原则：
// 1. 简单的缓存机制，基于文件系统和 JSON
// 2. 清晰的错误处理和资源管理
// 3. 避免竞态条件（单线程使用，无需锁）
// 4. 提供合理的默认值（24小时过期）

package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sylvan-lyon/mihomo-update/internal/errors"
)

// CacheEntry 表示一个缓存条目
//
// 结构体字段使用 JSON 标签控制序列化：
// 1. Data: 缓存的数据内容（通常是 YAML 配置）
// 2. CreatedAt: 缓存创建时间（RFC3339 格式）
// 3. URL: 缓存的订阅 URL（用于验证）
//
// 注意：结构体字段需要导出（首字母大写）才能被 JSON 包访问。
type CacheEntry struct {
	// Data 是缓存的配置数据
	// 类型为 any（interface{}），可以存储任意 YAML 数据
	// JSON 标签：data
	Data any `json:"data"`

	// CreatedAt 是缓存创建时间
	// 使用 RFC3339 格式字符串存储，便于人类阅读和跨平台
	// JSON 标签：created_at
	CreatedAt string `json:"created_at"`

	// URL 是缓存的订阅地址
	// 用于验证缓存是否对应正确的 URL
	// JSON 标签：url
	URL string `json:"url"`
}

// GetCacheAge 获取缓存年龄（用于调试和日志）
//
// 功能：计算缓存已存在的时间。
// 参数：
//   - cachePath: 缓存文件路径
//
// 返回值：
//   - time.Duration: 缓存年龄，如果缓存不存在返回 0
//   - error: 时间解析失败时返回错误
//
// 注意：此函数主要用于调试和日志记录。
func (entry *CacheEntry) Age() (time.Duration, error) {
	lastUpdatedAt, err := time.Parse(time.RFC3339, entry.CreatedAt)
	if err != nil {
		return 0, errors.Wrap(err, "在试图解析缓存的 created_at 时")
	} else {
		return time.Since(lastUpdatedAt), nil
	}
}

// IsValid 检查缓存是否有效（存在且未过期）
// 功能：检查缓存文件是否存在，且创建时间在 TTL 内。
// 算法：
// 1. 检查缓存文件是否存在
// 2. 读取缓存条目的 CreatedAt 时间
// 3. 比较当前时间与创建时间的时间差
// 4. 判断是否超过 cacheTTL
//
// 参数：
//   - cacheDir: 缓存文件目录
//   - url: 预期的订阅 URL（用于验证缓存对应正确的 URL）
//
// 返回值：
//   - bool: 缓存有效返回 true，否则返回 false
//   - error: JSON 解析失败时返回错误
//
// 注意：即使缓存无效，函数也可能返回 false 和 nil 错误。
func (entry *CacheEntry) IsValid() (bool, error) {
	age, err := entry.Age()
	if err != nil {
		return false, errors.Wrap(err, "在试图解析缓存的 created_at 时")
	} else {
		return age < cacheTTL, nil
	}
}

// CacheTTL 定义缓存的生存时间（Time To Live）
//
// 使用 time.Duration 类型表示时间间隔。
// 常量命名使用驼峰式，虽然通常常量使用大写，但包内私有常量可以使用小写。
const cacheTTL = 24 * time.Hour

func GetCacheDir(baseDir string) string {
	return baseDir + "mihomo-update"
}

// GetCachePath 根据 URL 生成缓存文件路径，你可能永远用不到这个函数
//
// 功能：将 URL 转换为安全的文件名，存储在缓存目录中。
// 算法：使用简单的哈希或转换，确保文件名安全和唯一。
//
// 参数：
//   - cacheDir: 缓存目录路径
//   - url: 订阅 URL
//
// 返回值：
//   - string: 完整的缓存文件路径
//
// 实现提示：
// 1. 确保缓存目录存在（os.MkdirAll）
// 2. 将 URL 转换为安全文件名（避免特殊字符）
// 3. 添加 .json 扩展名表明文件格式
func GetCachePath(cacheDir, url string) string {
	digested := sha256.Sum256([]byte(url))
	cacheFile := filepath.Join(cacheDir, hex.EncodeToString(digested[:])) + ".json"

	return cacheFile
}

// ReadCache 读取缓存文件
//
// 功能：从缓存文件读取 CacheEntry 结构体。
// 算法：
// 1. 打开缓存文件（os.Open）
// 2. 使用 json.Decoder 或 json.Unmarshal 解析 JSON
// 3. 返回 CacheEntry 结构体
//
// 参数：
//   - cacheDir: 缓存路径
//   - url: 订阅 url
//
// 返回值：
//   - *CacheEntry: 缓存条目指针，nil 表示缓存不存在或无效
//   - error: 文件操作或 JSON 解析失败时返回错误
//
// 注意：调用者应检查返回的 CacheEntry 是否为 nil。
func ReadCache(cacheDir, url string) (*CacheEntry, error) {
	cacheFile := GetCachePath(cacheDir, url)

	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, errors.Wrap(err, "在读取缓存文件时")
	}
	defer file.Close()

	var data CacheEntry
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "在解析缓存文件时")
	}

	return &data, nil
}

// WriteCache 写入缓存文件
//
// 功能：将数据写入缓存文件，包含创建时间和 URL。
// 算法：
// 1. 创建 CacheEntry 结构体
// 2. 使用 json.MarshalIndent 生成格式化的 JSON
// 3. 写入文件（os.WriteFile 或 os.Create）
//
// 参数：
//   - cacheDir: 缓存目录路径
//   - data: 要缓存的数据（通常是 YAML 配置）
//   - url: 订阅 URL
//
// 返回值：
//   - error: 文件操作或 JSON 序列化失败时返回错误
//
// 注意：此函数会覆盖已存在的缓存文件。
func WriteCache(cacheDir string, data any, url string) error {
	cache := CacheEntry{
		Data:      data,
		CreatedAt: time.Now().Local().Format(time.RFC3339),
		URL:       url,
	}

	jsonData, err := json.MarshalIndent(&cache, "", "    ")
	if err != nil {
		return errors.Wrap(err, "创建缓存文件内容时")
	}

	cacheFile := GetCachePath(cacheDir, url)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return errors.Wrap(err, "创建缓存文件所在的目录时")
	}

	if err := os.WriteFile(cacheFile, jsonData, 0644); err != nil {
		return errors.Wrap(err, "写入缓存文件时")
	}

	return nil
}

// ClearCache 清除缓存文件
//
// 功能：删除指定的缓存文件。
// 参数：
//   - cachePath: 缓存文件路径
//
// 返回值：
//   - error: 文件删除失败时返回错误
//
// 注意：如果文件不存在，返回 nil（成功）。
func ClearCache(cacheDir, url string) error {
	cacheFile := GetCachePath(cacheDir, url)
	err := os.Remove(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return errors.Wrap(err, "在删除缓存文件时")
		}
	} else {
		return nil
	}
}

// 常见问题与解决方案：
//
// 1. 问题：time.Parse 解析时间字符串失败
//    解决：确保使用一致的 RFC3339 格式，或使用 time.ParseInLocation
//
// 2. 问题：JSON 序列化 any 类型时丢失类型信息
//    解决：YAML 数据在 JSON 中会保持基本类型（map[string]any, []any 等）
//
// 3. 问题：文件权限问题（无法创建目录或文件）
//    解决：使用 os.MkdirAll 创建目录，检查进程权限
//
// 4. 问题：竞态条件（并发读写）
//    解决：本项目单线程使用，无需锁。如需并发，使用 sync.Mutex
//
// 5. 问题：缓存文件过大
//    解决：定期清理旧缓存，或限制缓存大小
//
// 6. 问题：URL 包含特殊字符，无法作为文件名
//    解决：使用 base64 编码或哈希（如 SHA256）生成安全文件名
//
// 测试建议：
// 1. 测试缓存文件读写功能
// 2. 测试缓存过期逻辑
// 3. 测试 URL 验证功能
// 4. 测试边界情况（空数据、无效 JSON、文件权限等）
//
// 扩展学习：
// 1. 使用 BoltDB 或 Badger 实现更复杂的缓存
// 2. 添加缓存压缩（gzip）减少磁盘占用
// 3. 实现 LRU（最近最少使用）缓存淘汰策略
// 4. 添加缓存统计信息（命中率、平均加载时间等）
