#!/usr/bin/env python3
"""Regenerate local-font assets from font files.

Scans docs/theme/assets/fonts/ for .ttf/.otf/.woff/.woff2 and emits:

- docs/theme/assets/stylesheets/fonts.css — one @font-face block per file
  (small fonts are inlined as base64 data URIs so first paint already
  uses the real glyphs — no swap, no rewrap, no scrollbar flicker)
- docs/theme/partials/fonts-preload.html — <link rel="preload"> per file
  that stayed on disk (inlined ones need none)
- syncs the ``local_fonts`` list in docs/theme/main.html so the theme
  skips the (dead) Google Fonts request for bundled families

Every font dropped in that folder then becomes usable via docs/mkdocs.yml:

    theme:
      font:
        text: Eager Naturalist
        code: Roboto Mono

Usage:
    python docs/tools/generate-fonts-css.py      # from repo root
    python generate-fonts-css.py                 # from docs/tools/

No third-party dependencies — only the stdlib.
"""

from __future__ import annotations

import base64
import json
import re
import struct
import sys
from pathlib import Path

# Fonts at or below this size are inlined into fonts.css as base64 data URIs,
# so the glyphs arrive with the render-blocking stylesheet and first paint
# already uses the real font: no swap, no line-rewrap, no scrollbar flicker
# on tab navigation. Larger fonts stay as files + <link rel="preload">.
INLINE_LIMIT = 102_400

FORMAT_BY_SUFFIX = {
    ".ttf": "truetype",
    ".otf": "opentype",
    ".woff": "woff",
    ".woff2": "woff2",
}

MIME_BY_SUFFIX = {
    ".ttf": "font/ttf",
    ".otf": "font/otf",
    ".woff": "font/woff",
    ".woff2": "font/woff2",
}

# Filename hints -> (weight, style). Checked case-insensitively.
WEIGHT_HINTS = [
    ("thin", 100),
    ("extralight", 200),
    ("ultralight", 200),
    ("light", 300),
    ("regular", 400),
    ("normal", 400),
    ("medium", 500),
    ("semibold", 600),
    ("demibold", 600),
    ("bold", 700),
    ("extrabold", 800),
    ("ultrabold", 800),
    ("black", 900),
    ("heavy", 900),
]


def _read_name_table(data: bytes) -> str | None:
    """Return preferred family name (nameID 1) from a SFNT (ttf/otf)."""
    if len(data) < 12:
        return None
    try:
        num_tables = struct.unpack(">H", data[4:6])[0]
    except struct.error:
        return None
    name_off = name_len = None
    for i in range(num_tables):
        entry = data[12 + i * 16 : 12 + (i + 1) * 16]
        if len(entry) < 16:
            break
        tag, _, offset, length = struct.unpack(">4sIII", entry)
        if tag == b"name":
            name_off, name_len = offset, length
            break
    if name_off is None:
        return None
    try:
        count = struct.unpack(">H", data[name_off + 2 : name_off + 4])[0]
        string_base = name_off + struct.unpack(
            ">H", data[name_off + 4 : name_off + 6]
        )[0]
    except struct.error:
        return None

    candidates: list[tuple[int, str]] = []  # (priority, text)
    for i in range(count):
        rec = data[name_off + 6 + i * 12 : name_off + 6 + (i + 1) * 12]
        if len(rec) < 12:
            break
        plat, enc, lang, nid, length, noff = struct.unpack(">HHHHHH", rec)
        if nid != 1:
            continue
        raw = data[string_base + noff : string_base + noff + length]
        try:
            if plat == 3 or (plat == 0):
                text = raw.decode("utf-16-be").strip()
            elif plat == 1 and enc == 0:
                text = raw.decode("mac_roman").strip()
            else:
                text = raw.decode("utf-16-be", errors="replace").strip()
        except (UnicodeDecodeError, struct.error):
            continue
        if not text:
            continue
        # Prefer Windows English, then Mac English, then anything.
        if plat == 3 and lang == 1033:
            priority = 0
        elif plat == 1 and lang == 0:
            priority = 1
        else:
            priority = 2
        candidates.append((priority, text))
    if not candidates:
        return None
    candidates.sort(key=lambda c: c[0])
    return candidates[0][1]


