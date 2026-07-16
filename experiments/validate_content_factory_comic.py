from collections import Counter
from pathlib import Path
import math
import struct
import zlib


IMAGE = Path("repos/content-factory_comic.png")
PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def read_png(path: Path):
    data = path.read_bytes()
    assert data.startswith(PNG_SIGNATURE), "missing PNG signature"

    offset = len(PNG_SIGNATURE)
    compressed = []
    chunks = []
    width = height = bit_depth = color_type = None
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        payload = data[offset + 8 : offset + 8 + length]
        expected_crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        assert zlib.crc32(kind + payload) & 0xFFFFFFFF == expected_crc, f"bad {kind!r} CRC"
        chunks.append(kind)
        if kind == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", payload[:10])
        elif kind == b"IDAT":
            compressed.append(payload)
        offset += 12 + length

    assert offset == len(data), "trailing or truncated PNG data"
    assert chunks[-1] == b"IEND", "missing IEND"
    return data, width, height, bit_depth, color_type, zlib.decompress(b"".join(compressed))


data, width, height, bit_depth, color_type, pixels = read_png(IMAGE)
assert (width, height) == (1664, 936), f"unexpected dimensions: {width}x{height}"
assert width * 9 == height * 16, "image is not exact 16:9"
assert bit_depth == 8 and color_type == 2, "image must be an 8-bit RGB PNG"
assert 1_000_000 <= len(data) <= 4_000_000, f"unexpected file size: {len(data)}"

counts = Counter(pixels)
entropy = -sum((count / len(pixels)) * math.log2(count / len(pixels)) for count in counts.values())
assert 2.0 <= entropy <= 7.5, f"suspicious decoded-byte entropy: {entropy:.3f} bits"

print(
    f"valid PNG: {width}x{height}, {len(data)} bytes, "
    f"decoded-byte entropy {entropy:.3f} bits"
)
