#!/usr/bin/env python3

import argparse
from pathlib import Path

from mihomo_update.error import MihomoUpdateError
from mihomo_update.i18n import get_translator
from mihomo_update.helper import (
    fatal,
    deep_merge,
    fetch_yaml,
    read_yaml,
    write_yaml,
)
i18n = get_translator()

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
    except MihomoUpdateError as e:
        fatal(i18n("Tried to read mihomo-server config, but {}").format(e.translate(i18n)))
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
    except MihomoUpdateError as e:
        fatal(i18n("Tried to fetch subscription config, but {}").format(e.translate(i18n)))
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
    except MihomoUpdateError as e:
        fatal(i18n("Cannot write merged configuration, because {}").format(e.translate(i18n)))
    else:
        print(i18n("mihomo configuration updated successfully!"))


if __name__ == "__main__":
    main()
