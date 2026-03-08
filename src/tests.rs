#![cfg(test)]
use serde_yml::{Value, from_str};

use crate::helper::{MergeStrategy, merge_yaml};

// 辅助函数：将 &str 解析为 Value
fn yaml(s: &str) -> Value {
    from_str(s).unwrap()
}

// ========== MergeStrategy::Keep 测试 ==========
#[test]
fn test_keep_basic_mapping() {
    let old = yaml("a: 1\nb: 2");
    let new = yaml("b: 3\nc: 4");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("a: 1\nb: 2\nc: 4"); // b 保留旧值 2
    assert_eq!(result, expected);
}

#[test]
fn test_keep_list_replace() {
    let old = yaml("list: [1, 2]");
    let new = yaml("list: [3, 4]");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("list: [3, 4]"); // 列表被取代
    assert_eq!(result, expected);
}

#[test]
fn test_keep_new_list_over_non_list() {
    let old = yaml("a: 1");
    let new = yaml("a: [2, 3]");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("a: [2, 3]"); // new 是列表，取代 old
    assert_eq!(result, expected);
}

#[test]
fn test_keep_old_list_preserved() {
    let old = yaml("a: [1, 2]");
    let new = yaml("a: 3");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("a: [1, 2]"); // new 不是列表，保留 old
    assert_eq!(result, expected);
}

#[test]
fn test_keep_nested_list_replace() {
    let old = yaml("x:\n  y: [1, 2]");
    let new = yaml("x:\n  y: [3, 4]");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("x:\n  y: [3, 4]"); // 替换为新的，应为 x.y = [3, 4]
    assert_eq!(result, expected);
}

#[test]
fn test_keep_type_mismatch() {
    let old = yaml("a: 1");
    let new = yaml("a: hello");
    let result = merge_yaml(old, new, MergeStrategy::Keep);
    let expected = yaml("a: 1"); // 保留旧值
    assert_eq!(result, expected);
}

// ========== MergeStrategy::KeepAll 测试 ==========
#[test]
fn test_keepall_basic_mapping() {
    let old = yaml("a: 1\nb: 2");
    let new = yaml("b: 3\nc: 4");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("a: 1\nb: 2\nc: 4"); // b 保留旧值
    assert_eq!(result, expected);
}

#[test]
fn test_keepall_list_append() {
    let old = yaml("list: [1, 2]");
    let new = yaml("list: [3, 4]");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("list: [1, 2, 3, 4]"); // 追加
    assert_eq!(result, expected);
}

#[test]
fn test_keepall_nested_list_append() {
    let old = yaml("x:\n  list: [1, 2]");
    let new = yaml("x:\n  list: [3, 4]");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("x:\n  list: [1, 2, 3, 4]");
    assert_eq!(result, expected);
}

#[test]
fn test_keepall_new_list_but_old_not_list() {
    let old = yaml("a: 1");
    let new = yaml("a: [2, 3]");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("a: 1"); // 不能追加，保留旧值
    assert_eq!(result, expected);
}

#[test]
fn test_keepall_old_list_but_new_not_list() {
    let old = yaml("a: [1, 2]");
    let new = yaml("a: 3");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("a: [1, 2]"); // 保留旧列表
    assert_eq!(result, expected);
}

#[test]
fn test_keepall_multiple_lists() {
    let old = yaml("list: [1, 2]");
    let new = yaml("list: [3, 4]\nother: [5]");
    let result = merge_yaml(old, new, MergeStrategy::KeepAll);
    let expected = yaml("list: [1, 2, 3, 4]\nother: [5]");
    assert_eq!(result, expected);
}

// ========== MergeStrategy::Force 测试（基于实际递归行为）==========
#[test]
fn test_force_basic_mapping() {
    let old = yaml("a: 1\nb: 2");
    let new = yaml("b: 3\nc: 4");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    let expected = yaml("a: 1\nb: 3\nc: 4"); // b 被覆盖为 3
    assert_eq!(result, expected);
}

#[test]
fn test_force_nested_mapping_merge() {
    let old = yaml("a:\n  b: 1\n  c: 2");
    let new = yaml("a:\n  b: 3\n  d: 4");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    // 递归合并，c 被保留，a，b 均被替换
    let expected = yaml("a:\n  b: 3\n  c: 2\n  d: 4");
    assert_eq!(result, expected);
}

#[test]
fn test_force_new_scalar_over_mapping() {
    let old = yaml("a:\n  b: 1");
    let new = yaml("a: 2");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    let expected = yaml("a: 2"); // 非映射直接覆盖
    assert_eq!(result, expected);
}

#[test]
fn test_force_new_mapping_over_scalar() {
    let old = yaml("a: 1");
    let new = yaml("a:\n  b: 2");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    let expected = yaml("a:\n  b: 2"); // 非映射 old 被 new 覆盖
    assert_eq!(result, expected);
}

#[test]
fn test_force_list_replace() {
    let old = yaml("list: [1, 2]");
    let new = yaml("list: [3, 4]");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    let expected = yaml("list: [3, 4]"); // 序列被替换
    assert_eq!(result, expected);
}

#[test]
fn test_force_complex_nested() {
    let old = yaml("a:\n  b: [1, 2]\n  c: 3");
    let new = yaml("a:\n  b: [4, 5]\n  d: 6");
    let result = merge_yaml(old, new, MergeStrategy::Force);
    // b 被替换，c 保留，d 新增
    let expected = yaml("a:\n  b: [4, 5]\n  c: 3\n  d: 6");
    assert_eq!(result, expected);
}
