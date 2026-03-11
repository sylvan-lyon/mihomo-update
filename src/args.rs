use crate::helper::MergeStrategy;
use clap::Parser;

#[derive(Debug, Parser)]
#[command(name = "mihomo update")]
#[command(about = "update your clash subscription")]
#[command(long_about = r"update your clash subscription
and merge them with your local mihomo configuration")]
pub struct Args {
    #[arg(long, short)]
    #[arg(value_name = "SUB", help = t!("cli.arg.url.help"))]
    pub url: String,

    #[arg(long, short, value_name = "PATH")]
    #[arg(help = t!("cli.arg.path.help"), long_help = t!("cli.arg.path.long-help"))]
    pub path: String,

    #[arg(long, short, default_value = "false")]
    #[arg(help = t!("cli.arg.force.help"), long_help = t!("cli.arg.force.long-help"))]
    pub force: bool,

    #[arg(long, short = 'M', default_value = "keep")]
    #[arg(help = t!("cli.arg.merge-strategy.help"), long_help = t!("cli.arg.merge-strategy.long-help"))]
    pub merge_strategy: MergeStrategy,

    #[arg(long, default_value = "60")]
    #[arg(help = t!("cli.arg.timeout.help"))]
    pub timeout: u64,

    #[arg(long, default_value = "clash-verge/v2.4.6")]
    #[arg(help = t!("cli.arg.user-agent.help"), long_help = t!("cli.arg.user-agent.long-help"))]
    pub user_agent: String,

    #[arg(long)]
    #[arg(help = t!("cli.arg.lang.help"))]
    pub lang: Option<String>,
}
