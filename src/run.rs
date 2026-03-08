use chrono::{DateTime, Local, TimeDelta};
use std::{
    path::{Path, PathBuf},
    pin::Pin,
    str::FromStr,
};

use crate::{
    AppResult, SkippableResult,
    args::Args,
    errors::{ResultExt, Skippable},
    helper::{self, cache_file, config_file, merge_yaml, read_yaml, remote_yaml, server_file, update_file, write_yaml},
};

pub async fn run(
    Args {
        url,
        path,
        force,
        timeout,
        user_agent: ua,
        lang: _,
    }: Args,
) -> AppResult<()> {
    let base = PathBuf::from(&path);

    let (remote_yaml, local_yaml) = tokio::join!(
        if force {
            fetch_and_cache(&base, &url, timeout, &ua)
        } else {
            try_read_from_cache(&base, &url, timeout, &ua)
        },
        read_yaml(server_file(&base))
    );

    let (remote_yaml, local_yaml) = (remote_yaml?, local_yaml?);

    let merged_yaml = merge_yaml(local_yaml, remote_yaml, helper::MergeStrategy::Keep);

    write_yaml(config_file(&base), &merged_yaml)
        .await
        .context(t!("process.save-merged"))
        .celebrate(t!("success.save-merged"))
}

async fn time_to_update(base: impl AsRef<Path>) -> bool {
    let path = update_file(&base);
    let yaml = read_yaml(&path).await.context(t!("process.record-update"));

    if let Ok(yaml) = yaml
        && let Some(value) = yaml.get("updated_at")
        && let serde_yml::Value::String(date) = value
        && let Ok(updated_at) = DateTime::<Local>::from_str(date)
    {
        if updated_at + TimeDelta::days(1) < Local::now() {
            println!("{}", t!("info.too-long-since-last-update"));
            true
        } else {
            false
        }
    } else {
        true
    }
}

async fn record_update(base: impl AsRef<Path>) -> SkippableResult {
    let path = update_file(&base);
    let yaml = read_yaml(&path).await.context(t!("process.record-update"));

    let yaml = if let Ok(mut yaml) = yaml
        && yaml.get("updated_at").is_some()
    {
        yaml["updated_at"] = serde_yml::Value::from(Local::now().to_rfc3339());
        yaml
    } else {
        let mut yaml = serde_yml::Mapping::new();
        yaml.insert("updated_at".into(), Local::now().to_rfc3339().into());
        serde_yml::Value::Mapping(yaml)
    };

    write_yaml(&path, &yaml).await
}

fn fetch_and_cache<'a>(
    base: impl AsRef<Path> + 'a,
    url: &'a str,
    timeout: u64,
    ua: &'a str,
) -> Pin<Box<dyn Future<Output = AppResult<serde_yml::Value>> + 'a>> {
    Box::pin(async move {
        let remote_yaml = remote_yaml(url, timeout, ua)
            .await
            .context(t!("process.fetch-sub"))
            .celebrate(t!("success.fetch-sub"))?;

        let (cache_res, record_res) = tokio::join!(
            write_yaml(cache_file(&base), &remote_yaml),
            record_update(&base),
        );

        cache_res
            .context(t!("process.re-cache"))
            .celebrate(t!("success.re-cache"))
            .skip_and_print();

        record_res
            .context(t!("process.record-update"))
            .celebrate(t!("success.record-update"))
            .skip_and_print();

        Ok(remote_yaml)
    })
}

fn try_read_from_cache<'a>(
    base: impl AsRef<Path> + 'a,
    url: &'a str,
    timeout: u64,
    ua: &'a str,
) -> Pin<Box<dyn Future<Output = AppResult<serde_yml::Value>> + 'a>> {
    Box::pin(async move {
        if time_to_update(&base).await {
            fetch_and_cache(base, url, timeout, ua)
                .await
                .context(t!("process.re-cache"))
        } else {
            let cache_res = read_yaml(cache_file(&base))
                .await
                .context(t!("process.read-cache"))
                .celebrate(t!("success.read-cache"));

            if let Err(e) = cache_res {
                SkippableResult::Err(e).skip_and_print();

                fetch_and_cache(base, url, timeout, ua)
                    .await
                    .context(t!("process.re-cache"))
                    .celebrate(t!("success.re-cache"))
            } else {
                cache_res
            }
        }
    })
}
