#!/usr/bin/env python3

import argparse
import copy
import sys
from pathlib import Path

import requests
import yaml

from mihomo_update.i18n import get_translator
from mihomo_update.error import *
i18n = get_translator()

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
    """
    # Raises
    NetworkError YamlParseError
    """
    try:
        resp = requests.get(
            url,
            headers={"User-Agent": user_agent},
            timeout=timeout,
        )
    except requests.RequestException as e:
        raise NetworkError(i18n("NetworkError: {}").format(str(e))) from e

    if not resp.ok:
        raise NetworkError(i18n("HTTP request failed with status {}").format(resp.status_code))

    try:
        return yaml.safe_load(resp.text)
    except yaml.YAMLError as e:
        raise YamlParseError(i18n("Failed to parse content of {} as YAML").format(url)) from e


def read_yaml(path: Path) -> dict:
    """
    # Raises
    FileNotFoundError YamlParseError
    """
    try:
        with path.open() as f:
            doc = yaml.safe_load(f)
    except FileNotFoundError as e:
        raise FileMissingError(i18n("File {} not exsists").format(str(path))) from e
    except yaml.YAMLError as e:
        raise YamlParseError(i18n("Failed to parse response as YAML: {}").format(str(path))) from e

    if not isinstance(doc, dict):
        raise YamlParseError(i18n("Invalid YAML structure in {}").format(path))

    return doc


def write_yaml(path: Path, data: dict):
    """
    # Raises
    FileWriteError
    """
    try:
        with path.open("w") as f:
            yaml.safe_dump(
                data,
                f,
                allow_unicode=True,
                default_flow_style=False,
            )
    except OSError as e:
        raise FileWriteError(i18n("Failed to write file: {}").format(e))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=i18n("Manage mihomo configuration")
    )

    parser.add_argument("--url", required=True, help=i18n("Subscription URL"))
    parser.add_argument("--path", required=True, help=i18n("Configuration directory"))
    parser.add_argument(
        "--timeout",
        type=int,
        default=60,
        help=i18n("Timeout for fetching subscription"),
    )
    parser.add_argument(
        "--user-agent",
        default="clash-verge/v2.4.0",
        help=i18n("User-Agent for subscription request"),
    )
    parser.add_argument(
        "--lang",
        choices=["en", "zh_CN"],
        help=i18n("Override language (e.g. en, zh_CN)"),
    )

    return parser.parse_args()

def main():
    global i18n

    args = parse_args()
    if args.lang:
        i18n = get_translator(args.lang)

    base = Path(args.path)
    mihomo_cfg, sub_doc = None, None

    # server config
    try:
        mihomo_cfg = read_yaml(base / "mihomo-server.yaml")
    except e:
        fatal(i18n("Tried to read mihomo-server config, but {}").format(str(e)))
    else:
        print(i18n("Loaded mihomo server configuration"))

    # fetch_yaml
    try:
        print(
            i18n("Fetching subscription... timeout={} UA={}").format(
                args.timeout, args.user_agent
            )
        )
        sub_doc = fetch_yaml(args.url, args.timeout, args.user_agent)
    except e:
        fatal(i18n("Tried to fetch subscription config, but {}").format(str(e)))
    else:
        print(i18n("Successfully got subscription config!"))

    # write cache
    # try:
    #     write_yaml(base / "cache.yaml", sub_doc)
    # except e:
    #     print(i18n("Cannot write subscription config to cache, because {}, skipped").format(str(e)))
    # else:
    #     print(i18n("Cached subscription successfully!"))

    # merge and write final config
    merged = deep_merge(mihomo_cfg, sub_doc)
    try:
        write_yaml(base / "config.yaml", merged)
    except e:
        fatal(i18n("Cannot write merged configuration, because {}").format(str(e)))
    else:
        print(i18n("mihomo configuration updated successfully!"))


if __name__ == "__main__":
    main()
