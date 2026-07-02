from pathlib import Path
import struct
import zlib


FILES = [
    "blog/backtest-engine-speed-ladder.png",
    "blog/backtest-engine-speed-ladder-the-ladder.png",
    "blog/backtest-engine-speed-ladder-pandas-baseline.png",
    "blog/backtest-engine-speed-ladder-numba-jit.png",
    "blog/backtest-engine-speed-ladder-prange-parallel.png",
    "blog/backtest-engine-speed-ladder-why-not-gpu.png",
    "blog/backtest-engine-speed-ladder-real-bottleneck.png",
]

PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
CHANNELS_BY_COLOR_TYPE = {0: 1, 2: 3, 3: 1, 4: 2, 6: 4}


def parse_png(path: Path):
    data = path.read_bytes()
    if not data.startswith(PNG_SIGNATURE):
        raise ValueError("missing PNG signature")

    offset = len(PNG_SIGNATURE)
    width = height = bit_depth = color_type = None
    compressed = []

    while offset < len(data):
        if offset + 12 > len(data):
            raise ValueError("truncated chunk header")
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_data = data[offset + 8 : offset + 8 + length]
        crc_expected = struct.unpack(">I", data[offset + 8 + length : offset + 12 + length])[0]
        crc_actual = zlib.crc32(chunk_type + chunk_data) & 0xFFFFFFFF
        if crc_actual != crc_expected:
            raise ValueError(f"CRC mismatch in {chunk_type!r}")

        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", chunk_data[:10])
        elif chunk_type == b"IDAT":
            compressed.append(chunk_data)
        elif chunk_type == b"IEND":
            break
        offset += 12 + length

    if width is None or height is None:
        raise ValueError("missing IHDR")
    if color_type not in CHANNELS_BY_COLOR_TYPE:
        raise ValueError(f"unsupported color type {color_type}")
    if bit_depth != 8:
        raise ValueError(f"unexpected bit depth {bit_depth}")

    raw = zlib.decompress(b"".join(compressed))
    channels = CHANNELS_BY_COLOR_TYPE[color_type]
    row_bytes = width * channels + 1
    if len(raw) != row_bytes * height:
        raise ValueError("unexpected decompressed scanline length")

    sample = raw[1:: max(1, len(raw) // 200_000)]
    distinct = len(set(sample))
    return width, height, bit_depth, color_type, len(raw), distinct


for name in FILES:
    path = Path(name)
    width, height, bit_depth, color_type, raw_len, distinct = parse_png(path)
    ratio_ok = width * 9 == height * 16
    size = path.stat().st_size
    print(
        f"{name}: {width}x{height}, {size / 1024:.1f} KiB, "
        f"bit_depth={bit_depth}, color_type={color_type}, raw={raw_len}, "
        f"sample_distinct={distinct}, 16:9={ratio_ok}"
    )
    if not ratio_ok:
        raise SystemExit(f"not 16:9: {name}")
    if distinct < 20:
        raise SystemExit(f"low scanline diversity/corruption suspected: {name}")
