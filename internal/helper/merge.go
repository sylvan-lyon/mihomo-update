// merge.go - YAML 合并策略算法
//
// 本文件实现了三种 YAML 数据合并策略，用于将远程配置与本地配置合并。
// 这是阶段 5 的核心内容，重点学习递归算法、类型断言和 Go 的动态类型处理。
//
// 三种合并策略（对应 args.MergeStrategy）:
// 1. Keep:    保留本地标量值，用远程替换序列（数组）
// 2. KeepAll: 保留本地标量值，将远程序列追加到本地
// 3. Force:   用远程值覆盖本地值（映射递归合并）
//
// Go 动态类型处理特点:
// 1. `interface{}` (或 `any`): 可以表示任何 Go 类型，但需要类型断言才能使用
// 2. 类型断言: `value.(type)` 在 switch 中使用，或 `value.(具体类型)` 获取具体类型
// 3. 递归处理: 需要根据值的实际类型（映射、序列、标量）决定如何合并
//
// 与 Rust 对比:
// - Rust 使用 serde_yml::Value 枚举类型，通过模式匹配处理不同变体
// - Go 使用 `interface{}` + 类型断言，更灵活但类型安全性稍弱
// - 两者都使用递归算法处理嵌套结构
//
// 术语表（中英对照）:
// - recursion: 递归，函数调用自身的编程技巧
// - type assertion: 类型断言，检查接口值的实际类型
// - mapping: 映射，键值对集合（对应 YAML 对象/字典）
// - sequence: 序列，有序元素集合（对应 YAML 数组/列表）
// - scalar: 标量，基本值（字符串、数字、布尔值、null）
// - interface{}: 空接口，可以保存任何类型的值
// - any: `interface{}` 的类型别名（Go 1.18+）
// - deep copy: 深拷贝，创建数据的完全独立副本
//
// 本文件设计原则:
// 1. 每个策略一个独立函数，职责单一
// 2. 递归算法清晰易懂
// 3. 正确处理边界情况（nil、类型不匹配等）
// 4. 避免修改输入参数，返回新值

package helper

import (
	"maps"
	"fmt"
	"reflect"

	// "reflect" // 将在实现 typeSwitch 函数时使用

	"github.com/sylvan-lyon/mihomo-update/internal/args"
	"github.com/sylvan-lyon/mihomo-update/internal/errors"
	// "github.com/sylvan-lyon/mihomo-update/internal/errors" // 将在实现错误处理时使用
)

// mergeYAML 合并两个 YAML 数据结构
//
// 功能：根据指定策略合并两个 YAML 数据，这是包内的实现函数。
// 参数：
//   - local: 本地 YAML 数据
//   - remote: 远程 YAML 数据
//   - strategy: 合并策略（Keep, KeepAll, Force）
//
// 返回值：
//   - any: 合并后的数据
//   - error: 合并失败时返回错误
//
// 实现原理：
// 根据策略调用相应的合并函数，进行递归合并。
// 注意：此函数会创建新数据，不会修改输入参数。
//
// 可能返回的错误：
//   - ErrYAMLType: YAML 类型不匹配（如尝试合并映射和序列）
//   - 递归过程中可能出现的其他错误
func mergeYAML(local, remote any, strategy args.MergeStrategy) (any, error) {
	// TODO: 实现 YAML 合并的主调度函数
	// 提示：
	// 1. 根据 strategy 参数调用不同的合并函数
	// 2. 处理可能的错误情况
	// 3. 确保返回合并结果

	// 示例代码框架（取消注释并修改）：
	switch strategy {
	case args.Keep:
	    return mergeKeep(local, remote)
	case args.KeepAll:
	    return mergeKeepAll(local, remote)
	case args.Force:
	    return mergeForce(local, remote)
	default:
	    return nil, fmt.Errorf("未知的合并策略: %v", strategy)
	}
}

// mergeKeep 实现 Keep 策略
//
// 功能：保留本地标量值，用远程替换序列。
// 算法规则：
// 1. 如果 local 和 remote 都是映射：递归合并每个键
// 2. 如果 remote 是序列：返回 remote（替换本地序列）
// 3. 其他情况：返回 local（保留本地标量值）
//
// 参数：
//   - local: 本地数据
//   - remote: 远程数据
//
// 返回值：
//   - any: 合并后的数据
//   - error: 合并失败时返回错误
func mergeKeep(local, remote any) (any, error) {
	// TODO: 实现 Keep 策略
	// 提示：
	// 1. 使用 typeSwitch 函数判断类型
	// 2. 如果是映射，递归合并每个键
	// 3. 如果是序列，返回远程值
	// 4. 其他情况返回本地值

	// 示例代码框架（取消注释并修改）：
	localType := typeSwitch(local)
	remoteType := typeSwitch(remote)

	// 双方都是映射，递归合并
	if localType == typeMapping && remoteType == typeMapping {
	    return mergeMappings(local, remote, mergeKeep)
	}

	// COMMENT: 为什么要远程替换本地？这里是 Keep
	// // 远程是序列，替换本地序列
	// if remoteType == typeSequence {
	//     return remote, nil
	// }
	//

	// 其他情况保留本地值
	return local, nil
}

