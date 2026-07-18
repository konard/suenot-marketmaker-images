#!/usr/bin/env python3
"""Validate the required issue #93 PNG deliverables."""

from pathlib import Path
import struct
import zlib


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = [
    "mev-supply-chain-pbs-mevboost.png",
    "mev-supply-chain-pbs-mevboost-pbs-pipeline.png",
    "mev-supply-chain-pbs-mevboost-bid-shading.png",
    "mev-supply-chain-pbs-mevboost-order-flow.png",
    "mev-supply-chain-pbs-mevboost-solana-jito.png",
]
README_SECTION = "## mev-supply-chain-pbs-mevboost"
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def validate_png(path: Path) -> tuple[int, int, int]:
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), f"{path}: invalid PNG signature"

    offset = len(PNG_SIGNATURE)
    chunks = []
    width = height = None
    while offset < len(data):
        assert offset + 12 <= len(data), f"{path}: truncated chunk header"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_data = data[offset + 8 : offset + 8 + length]
        crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert len(chunk_data) == length, f"{path}: truncated {chunk_type!r} chunk"
        expected_crc = zlib.crc32(chunk_type + chunk_data) & 0xFFFFFFFF
        assert crc == expected_crc, f"{path}: corrupt {chunk_type!r} chunk"
        chunks.append(chunk_type)
        if chunk_type == b"IHDR":
            width, height = struct.unpack(">II", chunk_data[:8])
        offset += 12 + length
        if chunk_type == b"IEND":
            break

    assert chunks[0] == b"IHDR", f"{path}: IHDR is not first"
    assert b"IDAT" in chunks, f"{path}: missing image data"
    assert chunks[-1] == b"IEND", f"{path}: missing IEND"
    assert offset == len(data), f"{path}: trailing bytes after IEND"
    assert (width, height) == (1664, 936), f"{path}: got {width}x{height}"
    assert 500_000 <= len(data) <= 3_000_000, f"{path}: implausible size {len(data)}"
    return width, height, len(data)


def main() -> None:
    expected = {ROOT / "blog" / name for name in EXPECTED}
    actual = set((ROOT / "blog").glob("mev-supply-chain-pbs-mevboost*.png"))
    assert actual == expected, f"unexpected image set: {sorted(actual ^ expected)}"
    for path in sorted(expected):
        width, height, size = validate_png(path)
        print(f"{path.relative_to(ROOT)}: {width}x{height}, {size:,} bytes, CRC OK")
    readme = (ROOT / "blog" / "README.md").read_text()
    assert readme.count(README_SECTION) == 1, "missing or duplicate README section"
    for name in EXPECTED:
        assert readme.count(f"`{name}`") == 1, f"README mapping missing for {name}"
    print("blog/README.md: all five mappings present")


if __name__ == "__main__":
    main()
