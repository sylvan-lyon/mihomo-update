#!/usr/bin/env python3

import sys
import argparse
from pathlib import Path

import requests
import yaml


def fatal(msg: str, code: int = 2):
    print(msg, file=sys.stderr)
    sys.exit(code)


def fetch_yaml(url: str, timeout: int) -> dict:
    try:
        resp = requests.get(
            url,
            headers={"User-Agent": "Clash"},
            timeout=timeout,
        )
    except requests.RequestException as e:
        fatal(f"request error: {e}")

    if not resp.ok:
        fatal(f"http error {resp.status_code}")

    try:
        return yaml.safe_load(resp.text)
    except yaml.YAMLError as e:
        fatal(f"yaml decode error: {e}")


def read_yaml(path: Path) -> dict:
    try:
        with path.open() as f:
            doc = yaml.safe_load(f)
    except FileNotFoundError:
        fatal(f"file not found: {path}")
    except yaml.YAMLError as e:
        fatal(f"yaml error in {path}: {e}")

    if not isinstance(doc, dict):
        fatal(f"{path} is not a yaml object")

    return doc


def extract_sub(doc: dict) -> tuple[list, list]:
    proxies = doc.get("proxies")
    rules = doc.get("rules")

    if proxies is None:
        fatal("subscription missing 'proxies'")

    if rules is None:
        fatal("subscription missing 'rules'")

    return proxies, rules


def write_yaml(path: Path, data: dict):
    try:
        with path.open("w") as f:
            yaml.safe_dump(
                data,
                f,
                allow_unicode=True,
                default_flow_style=False,
            )
    except OSError as e:
        fatal(f"write error: {e}")


def main():
    parser = argparse.ArgumentParser(
        description="Update mihomo proxies from subscription"
    )

    parser.add_argument("--url", required=True)
    parser.add_argument("--path", required=True)
    parser.add_argument("--timeout", required=False)

    args = parser.parse_args()

    base = Path(args.path)

    mihomo_cfg = read_yaml(base / "mihomo-server.yaml")
    print("Loaded mihomo configuration")

    timeout = args.timeout if args.timeout is not None else 60
    sub_doc = fetch_yaml(args.url, timeout)
    print(f"Fetched subscription, timeout {timeout}")

    proxies, rules = extract_sub(sub_doc)

    mihomo_cfg["proxies"] = proxies
    mihomo_cfg["rules"] = rules

    write_yaml(base / "config.yaml", mihomo_cfg)

    print("Configuration updated")


if __name__ == "__main__":
    main()
