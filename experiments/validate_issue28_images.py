from __future__ import annotations

import binascii
import struct
import zlib
from pathlib import Path
from statistics import mean


FILES = [
    "blog/two-axis-parameter-space.png",
    "blog/two-axis-parameter-space-two-axes.png",
    "blog/two-axis-parameter-space-cache-sweep.png",
    "blog/two-axis-parameter-space-cost-gap.png",
    "blog/two-axis-parameter-space-repriced-curse.png",
]


def read_png(path: Path) -> tuple[int, int, bytes]:
    data = path.read_bytes()
    if not data.startswith(b"\x89PNG\r\n\x1a\n"):
        raise AssertionError(f"{path}: missing PNG signature")

    offset = 8
    width = height = None
    idat = bytearray()

    while offset < len(data):
        if offset + 12 > len(data):
            raise AssertionError(f"{path}: truncated chunk header")
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_data = data[offset + 8 : offset + 8 + length]
        crc_expected = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        crc_actual = binascii.crc32(chunk_type + chunk_data) & 0xFFFFFFFF
        if crc_actual != crc_expected:
            raise AssertionError(f"{path}: bad CRC in {chunk_type!r}")

        if chunk_type == b"IHDR":
            width, height = struct.unpack(">II", chunk_data[:8])
        elif chunk_type == b"IDAT":
            idat.extend(chunk_data)
        elif chunk_type == b"IEND":
            break

        offset += 12 + length

    if width is None or height is None:
        raise AssertionError(f"{path}: missing IHDR")
    return width, height, zlib.decompress(bytes(idat))


def main() -> None:
    for name in FILES:
        path = Path(name)
        size = path.stat().st_size
        width, height, inflated = read_png(path)
        sample = inflated[:: max(1, len(inflated) // 100_000)]
        byte_mean = mean(sample)
        byte_span = max(sample) - min(sample)

        if (width, height) != (1664, 936):
            raise AssertionError(f"{name}: expected 1664x936, got {width}x{height}")
        if size > 3_500_000:
            raise AssertionError(f"{name}: unusually large PNG, got {size} bytes")
        if byte_mean < 5:
            raise AssertionError(f"{name}: image appears too dark or blank")
        if byte_span < 10:
            raise AssertionError(f"{name}: image has suspiciously low contrast")

        print(f"{name}: {width}x{height}, {size / 1_000_000:.2f} MB, sample_mean={byte_mean:.2f}")


if __name__ == "__main__":
    main()
