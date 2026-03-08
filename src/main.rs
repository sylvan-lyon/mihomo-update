#[macro_use]
extern crate rust_i18n;

mod args;
mod helper;

use std::time::Duration;

use clap::Parser;
use reqwest::Client;

use crate::args::Args;

rust_i18n::i18n!("locales", fallback = "en");
pub type Translated = Cow<'static, str>;

async fn run(
    Args {
        url,
        path,
        force,
        timeout,
        user_agent,
        lang: _,
    }: Args,
) {
    let resp = Client::new()
        .get(url)
        .header("User-Agent", user_agent)
        .timeout(Duration::new(timeout, 0))
        .send()
        .await.unwrap();

    println!("{}", resp.text().await.unwrap())
}

#[tokio::main]
async fn main() {
    let args: Vec<_> = std::env::args().into_iter().collect();
    if let Some(idx) = args.iter().position(|args| args.eq("--lang"))
        && let Some(locale) = args.get(idx + 1)
    {
        rust_i18n::set_locale(locale);
    }

    run(args::Args::parse()).await;
}
