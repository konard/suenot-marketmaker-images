#!/usr/bin/env python3
"""Validate the structural delivery requirements for issue #46's PNG."""

from pathlib import Path
import struct
import sys
import zlib


PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
MAX_FILE_SIZE = 5_000_000
MAX_DIMENSION = 4_096


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def validate(path: Path) -> None:
    data = path.read_bytes()
    require(data.startswith(PNG_SIGNATURE), "not a PNG")
    require(500_000 <= len(data) <= MAX_FILE_SIZE, f"implausible file size: {len(data)} bytes")

    offset = len(PNG_SIGNATURE)
    chunks: list[str] = []
    width = height = None
    idat = bytearray()
    bit_depth = color_type = interlace = None

    while offset < len(data):
        require(offset + 12 <= len(data), "truncated PNG chunk header")
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        require(offset + 12 + length <= len(data), f"truncated {kind!r} chunk")
        payload = data[offset + 8 : offset + 8 + length]
        crc = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        require(zlib.crc32(kind + payload) & 0xFFFFFFFF == crc, f"bad {kind!r} CRC")
        chunks.append(kind.decode("ascii"))
        if kind == b"IHDR":
            width, height = struct.unpack(">II", payload[:8])
            bit_depth, color_type, _, _, interlace = payload[8:13]
        elif kind == b"IDAT":
            idat.extend(payload)
        elif kind == b"IEND":
            break
        offset += 12 + length

    require(bool(width and height), "missing IHDR")
    require(chunks[-1] == "IEND", "missing IEND")
    require(width <= MAX_DIMENSION and height <= MAX_DIMENSION, f"dimensions too large: {width}x{height}")
    require(abs(width / height - 16 / 9) < 0.02, f"not approximately 16:9: {width}x{height}")
    require(bit_depth == 8 and color_type in (2, 6), "expected 8-bit RGB or RGBA")
    require(interlace == 0, "interlaced PNG is not supported by this validator")

    channels = 3 if color_type == 2 else 4
    stride = width * channels
    expected_size = height * (stride + 1)
    decompressor = zlib.decompressobj()
    raw = decompressor.decompress(idat, expected_size + 1)
    require(len(raw) == expected_size, "unexpected decompressed scanline size")
    require(decompressor.eof and not decompressor.unused_data, "IDAT stream did not finish exactly")
    require(not decompressor.unconsumed_tail, "IDAT exceeds expected scanline size")
    previous = bytearray(stride)
    non_black = 0
    bright = 0
    sample_count = 0
    for row_index in range(height):
        start = row_index * (stride + 1)
        filter_type = raw[start]
        encoded = raw[start + 1 : start + 1 + stride]
        decoded = bytearray(stride)
        for index, value in enumerate(encoded):
            left = decoded[index - channels] if index >= channels else 0
            above = previous[index]
            upper_left = previous[index - channels] if index >= channels else 0
            if filter_type == 0:
                predictor = 0
            elif filter_type == 1:
                predictor = left
            elif filter_type == 2:
                predictor = above
            elif filter_type == 3:
                predictor = (left + above) // 2
            elif filter_type == 4:
                estimate = left + above - upper_left
                distances = abs(estimate - left), abs(estimate - above), abs(estimate - upper_left)
                predictor = (left, above, upper_left)[distances.index(min(distances))]
            else:
                raise ValueError(f"invalid PNG filter {filter_type}")
            decoded[index] = (value + predictor) & 0xFF
        previous = decoded
        if row_index % 8 == 0:
            for pixel in range(0, width, 8):
                base = pixel * channels
                red, green, blue = decoded[base : base + 3]
                alpha = decoded[base + 3] if channels == 4 else 255
                visible_max = max(red, green, blue) * alpha // 255
                sample_count += 1
                non_black += visible_max > 20
                bright += visible_max > 180
    require(non_black / sample_count > 0.70, "image is unexpectedly blank or black")
    require(bright / sample_count > 0.05, "image lacks expected visual range")
    print(f"PASS {path}: {width}x{height}, {len(data):,} bytes, chunks={','.join(chunks)}")


if __name__ == "__main__":
    validate(Path(sys.argv[1] if len(sys.argv) > 1 else "repos/video-metadata_comic.png"))