// mergeKeepAll 实现 KeepAll 策略
//
// 功能：保留本地标量值，将远程序列追加到本地。
// 算法规则：
// 1. 如果 local 和 remote 都是映射：递归合并每个键
// 2. 如果 local 和 remote 都是序列：将 remote 追加到 local
// 3. 其他情况：返回 local（保留本地标量值）
//
// 参数：
//   - local: 本地数据
//   - remote: 远程数据
//
// 返回值：
//   - any: 合并后的数据
//   - error: 合并失败时返回错误
func mergeKeepAll(local, remote any) (any, error) {
	// TODO: 实现 KeepAll 策略
	// 提示：
	// 1. 使用 typeSwitch 函数判断类型
	// 2. 如果是映射，递归合并每个键
	// 3. 如果是序列，合并两个序列
	// 4. 其他情况返回本地值

	// 示例代码框架（取消注释并修改）：
	localType := typeSwitch(local)
	remoteType := typeSwitch(remote)

	// 双方都是映射，递归合并
	if localType == typeMapping && remoteType == typeMapping {
	    return mergeMappings(local, remote, mergeKeepAll)
	}

	// 双方都是序列，合并序列
	if localType == typeSequence && remoteType == typeSequence {
	    return mergeSequences(local, remote)
	}

	// 其他情况保留本地值
	return local, nil
}

// mergeForce 实现 Force 策略
//
// 功能：用远程值覆盖本地值，映射递归合并。
// 算法规则：
// 1. 如果 local 和 remote 都是映射：递归合并每个键
// 2. 其他情况：返回 remote（完全覆盖）
//
// 参数：
//   - local: 本地数据
//   - remote: 远程数据
//
// 返回值：
//   - any: 合并后的数据
//   - error: 合并失败时返回错误
func mergeForce(local, remote any) (any, error) {
	// TODO: 实现 Force 策略
	// 提示：
	// 1. 使用 typeSwitch 函数判断类型
	// 2. 如果是映射，递归合并每个键
	// 3. 其他情况返回远程值

	// 示例代码框架（取消注释并修改）：
	localType := typeSwitch(local)
	remoteType := typeSwitch(remote)

	// 双方都是映射，递归合并
	if localType == typeMapping && remoteType == typeMapping {
	    return mergeMappings(local, remote, mergeForce)
	}

	// 其他情况返回远程值
	return remote, nil
}

// ============================================================================
// 类型判断辅助函数和常量
// ============================================================================

// 值类型枚举
const (
	typeUnknown  = iota
	typeMapping  // 映射（map[any]any）
	typeSequence // 序列（[]any）
	typeScalar   // 标量（字符串、数字、布尔值等）
)

// typeSwitch 判断值的类型
//
// 功能：判断给定值属于映射、序列还是标量。
// 参数：
//   - v: 要判断的值
//
// 返回值：
//   - int: 类型常量（typeMapping, typeSequence, typeScalar, typeUnknown）
//
// 实现原理：
// 使用反射（reflect）检查值的实际类型。
// 注意：nil 被视为 typeScalar（空值）。
func typeSwitch(v any) int {
	// TODO: 实现类型判断
	// 提示：
	// 1. 使用 reflect.TypeOf(v).Kind() 获取类型种类
	// 2. 判断是否为 map[interface{}]interface{} 或类似类型
	// 3. 判断是否为 []interface{} 切片
	// 4. 其他情况视为标量

	// 示例代码框架（取消注释并修改）：
	if v == nil {
	    return typeScalar // nil 视为标量（空值）
	}

	rt := reflect.TypeOf(v)

	switch rt.Kind() {
	case reflect.Map:
	    // 检查是否是 map[interface{}]interface{} 或兼容类型
	    // 注意：yaml.Unmarshal 可能返回 map[interface{}]interface{}
	    return typeMapping
	case reflect.Slice, reflect.Array:
	    // 检查是否是 []interface{} 或兼容类型
	    return typeSequence
	default:
	    // 字符串、数字、布尔值等
	    return typeScalar
	}
}

