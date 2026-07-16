#!/usr/bin/env python3
"""Validate the README hero comic generated for issue 45."""

from pathlib import Path
import math
import struct
import sys
import zlib


PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
EXPECTED_ASPECT_RATIO = 16 / 9


def validate(path: Path) -> None:
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), "not a PNG file"

    offset = len(PNG_SIGNATURE)
    chunks: list[bytes] = []
    compressed_data = bytearray()
    width = height = bit_depth = color_type = None
    saw_iend = False

    while offset < len(data):
        assert offset + 12 <= len(data), "truncated PNG chunk"
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_data = data[offset + 8 : offset + 8 + length]
        crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert len(chunk_data) == length, "truncated PNG chunk data"
        assert zlib.crc32(chunk_type + chunk_data) & 0xFFFFFFFF == crc, (
            f"bad CRC for {chunk_type.decode('ascii', errors='replace')}"
        )
        chunks.append(chunk_type)

        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", chunk_data[:10])
        elif chunk_type == b"IDAT":
            compressed_data.extend(chunk_data)
        elif chunk_type == b"IEND":
            saw_iend = True
            assert offset + 12 + length == len(data), "trailing data after IEND"

        offset += 12 + length

    assert chunks[0] == b"IHDR", "IHDR must be the first chunk"
    assert saw_iend, "missing IEND chunk"
    assert b"IDAT" in chunks, "missing image data"
    assert width is not None and height is not None
    assert width >= 1600 and height >= 900, f"resolution too small: {width}x{height}"
    assert abs(width / height - EXPECTED_ASPECT_RATIO) <= 0.02, (
        f"expected approximately 16:9, got {width}:{height}"
    )
    assert bit_depth == 8, f"expected 8-bit PNG, got {bit_depth}-bit"
    assert color_type in (2, 6), f"expected RGB/RGBA PNG, got color type {color_type}"
    assert len(data) >= 500_000, f"unexpectedly small image ({len(data)} bytes)"

    channels = 3 if color_type == 2 else 4
    scanline_size = 1 + width * channels
    pixels = zlib.decompress(compressed_data)
    assert len(pixels) == scanline_size * height, "decoded image data has wrong size"
    assert all(pixels[row * scanline_size] <= 4 for row in range(height)), (
        "invalid PNG scanline filter"
    )

    frequencies = [0] * 256
    for value in pixels:
        frequencies[value] += 1
    entropy = -sum(
        (count / len(pixels)) * math.log2(count / len(pixels))
        for count in frequencies
        if count
    )
    assert entropy < 7.9, f"suspiciously noise-like payload entropy: {entropy:.3f} bits/byte"

    print(
        f"OK: {path} is a complete {width}x{height} PNG "
        f"({len(data):,} bytes, decoded entropy {entropy:.3f} bits/byte)"
    )


if __name__ == "__main__":
    image = Path(sys.argv[1] if len(sys.argv) > 1 else "repos/video-maker_comic.png")
    validate(image)
