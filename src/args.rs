use clap::Parser;

#[derive(Debug, Parser)]
#[command(name = "mihomo update")]
#[command(about = "update your clash subscription")]
#[command(long_about = r"update your clash subscription
and merge them with your local mihomo configuration")]
pub struct Args {
    #[arg(long, short)]
    #[arg(value_name = "SUB", help = url::HELP.to_owned())]
    pub url: String,

    #[arg(long, short, value_name = "PATH")]
    #[arg(help = path::HELP.to_owned(), long_help = path::LONG_HELP.to_owned())]
    pub path: String,

    #[arg(long, short, default_value = "false")]
    #[arg(help = force::HELP.to_owned(), long_help = force::LONG_HELP.to_owned())]
    pub force: bool,

    #[arg(long, default_value = "60")]
    #[arg(help = timeout::HELP.to_owned())]
    pub timeout: u64,

    #[arg(long, default_value = "clash-verge/v2.4.6")]
    #[arg(help = user_agent::HELP.to_owned(), long_help = user_agent::LONG_HELP.to_owned())]
    pub user_agent: String,

    #[arg(long)]
    #[arg(help = lang::HELP.to_owned())]
    pub lang: Option<String>,
}

mod url {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.url.help"));
}

mod path {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.path.help"));
    pub const LONG_HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.path.long_help"));
}

mod force {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.force.help"));
    pub const LONG_HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.force.long_help"));
}

mod timeout {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.timeout.help"));
}

mod user_agent {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.user-agent.help"));
    pub const LONG_HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.user-agent.long_help"));
}

mod lang {
    use std::sync::LazyLock;
    use crate::Translated;

    pub const HELP: LazyLock<Translated> = LazyLock::new(|| t!("cli.arg.lang.help"));
}
