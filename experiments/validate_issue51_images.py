#!/usr/bin/env python3
"""Validate the generated image set for issue #51."""

from pathlib import Path
from struct import unpack
from zlib import crc32, decompressobj


ROOT = Path(__file__).resolve().parents[1]
EXPECTED_SIZE = (1664, 936)
MIN_BYTES = 500_000
MAX_BYTES = 1_500_000
FILES = (
    "asymmetric-garch-crypto-leverage.png",
    "asymmetric-garch-crypto-leverage-news-impact.png",
    "asymmetric-garch-crypto-leverage-fat-tails.png",
    "asymmetric-garch-crypto-leverage-var-es.png",
    "asymmetric-garch-crypto-leverage-model-selection.png",
)


def decode_png(path: Path) -> tuple[int, int, int]:
    data = path.read_bytes()
    assert data.startswith(b"\x89PNG\r\n\x1a\n"), f"{path}: not a PNG"

    offset = 8
    width = height = color_type = None
    decompressor = decompressobj()
    decompressed_bytes = 0
    saw_idat = False
    saw_iend = False

    while offset < len(data):
        assert offset + 8 <= len(data), f"{path}: truncated chunk header"
        length = unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_start = offset + 8
        chunk_end = chunk_start + length
        crc_end = chunk_end + 4
        assert crc_end <= len(data), f"{path}: truncated {chunk_type!r} chunk"

        payload = data[chunk_start:chunk_end]
        expected_crc = unpack(">I", data[chunk_end:crc_end])[0]
        actual_crc = crc32(chunk_type + payload) & 0xFFFFFFFF
        assert actual_crc == expected_crc, f"{path}: bad CRC in {chunk_type!r}"

        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type = unpack(">IIBB", payload[:10])
            assert bit_depth == 8, f"{path}: expected 8-bit PNG, got {bit_depth}"
        elif chunk_type == b"IDAT":
            saw_idat = True
            decompressed_bytes += len(decompressor.decompress(payload))
        elif chunk_type == b"IEND":
            decompressed_bytes += len(decompressor.flush())
            saw_iend = True
            offset = crc_end
            break

        offset = crc_end

    assert width is not None and height is not None, f"{path}: missing IHDR"
    assert saw_idat, f"{path}: missing IDAT"
    assert saw_iend, f"{path}: missing IEND"
    assert decompressor.eof, f"{path}: incomplete compressed image stream"
    assert offset == len(data), f"{path}: data after IEND"

    channels = {2: 3, 6: 4}.get(color_type)
    assert channels is not None, f"{path}: expected RGB/RGBA, got color type {color_type}"
    expected_bytes = height * (1 + width * channels)
    assert decompressed_bytes == expected_bytes, (
        f"{path}: decoded {decompressed_bytes:,} bytes, expected {expected_bytes:,}"
    )
    return width, height, color_type


def main() -> None:
    for filename in FILES:
        path = ROOT / "blog" / filename
        assert path.is_file(), f"missing image: {path}"

        file_size = path.stat().st_size
        assert MIN_BYTES <= file_size <= MAX_BYTES, (
            f"{filename}: expected {MIN_BYTES:,}-{MAX_BYTES:,} bytes, "
            f"got {file_size:,}"
        )

        width, height, _ = decode_png(path)
        assert (width, height) == EXPECTED_SIZE, (
            f"{filename}: expected {EXPECTED_SIZE}, got {(width, height)}"
        )

        print(f"ok: {filename} ({file_size:,} bytes, {EXPECTED_SIZE[0]}x{EXPECTED_SIZE[1]})")


if __name__ == "__main__":
    main()
