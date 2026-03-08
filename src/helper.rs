use std::{path::Path, time::Duration};

use clap::ValueEnum;
use reqwest::Client;

use crate::{AppResult, errors::AppError};

pub async fn fetch_sub(url: &str, timeout: u64, user_agent: String) -> AppResult<serde_yml::Value> {
    let resp = Client::new()
        .get(url)
        .header("User-Agent", user_agent)
        .timeout(Duration::new(timeout, 0))
        .send()
        .await?
        .error_for_status()?;

    let doc = resp.text().await?;
    let doc: serde_yml::Value = serde_yml::from_str(&doc)?;

    if doc.is_mapping() {
        Ok(doc)
    } else {
        Err(AppError(t!("errors.yaml.bad-structure"), None))
    }
}

pub async fn read_yaml(path: impl AsRef<Path>) -> AppResult<serde_yml::Value> {
    let yaml = tokio::fs::read_to_string(path).await?;
    Ok(serde_yml::from_str(&yaml)?)
}

pub async fn write_yaml(path: impl AsRef<Path>, value: &serde_yml::Value) -> AppResult<()> {
    let yaml = serde_yml::to_string(value)?;
    Ok(tokio::fs::write(path, yaml).await?)
}

/// yaml merge 的策略
#[derive(ValueEnum, Debug, Clone, Copy)]
pub enum MergeStrategy {
    /// 尽量保持原 yaml 的值，但是如果遇到列表，取代之
    Keep,
    /// 完全保留 yaml 的值，如果遇到列表，追加到 old 对应值后
    KeepAll,
    /// 如果两个 yaml 都在一个键下有值，保留 new 中的，删除原来的
    Force,
}

pub fn merge_yaml(
    old: serde_yml::Value,
    new: serde_yml::Value,
    strategy: MergeStrategy,
) -> serde_yml::Value {
    match strategy {
        MergeStrategy::Keep => merge_yaml_keep(old, new),
        MergeStrategy::KeepAll => merge_yaml_keep_all(old, new),
        MergeStrategy::Force => merge_yaml_force(old, new),
    }
}

pub fn merge_yaml_keep(old: serde_yml::Value, new: serde_yml::Value) -> serde_yml::Value {
    use serde_yml::Value;

    match (old, new) {
        (Value::Mapping(mut old_map), Value::Mapping(new_map)) => {
            for (k, new_v) in new_map {
                match old_map.remove(&k) {
                    Some(old_v) => {
                        old_map.insert(k, merge_yaml_keep(old_v, new_v));
                    }
                    None => {
                        old_map.insert(k, new_v);
                    }
                }
            }
            Value::Mapping(old_map)
        }

        (_, Value::Sequence(new_seq)) => Value::Sequence(new_seq),

        (old_val, _) => old_val,
    }
}

pub fn merge_yaml_keep_all(old: serde_yml::Value, new: serde_yml::Value) -> serde_yml::Value {
    use serde_yml::Value;
    match (old, new) {
        (Value::Mapping(mut old_map), Value::Mapping(new_map)) => {
            for (k, new_v) in new_map {
                match old_map.remove(&k) {
                    Some(old_v) => {
                        old_map.insert(k, merge_yaml_keep_all(old_v, new_v));
                    }
                    None => {
                        old_map.insert(k, new_v);
                    }
                }
            }
            Value::Mapping(old_map)
        }

        (Value::Sequence(mut old_seq), Value::Sequence(new_seq)) => {
            old_seq.extend(new_seq);
            Value::Sequence(old_seq)
        }

        (old_val, _) => old_val,
    }
}

pub fn merge_yaml_force(old: serde_yml::Value, new: serde_yml::Value) -> serde_yml::Value {
    use serde_yml::Value;
    match (old, new) {
        (Value::Mapping(mut old_map), Value::Mapping(new_map)) => {
            for (k, new_v) in new_map {
                match old_map.remove(&k) {
                    Some(old_v) => {
                        old_map.insert(k, merge_yaml_force(old_v, new_v));
                    }
                    None => {
                        old_map.insert(k, new_v);
                    }
                }
            }

            Value::Mapping(old_map)
        }

        (_, new_val) => new_val,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_yml::{Value, from_str};

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
}
