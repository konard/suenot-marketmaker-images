#!/usr/bin/env python3
"""Validate the implementation-shortfall image family for issue 78."""

from __future__ import annotations

import struct
import zlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = (
    "implementation-shortfall-tca-execution.png",
    "implementation-shortfall-tca-execution-perold-gap.png",
    "implementation-shortfall-tca-execution-cost-decomposition.png",
    "implementation-shortfall-tca-execution-markout-curves.png",
    "implementation-shortfall-tca-execution-close-the-loop.png",
)
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
WIDTH, HEIGHT = 1664, 936
MIN_SIZE, MAX_SIZE = 500_000, 1_500_000


def validate_png(path: Path) -> str:
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), f"{path}: invalid PNG signature"
    assert MIN_SIZE <= len(data) <= MAX_SIZE, (
        f"{path}: {len(data):,} bytes outside requested 0.5-1.5 MB envelope"
    )

    offset = len(PNG_SIGNATURE)
    width = height = bit_depth = color_type = None
    idat = bytearray()
    saw_iend = False

    while offset < len(data):
        assert offset + 12 <= len(data), f"{path}: truncated PNG chunk header"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_end = offset + 12 + length
        assert chunk_end <= len(data), f"{path}: truncated {chunk_type!r} chunk"
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : chunk_end])[0]
        actual_crc = zlib.crc32(chunk_type + payload) & 0xFFFFFFFF
        assert actual_crc == expected_crc, f"{path}: bad CRC in {chunk_type!r}"

        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type, compression, filtering, interlace = (
                struct.unpack(">IIBBBBB", payload)
            )
            assert compression == filtering == interlace == 0, (
                f"{path}: unsupported PNG encoding"
            )
        elif chunk_type == b"IDAT":
            idat.extend(payload)
        elif chunk_type == b"IEND":
            saw_iend = True
            assert chunk_end == len(data), f"{path}: trailing bytes after IEND"
            break
        offset = chunk_end

    assert saw_iend, f"{path}: missing IEND"
    assert (width, height) == (WIDTH, HEIGHT), (
        f"{path}: expected {WIDTH}x{HEIGHT}, got {width}x{height}"
    )
    assert bit_depth == 8 and color_type in (2, 6), (
        f"{path}: expected 8-bit RGB/RGBA, got depth={bit_depth}, type={color_type}"
    )
    decoded = zlib.decompress(idat)
    channels = 3 if color_type == 2 else 4
    assert len(decoded) == HEIGHT * (1 + WIDTH * channels), (
        f"{path}: incomplete decompressed pixel stream"
    )
    return f"{path.relative_to(ROOT)}: {width}x{height}, {len(data):,} bytes"


def main() -> None:
    for filename in EXPECTED:
        print(validate_png(ROOT / "blog" / filename))


if __name__ == "__main__":
    main()