// mergeMappings 合并两个映射
//
// 功能：递归合并两个映射，使用指定的合并函数处理值。
// 参数：
//   - localMap: 本地映射
//   - remoteMap: 远程映射
//   - mergeFunc: 值合并函数（如 mergeKeep, mergeKeepAll, mergeForce）
//
// 返回值：
//   - any: 合并后的映射
//   - error: 合并失败时返回错误
//
// 实现原理：
// 1. 创建本地映射的深拷贝（避免修改输入）
// 2. 遍历远程映射的所有键
// 3. 对每个键，使用 mergeFunc 合并本地和远程值
// 4. 将结果放入新映射
func mergeMappings(local, remote any, mergeFunc func(any, any) (any, error)) (any, error) {
	// TODO: 实现映射合并
	// 提示：
	// 1. 将 local 和 remote 转换为 map[any]any
	// 2. 创建结果映射（深拷贝本地映射）
	// 3. 遍历远程映射的键值对
	// 4. 对每个键，递归调用 mergeFunc 合并值
	// 5. 将合并结果放入结果映射

	// 示例代码框架（取消注释并修改）：
	// 类型断言，确保是映射
	localMap, ok := local.(map[any]any)
	if !ok {
	    return nil, errors.ErrYAMLType
	}

	remoteMap, ok := remote.(map[any]any)
	if !ok {
	    return nil, errors.ErrYAMLType
	}

	// 深拷贝本地映射
	result := make(map[any]any)
	maps.Copy(result, localMap)

	// 合并远程映射
	for k, remoteValue := range remoteMap {
	    if localValue, exists := result[k]; exists {
	        // 键存在，递归合并
	        merged, err := mergeFunc(localValue, remoteValue)
	        if err != nil {
	            return nil, errors.Wrapf(err, "合并键 %v 失败", k)
	        }
	        result[k] = merged
	    } else {
	        // 键不存在，直接添加
	        result[k] = remoteValue
	    }
	}

	return result, nil
}

// mergeSequences 合并两个序列
//
// 功能：将远程序列追加到本地序列之后。
// 参数：
//   - localSeq: 本地序列
//   - remoteSeq: 远程序列
//
// 返回值：
//   - any: 合并后的序列
//   - error: 合并失败时返回错误
func mergeSequences(local, remote any) (any, error) {
	// TODO: 实现序列合并
	// 提示：
	// 1. 将 local 和 remote 转换为 []any
	// 2. 创建新切片，容量为两个切片长度之和
	// 3. 先添加本地元素，再添加远程元素

	// 示例代码框架（取消注释并修改）：
	// 类型断言，确保是切片
	localSlice, ok := local.([]any)
	if !ok {
	    return nil, errors.ErrYAMLType
	}

	remoteSlice, ok := remote.([]any)
	if !ok {
	    return nil, errors.ErrYAMLType
	}

	// 创建新切片，合并两个切片
	result := make([]any, 0, len(localSlice)+len(remoteSlice))
	result = append(result, localSlice...)
	result = append(result, remoteSlice...)

	return result, nil
}

// deepCopyValue 深拷贝一个值
//
// 功能：创建值的完全独立副本，用于避免修改输入参数。
// 参数：
//   - v: 要拷贝的值
//
// 返回值：
//   - any: 拷贝后的值
//   - error: 拷贝失败时返回错误
//
// 注意：此函数是可选辅助函数，如果实现复杂可以简化或省略。
func deepCopyValue(v any) (any, error) {
	// TODO: 实现深拷贝（可选）
	// 提示：
	// 1. 根据类型递归拷贝
	// 2. 对于映射，创建新映射并递归拷贝每个键值
	// 3. 对于序列，创建新切片并递归拷贝每个元素
	// 4. 对于标量，直接返回（值类型）

	// 简单实现：使用序列化/反序列化
	// 更简单的实现：直接返回 v（如果保证不修改）

	return v, nil
}

// 常见问题与解决方案：
//
// 1. 问题：类型断言失败导致 panic
//    解决：使用类型断言的安全形式 `value, ok := v.(目标类型)`
//
// 2. 问题：递归深度过大导致栈溢出
//    解决：YAML 配置通常不会太深，但可以考虑添加深度限制
//
// 3. 问题：循环引用导致无限递归
//    解决：YAML 通常没有循环引用，可以不处理
//
// 4. 问题：map[any]any 的键可能不是可比较类型
//    解决：YAML 键通常是字符串，确保类型安全
//
// 5. 问题：性能问题（大量反射）
//    解决：只在必要时使用反射，缓存类型信息

// 测试建议：
// 1. 为每种策略编写测试用例
// 2. 测试嵌套结构（深度嵌套的映射和序列）
// 3. 测试边界情况（nil、空值、类型不匹配）
// 4. 测试大配置文件的合并性能
