#[macro_use]
extern crate rust_i18n;

mod args;
mod errors;
mod helper;
mod run;
mod tests;


use clap::Parser;

use crate::errors::AppError;

rust_i18n::i18n!("locales", fallback = "en");

pub type Translated = Cow<'static, str>;
pub type AppResult<T> = Result<T, AppError>;
pub type SkippableResult = Result<(), AppError>;


#[tokio::main]
async fn main() {
    init_locale();

    if let Err(e) = run::run(args::Args::parse()).await {
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
