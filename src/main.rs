#[macro_use]
extern crate rust_i18n;

mod args;
mod errors;
mod helper;

use std::path::PathBuf;

use clap::Parser;

use crate::{
    args::Args,
    errors::{AppError, ResultExt},
    helper::{fetch_sub, merge_yaml, read_yaml, write_yaml},
};

rust_i18n::i18n!("locales", fallback = "en");

pub type Translated = Cow<'static, str>;
pub type AppResult<T> = Result<T, AppError>;

#[allow(unused_variables)]
async fn run(
    Args {
        url,
        path,
        force,
        timeout,
        user_agent: ua,
        lang: _,
    }: Args,
) -> AppResult<()> {
    let base = PathBuf::from(path);

    let sub_yaml = fetch_sub(&url, timeout, ua)
        .await
        .context(t!("process.fetch-sub"))
        .celebrate(t!("success.fetch-sub"))?;

    let local_yaml = read_yaml(&base.join("mihomo-server.yaml")).await.context(t!("process.read-local"))
        .celebrate(t!("success.read-local"))?;

    let merged = merge_yaml(local_yaml, sub_yaml, helper::MergeStrategy::Keep);

    write_yaml(&base.join("config.yaml"), &merged)
        .await
        .context(t!("process.save-merged"))
        .celebrate(t!("success.save-merged"))
}

#[tokio::main]
async fn main() {
    init_locale();

    if let Err(e) = run(args::Args::parse()).await {
        eprintln!("{e}");
        std::process::exit(1);
    }
}

fn init_locale() {
    let mut args = std::env::args();

    while let Some(arg) = args.next() {
        if arg == "--lang" {
            if let Some(locale) = args.next() {
                rust_i18n::set_locale(&locale);
            }
        }
    }
}
