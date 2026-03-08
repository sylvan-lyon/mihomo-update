#[macro_use]
extern crate rust_i18n;

use clap::Parser;
rust_i18n::i18n!("locales", fallback = "en");

mod args;

pub type Translated = Cow<'static, str>;

#[tokio::main]
async fn main() {
    let args: Vec<_> = std::env::args().into_iter().collect();
    if let Some(idx) = args.iter().position(|args| args.eq("--lang"))
        && let Some(locale) = args.get(idx + 1)
    {
        rust_i18n::set_locale(locale);
    }

    let args = args::Args::parse();
    println!("{:#?}", args)
}
