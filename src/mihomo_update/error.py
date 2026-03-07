class MihomoUpdateError(Exception):
    """Base error for mihomo-update."""


class NetworkError(MihomoUpdateError):
    pass


class YamlParseError(MihomoUpdateError):
    pass


class FileMissingError(MihomoUpdateError):
    pass


class FileWriteError(MihomoUpdateError):
    pass
