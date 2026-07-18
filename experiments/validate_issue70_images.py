#!/usr/bin/env python3
"""Validate the image deliverables for issue 70."""

import struct
import zlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = (
    "onchain-liquidations-aave-compound.png",
    "onchain-liquidations-aave-compound-health-factor.png",
    "onchain-liquidations-aave-compound-oracle-trigger.png",
    "onchain-liquidations-aave-compound-liquidation-bot.png",
    "onchain-liquidations-aave-compound-bonus-competition.png",
)


def validate_png(path: Path) -> None:
    data = path.read_bytes()
    assert data[:8] == b"\x89PNG\r\n\x1a\n", f"{path}: invalid PNG signature"

    offset = 8
    width = height = None
    saw_iend = False
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk = data[offset + 8 : offset + 8 + length]
        crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        expected_crc = zlib.crc32(chunk_type + chunk) & 0xFFFFFFFF
        assert crc == expected_crc, f"{path}: corrupt {chunk_type!r} chunk"
        if chunk_type == b"IHDR":
            width, height = struct.unpack(">II", chunk[:8])
        if chunk_type == b"IEND":
            saw_iend = True
            break
        offset += 12 + length

    assert saw_iend, f"{path}: missing IEND chunk"
    assert (width, height) == (1664, 936), f"{path}: got {width}x{height}"


def main() -> None:
    for filename in EXPECTED:
        path = ROOT / "blog" / filename
        assert path.is_file(), f"missing {path}"
        validate_png(path)
        print(f"ok: {path.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
