use std::{path::{Path, PathBuf}, time::Duration};

use clap::ValueEnum;
use reqwest::Client;

use crate::{AppResult, errors::AppError};

pub fn cache_file(base: impl AsRef<Path>) -> PathBuf {
    base.as_ref().join("mihomo-cache.yaml")
}

pub fn update_file(base: impl AsRef<Path>) -> PathBuf {
    base.as_ref().join("mihomo-update.yaml")
}

pub fn server_file(base: impl AsRef<Path>) -> PathBuf {
    base.as_ref().join("mihomo-server.yaml")
}

pub fn config_file(base: impl AsRef<Path>) -> PathBuf {
    base.as_ref().join("config.yaml")
}

pub async fn fetch_yaml(url: &str, timeout: u64, user_agent: &str) -> AppResult<serde_yml::Value> {
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
        Err(AppError {
            msg: t!("errors.yaml.bad-structure"),
            context: None,
            skippable: false,
        })
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
