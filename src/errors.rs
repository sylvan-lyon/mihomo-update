use std::fmt::Display;

use crate::Translated;

pub trait ResultExt<T> {
    fn context(self, ctx: Translated) -> Self;
    fn celebrate(self, msg: Translated) -> Self;
    fn print(self) -> Self;
}

impl<T> ResultExt<T> for Result<T, AppError> {
    /// 此函数可以轻松的添加上下文信息
    #[inline]
    fn context(self, ctx: Translated) -> Self {
        match self {
            Err(mut e) => Err({
                e.context = Some(ctx);
                e
            }),
            Ok(v) => Ok(v),
        }
    }

    /// 成功了就庆祝一下
    #[inline]
    fn celebrate(self, msg: Translated) -> Self {
        if self.is_ok() {
            println!("{msg}");
        }
        self
    }

    fn print(self) -> Self {
        self.map_err(|mut e| {
            e.skippable = true;
            println!("{e}");
            e
        })
    }
}

/// 第一个 `Translated` 表示错误原因，是错误的第一手表示
///
/// 第二个 `Option<Translated>` 表示在哪出错了，是上下文信息
pub struct AppError {
    pub msg: Translated,
    pub context: Option<Translated>,
    pub skippable: bool,
}

impl Display for AppError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let Self {
            msg: message,
            context,
            skippable,
        } = self;
        match (context, skippable) {
            (None, true) => f.write_str(&t!("errors.fmt.without-context.skippable", msg = message)),
            (None, false) => f.write_str(&t!("errors.fmt.without-context.fatal", msg = message)),
            (Some(ctx), true) => f.write_str(&t!(
                "errors.fmt.with-context.skippable",
                msg = message,
                context = ctx
            )),
            (Some(ctx), false) => f.write_str(&t!(
                "errors.fmt.with-context.fatal",
                msg = message,
                context = ctx
            )),
        }
    }
}

impl From<reqwest::Error> for AppError {
    fn from(err: reqwest::Error) -> Self {
        let msg = if err.is_builder() {
            t!("errors.network.builder")
        } else if err.is_request() {
            t!("errors.network.request")
        } else if err.is_redirect() {
            t!("errors.network.redirect")
        } else if err.is_body() {
            t!("errors.network.body")
        } else if err.is_decode() {
            t!("errors.network.decode")
        } else if err.is_upgrade() {
            t!("errors.network.upgrade")
        } else if err.is_status() {
            t!("errors.network.status", status = err.status().unwrap())
        } else {
            unreachable!("network error only has those variants above")
        };

        Self {
            msg,
            context: None,
            skippable: false,
        }
    }
}

impl From<std::io::Error> for AppError {
    #[inline]
    fn from(err: std::io::Error) -> Self {
        Self {
            msg: err.to_string().into(),
            context: None,
            skippable: false,
        }
    }
}

impl From<serde_yml::Error> for AppError {
    #[inline]
    fn from(err: serde_yml::Error) -> Self {
        Self {
            msg: err.to_string().into(),
            context: None,
            skippable: false,
        }
    }
}
