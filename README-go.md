# mihomo-update Go版本

这是mihomo订阅更新器CLI工具的Go实现，专为学习Go最佳实践而创建。

## 项目结构

```
cmd/mihomo-update/      # 主CLI入口点
internal/               # 私有应用程序代码
├── config/             # 配置加载和验证
├── http/               # 用于获取订阅的HTTP客户端
├── merge/              # YAML合并策略
└── cli/                # CLI命令解析和执行
pkg/types/              # 公共类型定义（如果需要）
configs/                # 配置文件示例
```

## 展示的关键Go概念

- **错误处理**: 使用`error`接口，通过`fmt.Errorf`进行错误包装，自定义错误类型
- **结构体组合**: 通过结构体嵌入实现代码复用
- **接口**: 为HTTP客户端、YAML合并器等定义契约
- **并发**: 使用goroutine和通道进行并行操作
- **测试**: 单元测试、表驱动测试、集成测试
- **包组织**: 内部包与公共包，依赖注入
- **CLI开发**: 使用Cobra构建命令行界面
- **配置管理**: 使用Viper进行配置管理，支持YAML

## 入门指南

```bash
# 安装依赖
go mod tidy

# 构建二进制文件
go build -o mihomo-update cmd/mihomo-update/main.go

# 运行测试
go test ./...

# 使用示例配置运行
./mihomo-update --config configs/config.example.yaml
```

## 学习资源

- [Effective Go](https://golang.org/doc/effective_go)（Go高效编程）
- [Go by Example](https://gobyexample.com/)（Go示例教程）
- [Standard Library Documentation](https://pkg.go.dev/std)（标准库文档）
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)（Go代码审查注释）