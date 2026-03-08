use std::time::Duration;

use reqwest::Client;

pub async fn fetch_sub(url: String, timeout: u64, user_agent: String) {
    let resp = Client::new()
        .get(url)
        .header("User-Agent", user_agent)
        .timeout(Duration::new(timeout, 0))
        .send()
        .await
        .unwrap();

    println!("{}", resp.text().await.unwrap())
}
