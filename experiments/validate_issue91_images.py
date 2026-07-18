#!/usr/bin/env python3
"""Validate the deliverables for issue 91."""

from pathlib import Path
import struct
import zlib


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = (
    "impermanent-loss-lvr-lp-profitability.png",
    "impermanent-loss-lvr-lp-profitability-il-curve.png",
    "impermanent-loss-lvr-lp-profitability-concentrated-leverage.png",
    "impermanent-loss-lvr-lp-profitability-lvr-decomposition.png",
    "impermanent-loss-lvr-lp-profitability-markout.png",
)
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def validate_png(path: Path) -> None:
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), f"{path}: invalid PNG signature"
    assert len(data) >= 500_000, f"{path}: unexpectedly small ({len(data)} bytes)"

    offset = len(PNG_SIGNATURE)
    chunks: list[tuple[bytes, bytes]] = []
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        payload = data[offset + 8 : offset + 8 + length]
        stored_crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert zlib.crc32(kind + payload) & 0xFFFFFFFF == stored_crc, (
            f"{path}: corrupt {kind.decode('ascii', errors='replace')} chunk"
        )
        chunks.append((kind, payload))
        offset += 12 + length
        if kind == b"IEND":
            break

    assert offset == len(data), f"{path}: trailing or truncated PNG data"
    assert chunks[0][0] == b"IHDR" and chunks[-1][0] == b"IEND", f"{path}: bad chunk order"
    width, height, bit_depth, color_type, compression, filter_method, interlace = struct.unpack(
        ">IIBBBBB", chunks[0][1]
    )
    assert (width, height) == (1664, 936), f"{path}: expected 1664x936, got {width}x{height}"
    assert bit_depth == 8, f"{path}: expected 8-bit channels"
    assert color_type in (2, 6), f"{path}: expected RGB/RGBA, got color type {color_type}"
    assert (compression, filter_method, interlace) == (0, 0, 0), f"{path}: unsupported encoding"

    channels = 3 if color_type == 2 else 4
    raw = zlib.decompress(b"".join(payload for kind, payload in chunks if kind == b"IDAT"))
    assert len(raw) == height * (1 + width * channels), f"{path}: decoded payload has wrong size"


def main() -> None:
    for filename in EXPECTED:
        path = ROOT / "blog" / filename
        assert path.is_file(), f"missing {path}"
        validate_png(path)
        print(f"OK {path.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
