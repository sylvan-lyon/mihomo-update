from pathlib import Path
import yaml.cyaml
import requests
import copy
import sys

from mihomo_update.error import (
    NetworkError,
    YamlParseError,
    FileMissingError,
    FileWriteError,
)

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
        raise NetworkError(e) from e

    if not resp.ok:
        raise NetworkError(resp.status_code)

    try:
        doc = yaml.safe_load(resp.text)
        if not isinstance(doc, dict):
            raise YamlParseError("structure")
        return doc
    except yaml.YAMLError as e:
        raise YamlParseError("deserialize") from e


def read_yaml(path: Path) -> dict:
    """
    # Raises
    FileNotFoundError YamlParseError
    """
    try:
        with path.open() as f:
            doc = yaml.safe_load(f)
    except FileNotFoundError as e:
        raise FileMissingError(str(path)) from e
    except yaml.YAMLError as e:
        raise YamlParseError("deserialize") from e

    if not isinstance(doc, dict):
        raise YamlParseError("structure")

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
        raise FileWriteError(e) from e
