use chrono::{DateTime, Local, TimeDelta};
use std::{
    path::{Path, PathBuf},
    pin::Pin,
    str::FromStr,
};

use crate::{
    AppResult, Skippable,
    args::Args,
    errors::ResultExt,
    helper::{
        self, cache_file, config_file, fetch_yaml, merge_yaml, read_yaml, server_file, update_file,
        write_yaml,
    },
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

fn fetch_and_cache<'a>(
    base: impl AsRef<Path> + 'a,
    url: &'a str,
    timeout: u64,
    ua: &'a str,
) -> Pin<Box<dyn Future<Output = AppResult<serde_yml::Value>> + 'a>> {
    Box::pin(async move {
        let remote_yaml = remote_yaml(url, timeout, ua).await?;
        let (_, _) = tokio::join!(write_cache(&base, &remote_yaml), record_update(&base));

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
            fetch_and_cache(base, url, timeout, ua).await
        } else {
            match read_cache(&base).await {
                Err(_) => fetch_and_cache(base, url, timeout, ua).await,
                Ok(cache) => Ok(cache),
            }
        }
    })
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

async fn record_update(base: impl AsRef<Path>) -> Skippable<()> {
    let path = update_file(&base);
    let ctx = t!("process.record-update");

    let yaml = read_yaml(&path).await.context(ctx.clone());

    fn get_new_yaml() -> serde_yml::Value {
        let mut yaml = serde_yml::Mapping::new();
        yaml.insert("updated_at".into(), Local::now().to_rfc3339().into());
        serde_yml::Value::Mapping(yaml)
    }

    let yaml = match yaml {
        Ok(mut yaml) if yaml.get("updated_at").is_some() => {
            yaml["updated_at"] = serde_yml::Value::from(Local::now().to_rfc3339());
            yaml
        }
        Err(e) => {
            let _ = Skippable::<()>::Err(e).context(ctx.clone()).or_skip_print();
            get_new_yaml()
        }
        _ => get_new_yaml(),
    };

    write_yaml(&path, &yaml)
        .await
        .context(ctx.clone())
        .or_skip_print()
}

async fn remote_yaml(url: &str, timeout: u64, ua: &str) -> AppResult<serde_yml::Value> {
    fetch_yaml(url, timeout, ua)
        .await
        .context(t!("process.fetch-sub"))
        .celebrate(t!("success.fetch-sub"))
        .or_skip_print()
}

async fn write_cache(base: impl AsRef<Path>, value: &serde_yml::Value) -> Skippable<()> {
    write_yaml(cache_file(&base), value)
        .await
        .context(t!("process.re-cache"))
        .celebrate(t!("success.re-cache"))
        .or_skip_print()
}

async fn read_cache(base: impl AsRef<Path>) -> Skippable<serde_yml::Value> {
    read_yaml(cache_file(&base))
        .await
        .context(t!("process.read-cache"))
        .celebrate(t!("success.read-cache"))
        .or_skip_print()
}
