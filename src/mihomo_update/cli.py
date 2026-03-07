#!/usr/bin/env python3

import argparse
import copy
import sys
from pathlib import Path

import requests
import yaml


def fatal(msg: str, code: int = 2):
    print(msg, file=sys.stderr)
    sys.exit(code)

def deep_merge(a, b):
    out = copy.deepcopy(b)

    for k, v in a.items():
        if (
            k in out
            and isinstance(out[k], dict)
            and isinstance(v, dict)
        ):
            out[k] = deep_merge(out[k], v)
        else:
            out[k] = v

    return out

def fetch_yaml(url: str, timeout: int, user_agent: str) -> dict:
    try:
        resp = requests.get(
            url,
            headers={"User-Agent": user_agent},
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
    parser.add_argument("--timeout", required=False, default=60)
    parser.add_argument("--user-agent", required=False, default="clash-verge/v2.4.0")

    args = parser.parse_args()

    base = Path(args.path)

    mihomo_cfg = read_yaml(base / "mihomo-server.yaml")
    print("Loaded mihomo configuration")

    sub_doc = fetch_yaml(args.url, timeout)
    print(f"Fetched subscription, timeout {timeout}")

    merged = deep_merge(mihomo_cfg, sub_doc)

    write_yaml(base / "config.yaml", merged)

    print("Configuration updated")


if __name__ == "__main__":
    main()
