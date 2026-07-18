#!/usr/bin/env python3
"""Decode and verify the liquidation-cascade blog image acceptance criteria."""

import binascii
import struct
import zlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BLOG = ROOT / "blog"
FILENAMES = (
    "liquidation-cascades-trading-signal.png",
    "liquidation-cascades-trading-signal-depth-chart.png",
    "liquidation-cascades-trading-signal-cex-heatmap.png",
    "liquidation-cascades-trading-signal-cascade-chain.png",
    "liquidation-cascades-trading-signal-signal-map.png",
)
EXPECTED_SIZE = (1664, 936)
SIGNATURE = b"\x89PNG\r\n\x1a\n"


def decode_png(path: Path) -> tuple[int, int]:
    data, offset, idat = path.read_bytes(), len(SIGNATURE), []
    assert data.startswith(SIGNATURE), f"invalid PNG signature: {path}"
    width = height = depth = color = interlace = None
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        payload = data[offset + 8 : offset + 8 + length]
        checksum = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert binascii.crc32(kind + payload) == checksum, f"bad chunk checksum: {path}"
        if kind == b"IHDR":
            width, height, depth, color, _, _, interlace = struct.unpack(">IIBBBBB", payload)
        elif kind == b"IDAT":
            idat.append(payload)
        elif kind == b"IEND":
            break
        offset += length + 12
    assert depth == 8 and color in (2, 6) and interlace == 0, f"unsupported PNG: {path}"
    channels = 3 if color == 2 else 4
    decoded = zlib.decompress(b"".join(idat))
    assert len(decoded) == height * (1 + width * channels), f"corrupt pixel data: {path}"
    assert all(decoded[row * (1 + width * channels)] <= 4 for row in range(height)), f"bad filter: {path}"
    return width, height


def main() -> None:
    readme = (BLOG / "README.md").read_text(encoding="utf-8")
    for filename in FILENAMES:
        path = BLOG / filename
        assert path.is_file(), f"missing image: {path}"
        assert filename in readme, f"missing README mapping: {filename}"
        assert decode_png(path) == EXPECTED_SIZE, f"wrong dimensions: {path}"
        size_mib = path.stat().st_size / (1024 * 1024)
        print(f"ok: {path.relative_to(ROOT)} 1664x936 {size_mib:.2f} MiB")


if __name__ == "__main__":
    main()