def _read_weight_style(data: bytes) -> tuple[int | None, str | None]:
    """Read OS/2 weight + head macStyle from SFNT. Returns (weight, style)."""
    try:
        num_tables = struct.unpack(">H", data[4:6])[0]
    except struct.error:
        return None, None
    tables: dict[bytes, tuple[int, int]] = {}
    for i in range(num_tables):
        entry = data[12 + i * 16 : 12 + (i + 1) * 16]
        if len(entry) < 16:
            break
        tag, _, offset, length = struct.unpack(">4sIII", entry)
        tables[tag] = (offset, length)
    weight = style = None
    if b"OS/2" in tables:
        off, _ = tables[b"OS/2"]
        try:
            weight = struct.unpack(">H", data[off + 4 : off + 6])[0]
        except struct.error:
            weight = None
    if b"head" in tables:
        off, _ = tables[b"head"]
        try:
            mac_style = struct.unpack(">H", data[off + 44 : off + 46])[0]
            if mac_style & 0b10:  # italic bit
                style = "italic"
        except struct.error:
            pass
    return weight, style


def _fallback_family(stem: str) -> str:
    return stem.replace("_", " ").replace("-", " ").strip().title() or stem


def _guess_from_filename(stem: str) -> tuple[int, str]:
    lowered = stem.lower().replace("_", "").replace("-", "")
    weight = 400
    for hint, value in WEIGHT_HINTS:
        if hint in lowered:
            weight = value
            break
    if "italic" in lowered or "oblique" in lowered:
        style = "italic"
    else:
        style = "normal"
    return weight, style


def font_face(font_path: Path) -> tuple[str, str, bool]:
    """Return (family, @font-face block, inlined) for one font file."""
    suffix = font_path.suffix.lower()
    css_format = FORMAT_BY_SUFFIX[suffix]
    data = font_path.read_bytes()

    family: str | None = None
    weight: int | None = None
    style: str | None = None
    if suffix in (".ttf", ".otf") and data[:4] in (
        b"\x00\x01\x00\x00",
        b"OTTO",
        b"true",
        b"typ1",
    ):
        family = _read_name_table(data)
        weight, style = _read_weight_style(data)

    guess_weight, guess_style = _guess_from_filename(font_path.stem)
    if weight not in range(1, 1000):
        weight = guess_weight
    if style not in ("normal", "italic", "oblique"):
        # Oblique is rarely declared in TTFs; keep italic vs normal.
        style = guess_style if style is None else "normal"
    if not family:
        family = _fallback_family(font_path.stem)

    if len(data) <= INLINE_LIMIT:
        mime = MIME_BY_SUFFIX[suffix]
        blob = base64.b64encode(data).decode("ascii")
        src = f'url("data:{mime};base64,{blob}") format("{css_format}")'
        inlined = True
    else:
        src = f'url("../fonts/{font_path.name}") format("{css_format}")'
        inlined = False

    block = (
        "@font-face {\n"
        f'  font-family: "{family}";\n'
        f"  src: {src};\n"
        f"  font-weight: {weight};\n"
        f"  font-style: {style};\n"
        # block (not swap): paint nothing until the font is ready
        # instead of flashing the fallback font on every navigation.
        # Inlined fonts are ready with the stylesheet, so this is instant.
        "  font-display: block;\n"
        "}"
    )
    return family, block, inlined


