#!/usr/bin/env python3

from mihomo_update.error import FileWriteError
from mihomo_update.error import YamlParseError
from mihomo_update.error import FileMissingError
import datetime
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

def fetch_and_cache(base: Path, args: argparse.Namespace) -> dict:
    # fetch_yaml
    try:
        print(
            i18n("Fetching subscription... timeout={} UA={}").format(
                args.timeout, args.user_agent
            )
        )
        doc = fetch_yaml(args.url, args.timeout, args.user_agent)
    except MihomoUpdateError as e:
        fatal(i18n("Tried to fetch subscription config, but {}").format(e.translate(i18n)))
    else:
        print(i18n("Successfully got subscription config!"))

    # write cache
    try:
        write_yaml(base / "cache.yaml", doc)
    except MihomoUpdateError as e:
        print(i18n("Cannot write subscription config to cache, because {}, skipped").format(str(e)))
    else:
        print(i18n("Cached subscription successfully!"))

    # touch mihomo-update.yaml
    try:
        write_yaml(base / "mihomo-update.yaml", {
            "updated_at": datetime.datetime.now()
        })
    except FileWriteError as e:
        print(i18n("Cannot write mihomo-update info, because {}, skipped").format(e.translate(i18n)))
    else:
        print(i18n("Successfully wrote mihomo-update info"))

    return doc

def try_read_cache(base: Path, args: argparse.Namespace) -> dict:
    try:
        doc = read_yaml(base / "cache.yaml")
    except FileMissingError as e:
        print(i18n("Cannot read cache file, because {}, re-caching").format(e.translate(i18n)))
        return fetch_and_cache(base, args)
    except YamlParseError as e:
        print(i18n("Cannot parse the cache file, because {}, re-caching").format(e.translate((i18n))))
        return fetch_and_cache(base, args)
    else:
        print(i18n("Read subscription config from cache, add `--force` to override."))
        return doc

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=i18n("Manage mihomo configuration")
    )

    parser.add_argument("--url", required=True, help=i18n("Subscription URL"))
    parser.add_argument("--path", required=True, help=i18n("Configuration directory"))
    parser.add_argument("--force",
        action="store_true",
        help=i18n("Force update without respect to cachefile")
    )
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


    # force update?
    if not args.force:
        # updated at
        try:
            mihomo_update_doc = read_yaml(base / "mihomo-update.yaml")
        except MihomoUpdateError as e:
            print(i18n("Tried to read mihomo-update info, but {}").format(e.translate(i18n)))
            sub_doc = fetch_and_cache(base, args)
        else:
            print(i18n("Loaded mihomo update info"))
            if mihomo_update_doc["updated_at"] + datetime.timedelta(days=1) > datetime.datetime.now():
                print(i18n("You've updated mihomo config recently, use cache"))
                sub_doc = try_read_cache(base, args)
            else:
                print(i18n("It's been a long time since last update, re-caching..."))
                sub_doc = fetch_and_cache(base, args)
    else:
        print(i18n("Foreced to update"))
        sub_doc = fetch_and_cache(base, args)



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
