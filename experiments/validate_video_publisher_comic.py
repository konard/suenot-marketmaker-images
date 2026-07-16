#!/usr/bin/env python3
"""Validate the required delivery properties of the video-publisher comic."""

from pathlib import Path
import struct
import sys
import zlib


PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def main() -> int:
    path = Path(sys.argv[1] if len(sys.argv) > 1 else "repos/video-publisher_comic.png")
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), "file is not a PNG"
    assert data[12:16] == b"IHDR", "PNG does not start with IHDR"
    width, height, bit_depth, color_type = struct.unpack(">IIBB", data[16:26])
    assert width > height, "comic must be landscape"
    assert 1.7 <= width / height <= 1.85, "comic must be approximately 16:9"
    assert bit_depth == 8, "expected an 8-bit PNG"
    assert color_type in (2, 6), "expected RGB or RGBA pixels"

    offset = 8
    compressed = bytearray()
    saw_iend = False
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert zlib.crc32(kind + payload) & 0xFFFFFFFF == expected_crc, f"invalid {kind!r} CRC"
        if kind == b"IDAT":
            compressed.extend(payload)
        if kind == b"IEND":
            saw_iend = True
            break
        offset += 12 + length

    assert saw_iend, "PNG is missing IEND"
    raw = zlib.decompress(compressed)
    channels = 3 if color_type == 2 else 4
    expected_size = height * (1 + width * channels)
    assert len(raw) == expected_size, "decoded raster size is inconsistent with IHDR"
    assert len(data) < len(raw), "PNG is not meaningfully compressed (possible noise/corruption)"
    print(f"valid PNG: {width}x{height}, {len(data):,} bytes, compression ratio {len(raw) / len(data):.2f}x")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
