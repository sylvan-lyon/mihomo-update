use std::fmt::Display;

use crate::Translated;

pub trait ResultExt<T> {
    fn context(self, ctx: Translated) -> Result<T, AppError>;
    fn celebrate(self, msg: Translated)-> Result<T, AppError>;
}

impl<T> ResultExt<T> for Result<T, AppError> {
    /// 此函数可以轻松的添加上下文信息
    fn context(self, ctx: Translated) -> Result<T, AppError> {
        match self {
            Err(mut e) => Err({
                e.1 = Some(ctx);
                e
            }),
            Ok(v) => Ok(v),
        }
    }

    /// 成功了就庆祝一下
    fn celebrate(self, msg: Translated)-> Result<T, AppError> {
        if self.is_ok() {
            println!("{msg}");
        }
        self
    }
}

/// 第一个 `Translated` 表示错误原因，是错误的第一手表示
///
/// 第二个 `Option<Translated>` 表示在哪出错了，是上下文信息
pub struct AppError(pub Translated, pub Option<Translated>);

impl Display for AppError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let Self(msg, when) = self;
        if let Some(when) = when {
            f.write_str(&t!("errors.fmt.with-context", msg = msg, context = when))
        } else {
            f.write_str(&t!("errors.fmt.without-context", msg = msg,))
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

        Self(msg, None)
    }
}

impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        Self(err.to_string().into(), None)
    }
}

impl From<serde_yml::Error> for AppError {
    fn from(err: serde_yml::Error) -> Self {
        Self(err.to_string().into(), None)
    }
}
