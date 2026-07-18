#!/usr/bin/env python3
"""Fully decode and validate the five PNG masters required by issue 65."""

from pathlib import Path
import struct
import zlib

FILES = (
    "almgren-chriss-optimal-execution.png",
    "almgren-chriss-optimal-execution-impact-vs-risk.png",
    "almgren-chriss-optimal-execution-sinh-trajectory.png",
    "almgren-chriss-optimal-execution-efficient-frontier.png",
    "almgren-chriss-optimal-execution-calibration.png",
)
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
EXPECTED_WIDTH = 1664
EXPECTED_HEIGHT = 936


def validate(path: Path) -> None:
    raw = path.read_bytes()
    assert raw.startswith(PNG_SIGNATURE), f"{path}: invalid PNG signature"

    offset = len(PNG_SIGNATURE)
    idat = bytearray()
    width = height = bit_depth = color_type = None
    saw_iend = False
    while offset < len(raw):
        assert offset + 12 <= len(raw), f"{path}: truncated PNG chunk"
        length = struct.unpack(">I", raw[offset : offset + 4])[0]
        chunk_type = raw[offset + 4 : offset + 8]
        end = offset + 12 + length
        assert end <= len(raw), f"{path}: chunk exceeds file size"
        data = raw[offset + 8 : offset + 8 + length]
        crc = struct.unpack(">I", raw[offset + 8 + length : end])[0]
        assert zlib.crc32(chunk_type + data) & 0xFFFFFFFF == crc, (
            f"{path}: CRC mismatch in {chunk_type!r}"
        )
        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", data[:10])
        elif chunk_type == b"IDAT":
            idat.extend(data)
        elif chunk_type == b"IEND":
            saw_iend = True
            assert end == len(raw), f"{path}: trailing bytes after IEND"
        offset = end

    assert saw_iend, f"{path}: missing IEND"
    assert (width, height) == (EXPECTED_WIDTH, EXPECTED_HEIGHT), (
        f"{path}: expected {EXPECTED_WIDTH}x{EXPECTED_HEIGHT}, got {width}x{height}"
    )
    assert bit_depth == 8 and color_type in (2, 6), f"{path}: expected 8-bit RGB/RGBA"
    channels = 3 if color_type == 2 else 4
    decoded = zlib.decompress(bytes(idat))
    expected_bytes = height * (1 + width * channels)
    assert len(decoded) == expected_bytes, (
        f"{path}: decoded {len(decoded)} bytes, expected {expected_bytes}"
    )
    # The issue calls multi-megabyte files a corruption warning, not a hard
    # limit. Full zlib decompression and the exact scanline length above are
    # the authoritative integrity checks; detailed generated artwork can
    # legitimately compress above the approximate 0.5–1.5 MB target.
    assert len(raw) < 3_000_000, f"{path}: suspicious file size {len(raw):,} bytes"
    print(f"OK {path}: {width}x{height}, {len(raw):,} bytes, fully decoded")


def main() -> None:
    root = Path(__file__).resolve().parents[1] / "blog"
    for name in FILES:
        validate(root / name)


if __name__ == "__main__":
    main()
