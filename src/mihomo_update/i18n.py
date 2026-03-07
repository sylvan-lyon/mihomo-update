import gettext
import locale
from pathlib import Path

DOMAIN = "mihomo_update"

def get_translator(lang: str | None = None):
    if lang is None:
        lang, _ = locale.getlocale()

    locale_dir = Path(__file__).parent / "locale"
    translation = gettext.translation(
        DOMAIN,
        localedir=locale_dir,
        languages=[lang] if lang else None,
        fallback=True,
    )

    return translation.gettext
