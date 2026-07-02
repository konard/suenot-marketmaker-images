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
CHANNELS_BY_COLOR_TYPE = {2: 3, 6: 4}
FILTER_NONE = 0


def read_png(path: Path):
    data = path.read_bytes()
    if not data.startswith(PNG_SIGNATURE):
        raise ValueError("missing PNG signature")
    offset = len(PNG_SIGNATURE)
    compressed = []
    width = height = bit_depth = color_type = None
    while offset < len(data):
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_data = data[offset + 8 : offset + 8 + length]
        if chunk_type == b"IHDR":
            width, height, bit_depth, color_type = struct.unpack(">IIBB", chunk_data[:10])
        elif chunk_type == b"IDAT":
            compressed.append(chunk_data)
        elif chunk_type == b"IEND":
            break
        offset += 12 + length
    if bit_depth != 8 or color_type not in CHANNELS_BY_COLOR_TYPE:
        raise ValueError(f"unsupported PNG format: bit_depth={bit_depth}, color_type={color_type}")
    channels = CHANNELS_BY_COLOR_TYPE[color_type]
    row_size = width * channels
    raw = zlib.decompress(b"".join(compressed))
    rows = [bytearray(raw[y * (row_size + 1) : (y + 1) * (row_size + 1)]) for y in range(height)]
    return width, height, channels, rows


def unfilter(rows, width, channels):
    bpp = channels
    prior = bytearray(width * channels)
    result = []
    for encoded in rows:
        filter_type = encoded[0]
        scan = bytearray(encoded[1:])
        out = bytearray(len(scan))
        for i, value in enumerate(scan):
            left = out[i - bpp] if i >= bpp else 0
            up = prior[i]
            up_left = prior[i - bpp] if i >= bpp else 0
            if filter_type == 0:
                recon = value
            elif filter_type == 1:
                recon = value + left
            elif filter_type == 2:
                recon = value + up
            elif filter_type == 3:
                recon = value + ((left + up) // 2)
            elif filter_type == 4:
                p = left + up - up_left
                pa, pb, pc = abs(p - left), abs(p - up), abs(p - up_left)
                predictor = left if pa <= pb and pa <= pc else up if pb <= pc else up_left
                recon = value + predictor
            else:
                raise ValueError(f"unknown filter {filter_type}")
            out[i] = recon & 0xFF
        result.append(out)
        prior = out
    return result


def chunk(kind, payload):
    return (
        struct.pack(">I", len(payload))
        + kind
        + payload
        + struct.pack(">I", zlib.crc32(kind + payload) & 0xFFFFFFFF)
    )


def write_png(path: Path, width, height, color_type, rows):
    raw = b"".join(bytes([FILTER_NONE]) + bytes(row) for row in rows)
    ihdr = struct.pack(">IIBBBBB", width, height, 8, color_type, 0, 0, 0)
    path.write_bytes(PNG_SIGNATURE + chunk(b"IHDR", ihdr) + chunk(b"IDAT", zlib.compress(raw, 9)) + chunk(b"IEND", b""))


for name in FILES:
    path = Path(name)
    width, height, channels, encoded_rows = read_png(path)
    rows = unfilter(encoded_rows, width, channels)
    target_width = min(width, height * 16 // 9)
    target_height = min(height, width * 9 // 16)
    target_width -= target_width % 16
    target_height = target_width * 9 // 16
    if target_height > height:
        target_height = height - (height % 9)
        target_width = target_height * 16 // 9
    x0 = (width - target_width) // 2
    y0 = (height - target_height) // 2
    cropped = [
        row[x0 * channels : (x0 + target_width) * channels]
        for row in rows[y0 : y0 + target_height]
    ]
    write_png(path, target_width, target_height, 2 if channels == 3 else 6, cropped)
    print(f"{name}: {width}x{height} -> {target_width}x{target_height}")
