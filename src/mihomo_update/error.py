import typing
import requests
class MihomoUpdateError(Exception):
    """Base error for mihomo-update."""
    def translate(self, translator) -> str:
        return translator(str(self))


class NetworkError(MihomoUpdateError):
    def __init__(self, cause: requests.RequestException | int):
        super()
        self.cause = cause

    def translate(self, translator) -> str:
        if isinstance(self.causes, int):
            return translator("HTTP request failed with status {}").format(self.cause)
        else:
            return translator("network error: {}").format(str(self.cause))


class YamlParseError(MihomoUpdateError):
    def __init__(self, cause: typing.Literal["deserialize", "structure"]):
        super()
        self.cause = cause

    def translate(self, translator) -> str:
        if self.cause == "deserialize":
            return translator("failed to parse the content as YAML")
        else:
            return translator("YAML structure is invalid")


class FileMissingError(MihomoUpdateError):
    def __init__(self, path: str):
        super()
        self.path = path

    def translate(self, translator) -> str:
        return translator("file {} not exists").format(self.path)


class FileWriteError(MihomoUpdateError):
    def __init__(self, err: OSError):
        super()
        self.err = err

    def translate(self, translator) -> str:
        return translator("failed to write file: {}").format(str(self.err))
