#!/usr/bin/env python3
"""Validate issue 79 PNG deliverables without third-party dependencies."""

from pathlib import Path
import struct
import zlib


FILES = (
    "maker-taker-fees-rebates-execution.png",
    "maker-taker-fees-rebates-execution-break-even.png",
    "maker-taker-fees-rebates-execution-adverse-selection.png",
    "maker-taker-fees-rebates-execution-fee-tiers.png",
    "maker-taker-fees-rebates-execution-queue-value.png",
)


def validate(path: Path) -> None:
    data = path.read_bytes()
    assert data.startswith(b"\x89PNG\r\n\x1a\n"), f"{path}: bad PNG signature"
    offset, idat, saw_iend = 8, bytearray(), False
    width = height = bit_depth = color_type = None

    while offset < len(data):
        assert offset + 12 <= len(data), f"{path}: truncated chunk header"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        end = offset + 12 + length
        assert end <= len(data), f"{path}: truncated {kind!r} chunk"
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : end])[0]
        actual_crc = zlib.crc32(kind + payload) & 0xFFFFFFFF
        assert actual_crc == expected_crc, f"{path}: bad {kind!r} CRC"
        if kind == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", payload[:10])
        elif kind == b"IDAT":
            idat.extend(payload)
        elif kind == b"IEND":
            saw_iend = True
            assert end == len(data), f"{path}: data after IEND"
        offset = end

    assert saw_iend, f"{path}: missing IEND"
    assert (width, height) == (1664, 936), f"{path}: got {width}x{height}"
    assert bit_depth == 8 and color_type in (2, 6), f"{path}: expected 8-bit RGB/RGBA"
    decompressor = zlib.decompressobj()
    decompressor.decompress(bytes(idat))
    decompressor.flush()
    assert decompressor.eof and not decompressor.unused_data, f"{path}: incomplete IDAT stream"
    assert 400_000 <= len(data) <= 4_000_000, f"{path}: implausible size {len(data)}"


def main() -> None:
    for filename in FILES:
        path = Path("blog") / filename
        validate(path)
        print(f"ok {path} ({path.stat().st_size / 1_000_000:.2f} MB)")


if __name__ == "__main__":
    main()
