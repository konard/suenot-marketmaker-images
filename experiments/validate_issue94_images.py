#!/usr/bin/env python3
"""Dependency-free PNG validation for issue 94's generated image family."""

from __future__ import annotations

import binascii
import struct
import zlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FILES = [
    ROOT / "blog/uniswap-v3-lp-strategies-hedging.png",
    ROOT / "blog/uniswap-v3-lp-strategies-hedging-range-from-vol.png",
    ROOT / "blog/uniswap-v3-lp-strategies-hedging-rebalance-band.png",
    ROOT / "blog/uniswap-v3-lp-strategies-hedging-delta-hedge.png",
    ROOT / "blog/uniswap-v3-lp-strategies-hedging-jit-liquidity.png",
]
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def validate(path: Path) -> None:
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), f"{path}: invalid PNG signature"
    offset = len(PNG_SIGNATURE)
    width = height = bit_depth = color_type = None
    compressed = bytearray()
    chunks: list[bytes] = []

    while offset < len(data):
        assert offset + 12 <= len(data), f"{path}: truncated chunk header"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_end = offset + 12 + length
        assert chunk_end <= len(data), f"{path}: truncated {chunk_type!r} chunk"
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : chunk_end])[0]
        actual_crc = binascii.crc32(chunk_type + payload) & 0xFFFFFFFF
        assert actual_crc == expected_crc, f"{path}: bad {chunk_type!r} CRC"
        chunks.append(chunk_type)

        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type, _, _, interlace = struct.unpack(
                ">IIBBBBB", payload
            )
            assert interlace == 0, f"{path}: unexpected interlacing"
        elif chunk_type == b"IDAT":
            compressed.extend(payload)
        elif chunk_type == b"IEND":
            assert chunk_end == len(data), f"{path}: bytes after IEND"
            break
        offset = chunk_end

    assert chunks[0] == b"IHDR" and chunks[-1] == b"IEND", f"{path}: malformed chunks"
    assert (width, height) == (1664, 936), f"{path}: got {width}x{height}"
    assert bit_depth == 8 and color_type in (2, 6), f"{path}: unsupported pixel format"
    channels = 3 if color_type == 2 else 4
    decoded = zlib.decompress(compressed)
    expected_size = height * (1 + width * channels)
    assert len(decoded) == expected_size, f"{path}: incomplete decoded pixel stream"
    assert not {b"tEXt", b"zTXt", b"iTXt"}.intersection(chunks), f"{path}: embedded text"
    assert 500_000 <= len(data) <= 2_500_000, f"{path}: implausible size {len(data)}"
    print(f"OK {path.relative_to(ROOT)}: {width}x{height}, {len(data):,} bytes")


if __name__ == "__main__":
    for image in FILES:
        validate(image)
