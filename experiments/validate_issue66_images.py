#!/usr/bin/env python3
"""Validate the exact image deliverables requested by issue 66."""

import struct
import zlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = [
    "twap-vwap-pov-execution-algorithms.png",
    "twap-vwap-pov-execution-algorithms-three-schedulers.png",
    "twap-vwap-pov-execution-algorithms-volume-curve.png",
    "twap-vwap-pov-execution-algorithms-pov-feedback.png",
    "twap-vwap-pov-execution-algorithms-benchmark-race.png",
]


def validate_png(path: Path) -> tuple[int, int, int, int]:
    data = path.read_bytes()
    assert data[:8] == b"\x89PNG\r\n\x1a\n", f"{path}: invalid PNG signature"

    offset = 8
    width = height = bit_depth = color_type = None
    compressed = bytearray()
    saw_iend = False
    while offset < len(data):
        assert offset + 12 <= len(data), f"{path}: truncated chunk"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        end = offset + 12 + length
        assert end <= len(data), f"{path}: chunk exceeds file"
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : end])[0]
        actual_crc = zlib.crc32(kind + payload) & 0xFFFFFFFF
        assert actual_crc == expected_crc, f"{path}: bad {kind!r} CRC"
        if kind == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", payload[:10])
        elif kind == b"IDAT":
            compressed.extend(payload)
        elif kind == b"IEND":
            saw_iend = True
            assert end == len(data), f"{path}: trailing bytes after IEND"
        offset = end

    assert saw_iend, f"{path}: missing IEND"
    assert (width, height) == (1664, 936), f"{path}: got {width}x{height}"
    assert bit_depth == 8 and color_type in (2, 6), f"{path}: expected 8-bit RGB/RGBA"
    channels = 3 if color_type == 2 else 4
    decoded = zlib.decompress(compressed)
    expected_length = height * (1 + width * channels)
    assert len(decoded) == expected_length, f"{path}: incomplete pixel decode"
    return width, height, channels, len(data)


def main() -> None:
    for filename in EXPECTED:
        path = ROOT / "blog" / filename
        assert path.is_file(), f"missing {path.relative_to(ROOT)}"
        width, height, channels, size = validate_png(path)
        print(f"ok {path.relative_to(ROOT)}: {width}x{height}, {channels} channels, {size} bytes")


if __name__ == "__main__":
    main()
