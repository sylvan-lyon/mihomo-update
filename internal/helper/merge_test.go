// merge_test.go - YAML 合并策略测试
//
// 本文件使用 Go 的 testing 框架测试 merge.go 中的合并函数。
// Go 测试特点：
// 1. 文件以 _test.go 结尾
// 2. 函数名以 Test 开头，接受 *testing.T 参数
// 3. 使用表格驱动测试（table-driven tests）组织测试用例
// 4. 支持子测试（t.Run）和并行测试
//
// 表格驱动测试模式：
// 1. 定义测试用例结构体（包含输入和期望输出）
// 2. 创建测试用例切片
// 3. 遍历切片，对每个用例运行子测试
// 4. 使用 t.Errorf 报告失败
//
// 术语表（中英对照）：
// - table-driven test: 表格驱动测试，使用数据表组织测试用例
// - subtest: 子测试，将大测试分解为小的独立测试
// - t.Run: 运行子测试的方法
// - t.Error/t.Errorf: 报告测试失败但继续执行
// - t.Fatal/t.Fatalf: 报告测试失败并立即终止
// - test fixture: 测试夹具，测试前的准备和清理工作

package helper

import (
	"reflect"
	"testing"

	"github.com/sylvan-lyon/mihomo-update/internal/args"
)

