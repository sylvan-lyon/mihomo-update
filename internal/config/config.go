package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 保存所有应用程序配置。
// 它使用结构体标签进行Viper绑定和验证。
type Config struct {
	// 订阅URL，用于获取配置
	SubscriptionURL string `mapstructure:"subscription_url" validate:"required,url"`

	// 本地Mihomo配置文件的路径
	ConfigPath string `mapstructure:"config_path" validate:"required"`

	// 使用的合并策略（keep、keepall、force）
	MergeStrategy string `mapstructure:"merge_strategy" validate:"oneof=keep keepall force"`

	// HTTP超时时间（秒）
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`

	// HTTP请求的User-Agent标头
	UserAgent string `mapstructure:"user_agent"`

	// 即使缓存存在也强制更新
	ForceUpdate bool `mapstructure:"force_update"`

	// 缓存过期时长
	CacheExpiration time.Duration `mapstructure:"cache_expiration"`

	// CLI消息的语言（en、zh-CN）
	Language string `mapstructure:"language" validate:"oneof=en zh-CN"`
}

// Load 从多个源按优先级顺序加载配置：
// 1. 命令行标志（最高优先级）
// 2. 环境变量
// 3. 配置文件
// 4. 默认值（最低优先级）
func Load() (*Config, error) {
	// TODO: 使用配置文件搜索路径初始化Viper
	// viper.SetConfigName("config")
	// viper.SetConfigType("yaml")
	// viper.AddConfigPath(".")
	// viper.AddConfigPath("$HOME/.config/mihomo")
	// viper.AddConfigPath("/etc/mihomo")

	// TODO: 绑定环境变量（自动转换为大写并使用下划线）
	// viper.SetEnvPrefix("MIHOMO")
	// viper.AutomaticEnv()

	// TODO: 设置默认值
	// viper.SetDefault("merge_strategy", "keep")
	// viper.SetDefault("http_timeout", 60*time.Second)
	// viper.SetDefault("user_agent", "clash-verge/v2.4.6")
	// viper.SetDefault("cache_expiration", 24*time.Hour)
	// viper.SetDefault("language", "en")

	// TODO: 读取配置文件（如果存在）
	// if err := viper.ReadInConfig(); err != nil {
	//     if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
	//         return nil, fmt.Errorf("读取配置文件失败: %w", err)
	//     }
	// }

	// TODO: 将配置反序列化到Config结构体
	// var cfg Config
	// if err := viper.Unmarshal(&cfg); err != nil {
	//     return nil, fmt.Errorf("反序列化配置失败: %w", err)
	// }

	// TODO: 使用验证器库验证配置
	// if err := validate.Struct(cfg); err != nil {
	//     return nil, fmt.Errorf("配置验证失败: %w", err)
	// }

	return &Config{}, fmt.Errorf("未实现")
}

// 展示的最佳实践:
// 1. 使用结构体标签进行配置绑定（mapstructure、validate）
// 2. 配置优先级（标志 > 环境变量 > 文件 > 默认值）
// 3. 使用适当的Go类型实现类型安全配置（time.Duration）
// 4. 验证必填字段和枚举值
// 5. 支持多种配置文件格式（YAML、JSON等）
// 6. 支持环境变量并自动映射

// 需要添加的依赖:
// go get github.com/go-playground/validator/v10
// go get github.com/spf13/viper

// 后续步骤:
// 1. 实现Viper配置加载
// 2. 使用go-playground/validator添加验证
// 3. 添加配置文件模板支持
// 4. 添加SIGHUP信号上的配置重载
// 5. 为文档生成配置模式