def sync_main_html(theme_dir: Path, families: list[str]) -> None:
    """Rewrite the ``local_fonts`` list in theme main.html (best effort)."""
    main_html = theme_dir / "main.html"
    if not main_html.is_file():
        return
    wanted = json.dumps(sorted(set(families)), ensure_ascii=False)
    text = main_html.read_text(encoding="utf-8")
    updated, count = re.subn(
        r"{%\s*set\s+local_fonts\s*=\s*\[.*?\]\s*%}",
        "{%% set local_fonts = %s %%}" % wanted,
        text,
        count=1,
        flags=re.DOTALL,
    )
    if count:
        main_html.write_text(updated, encoding="utf-8")
        print(f"synced local_fonts in {main_html}: {wanted}")
    else:
        print(
            f"note: no local_fonts list found in {main_html}, "
            "Google Fonts requests are not skipped for local families",
            file=sys.stderr,
        )


def main() -> int:
    here = Path(__file__).resolve()
    # .../docs/tools/generate-fonts-css.py -> .../docs/
    docs_dir = here.parent.parent
    theme_dir = docs_dir / "theme"
    fonts_dir = theme_dir / "assets" / "fonts"
    css_path = theme_dir / "assets" / "stylesheets" / "fonts.css"
    preload_path = theme_dir / "partials" / "fonts-preload.html"

    if not fonts_dir.is_dir():
        print(f"fonts dir not found: {fonts_dir}", file=sys.stderr)
        return 1

    files = sorted(
        p
        for p in fonts_dir.iterdir()
        if p.is_file() and p.suffix.lower() in FORMAT_BY_SUFFIX
    )
    if not files:
        print(f"no font files in {fonts_dir}", file=sys.stderr)
        return 1

    faces: list[str] = []
    families: list[str] = []
    kept_files: list[Path] = []
    inlined: list[bool] = []
    for path in files:
        try:
            family, block, inline = font_face(path)
        except OSError as exc:
            print(f"skip {path.name}: {exc}", file=sys.stderr)
            continue
        faces.append(block)
        families.append(family)
        kept_files.append(path)
        inlined.append(inline)

    example = families[0] if families else "Eager Naturalist"
    header = (
        "/* Local fonts bundled with the theme.\n"
        " *\n"
        " * Generated from docs/theme/assets/fonts/ — do not edit by hand.\n"
        " * Re-run `python docs/tools/generate-fonts-css.py` after adding/removing\n"
        " * font files. Then use the Family name in docs/mkdocs.yml:\n"
        " *\n"
        " *   theme:\n"
        " *     font:\n"
        f" *       text: {example}   # body font\n"
        " *       code: Roboto Mono        # code font (or another local family)\n"
        " */\n"
    )
    css_path.parent.mkdir(parents=True, exist_ok=True)
    css_path.write_text(header + "\n" + "\n\n".join(faces) + "\n", encoding="utf-8")

    print(f"wrote {css_path} ({len(faces)} font(s)):")
    for path, family, inline in zip(kept_files, families, inlined):
        how = "inlined" if inline else "file"
        print(f"  {path.name} -> \"{family}\" ({how})")

    # Preload hints for the fonts that stayed as files (inlined ones ride
    # along with the stylesheet, so preloading them would double-fetch).
    # This partial is included from main.html.
    # NOTE: the {{ ... | url }} bits below are Jinja for MkDocs, not
    # Python — they are written through verbatim.
    lines = [
        "{# Local font preloads — generated by docs/tools/generate-fonts-css.py, do not edit by hand. #}"
    ]
    for path, inline in zip(kept_files, inlined):
        if inline:
            continue
        mime = MIME_BY_SUFFIX[path.suffix.lower()]
        lines.append(
            "<link rel=\"preload\" href=\"{{ 'assets/fonts/%s' | url }}\""
            " as=\"font\" type=\"%s\" crossorigin>" % (path.name, mime)
        )
    preload_path.parent.mkdir(parents=True, exist_ok=True)
    preload_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {preload_path} ({len(lines) - 1} preload(s))")

    sync_main_html(theme_dir, families)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