// TestTypeSwitch 测试 typeSwitch 函数
func TestTypeSwitch(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		// 标量测试
		{"nil", nil, typeScalar},
		{"string", "hello", typeScalar},
		{"int", 42, typeScalar},
		{"float", 3.14, typeScalar},
		{"bool", true, typeScalar},

		// 序列测试
		{"empty slice", []any{}, typeSequence},
		{"string slice", []any{"a", "b", "c"}, typeSequence},
		{"mixed slice", []any{1, "two", true}, typeSequence},
		{"nested slice", []any{[]any{1, 2}, []any{3, 4}}, typeSequence},

		// 映射测试
		{"empty map", map[any]any{}, typeMapping},
		{"simple map", map[any]any{"key": "value"}, typeMapping},
		{"nested map", map[any]any{"nested": map[any]any{"a": 1}}, typeMapping},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeSwitch(tt.input)
			if got != tt.expected {
				t.Errorf("typeSwitch(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestMergeSequences 测试 mergeSequences 函数
func TestMergeSequences(t *testing.T) {
	tests := []struct {
		name     string
		local    any
		remote   any
		expected any
		wantErr  bool
	}{
		{
			name:     "两个空切片",
			local:    []any{},
			remote:   []any{},
			expected: []any{},
			wantErr:  false,
		},
		{
			name:     "本地空，远程有元素",
			local:    []any{},
			remote:   []any{1, 2, 3},
			expected: []any{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "本地有元素，远程空",
			local:    []any{1, 2, 3},
			remote:   []any{},
			expected: []any{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "合并两个切片",
			local:    []any{"a", "b"},
			remote:   []any{"c", "d"},
			expected: []any{"a", "b", "c", "d"},
			wantErr:  false,
		},
		{
			name:     "类型不匹配 - 本地不是切片",
			local:    "not a slice",
			remote:   []any{1, 2, 3},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "类型不匹配 - 远程不是切片",
			local:    []any{1, 2, 3},
			remote:   "not a slice",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeSequences(tt.local, tt.remote)

			if (err != nil) != tt.wantErr {
				t.Errorf("mergeSequences() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflectDeepEqual(got, tt.expected) {
				t.Errorf("mergeSequences() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMergeMappings 测试 mergeMappings 函数
func TestMergeMappings(t *testing.T) {
	// 简单的合并函数，用于测试
	simpleMerge := func(local, remote any) (any, error) {
		// 总是返回远程值，模拟 Force 策略
		return remote, nil
	}

	tests := []struct {
		name     string
		local    any
		remote   any
		expected any
		wantErr  bool
	}{
		{
			name:     "两个空映射",
			local:    map[any]any{},
			remote:   map[any]any{},
			expected: map[any]any{},
			wantErr:  false,
		},
		{
			name:     "添加新键",
			local:    map[any]any{"a": 1},
			remote:   map[any]any{"b": 2},
			expected: map[any]any{"a": 1, "b": 2},
			wantErr:  false,
		},
		{
			name:     "覆盖现有键",
			local:    map[any]any{"a": 1, "b": 2},
			remote:   map[any]any{"b": 20, "c": 3},
			expected: map[any]any{"a": 1, "b": 20, "c": 3}, // b 被覆盖
			wantErr:  false,
		},
		{
			name:     "类型不匹配 - 本地不是映射",
			local:    "not a map",
			remote:   map[any]any{"a": 1},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "类型不匹配 - 远程不是映射",
			local:    map[any]any{"a": 1},
			remote:   "not a map",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeMappings(tt.local, tt.remote, simpleMerge)

			if (err != nil) != tt.wantErr {
				t.Errorf("mergeMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflectDeepEqual(got, tt.expected) {
				t.Errorf("mergeMappings() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMergeStrategies 测试三种合并策略
func TestMergeStrategies(t *testing.T) {
	// 辅助函数：创建测试用的 YAML 数据结构
	createTestData := func() (local, remote map[any]any) {
		local = map[any]any{
			"scalar": "local-value",
			"list":   []any{"local-item1", "local-item2"},
			"nested": map[any]any{
				"inner": "local-inner",
				"list":  []any{"nested-local"},
			},
		}

		remote = map[any]any{
			"scalar": "remote-value",       // 标量，与本地冲突
			"list":   []any{"remote-item"}, // 序列，与本地冲突
			"nested": map[any]any{
				"inner": "remote-inner",         // 嵌套标量
				"list":  []any{"nested-remote"}, // 嵌套序列
			},
			"new": "remote-only", // 远程独有的键
		}

		return local, remote
	}

	tests := []struct {
		name     string
		strategy args.MergeStrategy
		// 期望结果的检查函数
		check func(t *testing.T, result any)
	}{
		{
			name:     "Keep 策略",
			strategy: args.Keep,
			check: func(t *testing.T, result any) {
				m, ok := result.(map[any]any)
				if !ok {
					t.Fatal("结果不是映射")
				}

				// Keep: 保留本地值，仅在双方均为映射时递归合并
				if m["scalar"] != "local-value" {
					t.Errorf("Keep 策略应保留本地标量，got %v", m["scalar"])
				}

				// 序列保留
				remoteList, ok := m["list"].([]any)
				if !ok || len(remoteList) != 2 || remoteList[0] != "local-item1" || remoteList[1] != "local-item2" {
					t.Errorf("Keep 策略应保留本地序列值，got %v", m["list"])
				}

				// 检查嵌套
				nested, ok := m["nested"].(map[any]any)
				if !ok {
					t.Fatal("嵌套不是映射")
				}

				if nested["inner"] != "local-inner" {
					t.Errorf("嵌套标量应保留本地值，got %v", nested["inner"])
				}

				nestedList, ok := nested["list"].([]any)
				if !ok || len(nestedList) != 1 || nestedList[0] != "nested-local" {
					t.Errorf("嵌套序列应保留本地值，got %v", nested["list"])
				}

				// 新键应该添加
				if m["new"] != "remote-only" {
					t.Errorf("新键应添加，got %v", m["new"])
				}
			},
		},
		{
			name:     "MergeAll 策略",
			strategy: args.MergeAll,
			check: func(t *testing.T, result any) {
				m, ok := result.(map[any]any)
				if !ok {
					t.Fatal("结果不是映射")
				}

				// MergeAll: 保留本地标量值，合并序列（融合所有选项）
				if m["scalar"] != "local-value" {
					t.Errorf("MergeAll 策略应保留本地标量，got %v", m["scalar"])
				}

				// 序列应该合并
				list, ok := m["list"].([]any)
				if !ok || len(list) != 3 {
					t.Errorf("MergeAll 策略应合并序列，期望 3 个元素，got %v", list)
				}

				// 检查嵌套
				nested, ok := m["nested"].(map[any]any)
				if !ok {
					t.Fatal("嵌套不是映射")
				}

				if nested["inner"] != "local-inner" {
					t.Errorf("嵌套标量应保留本地值，got %v", nested["inner"])
				}

				nestedList, ok := nested["list"].([]any)
				if !ok || len(nestedList) != 2 {
					t.Errorf("嵌套序列应合并，期望 2 个元素，got %v", nestedList)
				}
			},
		},
		{
			name:     "Force 策略",
			strategy: args.Force,
			check: func(t *testing.T, result any) {
				m, ok := result.(map[any]any)
				if !ok {
					t.Fatal("结果不是映射")
				}

				// Force: 远程覆盖本地
				if m["scalar"] != "remote-value" {
					t.Errorf("Force 策略应用远程覆盖标量，got %v", m["scalar"])
				}

				// 序列应该被远程替换
				list, ok := m["list"].([]any)
				if !ok || len(list) != 1 || list[0] != "remote-item" {
					t.Errorf("Force 策略应用远程替换序列，got %v", m["list"])
				}

				// 检查嵌套
				nested, ok := m["nested"].(map[any]any)
				if !ok {
					t.Fatal("嵌套不是映射")
				}

				if nested["inner"] != "remote-inner" {
					t.Errorf("嵌套标量应用远程覆盖，got %v", nested["inner"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local, remote := createTestData()

			result, err := mergeYAML(local, remote, tt.strategy)
			if err != nil {
				t.Fatalf("mergeYAML() 失败: %v", err)
			}

			tt.check(t, result)
		})
	}
}

// TestMergeYAMLEdgeCases 测试边界情况
func TestMergeYAMLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		local    any
		remote   any
		strategy args.MergeStrategy
		wantErr  bool
	}{
		{
			name:     "nil 值",
			local:    nil,
			remote:   "value",
			strategy: args.Keep,
			wantErr:  false,
		},
		{
			name:     "类型不匹配：映射 vs 序列",
			local:    map[any]any{"a": 1},
			remote:   []any{1, 2, 3},
			strategy: args.Keep,
			wantErr:  false, // Keep 策略应该处理这种情况
		},
		{
			name:     "未知策略",
			local:    map[any]any{"a": 1},
			remote:   map[any]any{"b": 2},
			strategy: args.MergeStrategy(99), // 无效策略
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mergeYAML(tt.local, tt.remote, tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// reflectDeepEqual 使用反射比较两个值是否深度相等
// Go 的 reflect.DeepEqual 可以比较复杂结构，包括切片和映射
func reflectDeepEqual(a, b any) bool {
	// 注意：reflect.DeepEqual 对于包含 nil 和空切片的比较可能有问题
	// 但对于测试目的，它通常足够好
	return reflect.DeepEqual(a, b)
}
