# Agent Guidelines for mihomo-update

This document provides guidelines for AI agents working on the mihomo-update Rust project. It includes build, test, and lint commands, as well as code style conventions.

## Build Commands

### Standard Build
```bash
cargo build
```
Build the project in debug mode.

### Release Build
```bash
cargo build --release
```
Build optimized binaries for production. The binary will be located at `target/release/mihomo_update`.

### Run with Default Arguments
```bash
cargo run --release -- \
  --url "https://example.com/sub" \
  --path /path/to/config
```
Run the application with the provided arguments. The `--release` flag is recommended for production-like performance.

### Run with All Options
```bash
cargo run --release -- \
  --url "https://example.com/sub" \
  --path /path/to/config \
  --force \
  --timeout 60 \
  --user-agent "clash-verge/v2.4.6"
```

## Test Commands

### Run All Tests
```bash
cargo test
```
Execute all unit tests located in `src/tests.rs`.

### Run a Single Test
```bash
cargo test test_keep_basic_mapping
```
Run a specific test by name (supports substring matching).

### Run Tests with Output
```bash
cargo test -- --nocapture
```
Show output (e.g., `println!`) during test execution.

### Test Coverage
No built-in coverage tool is configured. Use external tools like `cargo tarpaulin` or `grcov` if needed.

## Lint and Formatting

### Format Code
```bash
cargo fmt
```
Apply Rust formatting using `rustfmt`.

### Check Formatting
```bash
cargo fmt -- --check
```
Verify code style without making changes.

### Lint with Clippy
```bash
cargo clippy
```
Run the Rust linter for common mistakes and improvements.

### Lint with Clippy (Strict)
```bash
cargo clippy -- -D warnings
```
Treat all warnings as errors.

## Code Style Guidelines

### Imports Order
1. Standard library (`std::`)
2. External crates (`clap`, `reqwest`, `tokio`, etc.)
3. Internal modules (`crate::errors`, `crate::helper`)

Example from `src/helper.rs`:
```rust
use std::{path::{Path, PathBuf}, time::Duration};
use clap::ValueEnum;
use reqwest::Client;
use crate::{AppResult, errors::AppError};
```

### Module Declaration
Modules are declared in `src/main.rs` and each has its own file:
```rust
mod args;
mod errors;
mod helper;
mod run;
mod tests;
```
Do not use `mod.rs` files.

### Type Aliases
The project defines several type aliases in `src/main.rs`:
```rust
pub type Translated = Cow<'static, str>;
pub type AppResult<T> = Result<T, AppError>;
pub type Skippable<T> = Result<T, AppError>;
```
Use `AppResult` for fallible functions that return `AppError`. Use `Skippable` for errors that can be safely ignored (printed but not fatal).

### Error Handling
- Use `AppError` struct for all errors.
- Implement `From` conversions for external error types (see `src/errors.rs`).
- Use the `ResultExt` trait to add context or celebrate successes:
  ```rust
  fn foo() -> AppResult<()> {
      some_operation().context(t!("errors.context.foo"))?;
      another_operation().celebrate(t!("success.foo"))?;
      Ok(())
  }
  ```
- Error messages are internationalized via `t!` macro from `rust_i18n`.

### Naming Conventions
- **Variables and functions**: `snake_case`
- **Structs and enums**: `PascalCase`
- **Constants**: `SCREAMING_SNAKE_CASE`
- **Modules**: `snake_case` (file names)
- **Traits**: `PascalCase` with `Ext` suffix for extension traits (`ResultExt`)

### Async/Await
The project uses Tokio runtime. The `main` function is annotated with `#[tokio::main]`. All I/O and network operations should be async and use `await`. Use `tokio::fs` instead of `std::fs`.

### Internationalization (i18n)
- The project uses `rust_i18n` for translations.
- Locale files are in `locales/` directory (YAML format).
- Use the `t!` macro to get translated strings:
  ```rust
  let msg = t!("cli.arg.url.help");
  ```
- Fallback language is English (`en`). Chinese (`zh-CN`) is also supported.
- The `--lang` command-line option can override the locale.

### YAML Handling
- Use `serde_yml` for YAML serialization/deserialization.
- The `serde_yml::Value` type represents arbitrary YAML data.
- Helper functions `read_yaml`, `write_yaml`, `fetch_yaml` are provided in `src/helper.rs`.

### Configuration Merging
Three merge strategies are defined in `MergeStrategy` enum:
- `Keep`: preserve local scalar values, replace sequences with remote.
- `KeepAll`: preserve local scalar values, append remote sequences.
- `Force`: override local values with remote values (recursive merge).
Use `merge_yaml` function to apply a strategy.

### Testing Patterns
- Tests are located in `src/tests.rs` under `#[cfg(test)]`.
- Use helper function `yaml` to parse YAML strings.
- Follow the pattern: arrange, act, assert with descriptive test names.
- Each test focuses on a single scenario.
- Use `assert_eq!` for comparing `serde_yml::Value`.

### Comments and Documentation
- Use regular comments (`//`) for implementation details.
- Use doc comments (`///`) for public items.
- Chinese comments are present in the codebase and are acceptable.
- Keep comments concise and relevant.

## Project Structure

```
src/
├── main.rs          # Entry point, type aliases, locale initialization
├── args.rs          # Command-line argument parsing (clap)
├── errors.rs        # AppError, ResultExt, From conversions
├── helper.rs        # File utilities, YAML fetching, merge strategies
├── run.rs           # Main application logic
└── tests.rs         # Unit tests for merge strategies
locales/
├── en.yml           # English translations
└── zh-CN.yml        # Chinese translations
```

## Common Workflows

### Adding a New Command-Line Argument
1. Add field to `Args` struct in `src/args.rs`.
2. Provide `#[arg]` attributes with help text (use `t!` macro).
3. Update `run.rs` to handle the new argument.
4. Update locale files with translation keys.

### Adding a New Error Variant
1. Add a new translation key in locale files.
2. Use `t!("errors.new.variant")` when creating `AppError`.
3. Consider adding `From` conversion if the error originates from a library.

### Adding a New Merge Strategy
1. Add variant to `MergeStrategy` enum in `src/helper.rs`.
2. Implement the merge logic in a new function (e.g., `merge_yaml_new_strategy`).
3. Add the function to the `merge_yaml` match.
4. Write comprehensive tests in `src/tests.rs`.

### Running CI Checks Locally
```bash
cargo fmt -- --check && cargo clippy -- -D warnings && cargo test
```
This ensures code passes formatting, linting, and tests.

## Cursor and Copilot Rules

No Cursor rules (`.cursor/rules/` or `.cursorrules`) or Copilot rules (`.github/copilot-instructions.md`) are present in this repository. Follow the guidelines in this document for consistent code style.

## Notes for AI Agents

- Always run `cargo fmt` after making changes to ensure consistent formatting.
- Run `cargo clippy` to catch common mistakes before submitting code.
- Ensure tests pass (`cargo test`) before considering a task complete.
- When editing locale files, maintain both English and Chinese translations.
- Use the existing patterns for error handling and async operations.
- Follow the import order outlined above.
- Keep functions small and focused; reuse existing helper functions.
- When in doubt, mimic the style of surrounding code.