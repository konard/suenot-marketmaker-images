from pathlib import Path
from struct import unpack
from zlib import crc32


ROOT = Path(__file__).resolve().parents[1]
FILES = [
    "blog/look-ahead-bias-taxonomy.png",
    "blog/look-ahead-bias-taxonomy-same-bar-fill.png",
    "blog/look-ahead-bias-taxonomy-noise-to-sharpe.png",
    "blog/look-ahead-bias-taxonomy-dose-gradient.png",
    "blog/look-ahead-bias-taxonomy-three-leaks.png",
    "blog/look-ahead-bias-taxonomy-ground-truth-sim.png",
    "blog/look-ahead-bias-taxonomy-shift-detector.png",
]


def read_png_chunks(path: Path):
    data = path.read_bytes()
    if not data.startswith(b"\x89PNG\r\n\x1a\n"):
        raise AssertionError(f"{path}: not a PNG")

    offset = 8
    saw_iend = False
    while offset < len(data):
        if offset + 8 > len(data):
            raise AssertionError(f"{path}: truncated chunk header")
        length = unpack(">I", data[offset : offset + 4])[0]
        chunk_type = data[offset + 4 : offset + 8]
        chunk_start = offset + 8
        chunk_end = chunk_start + length
        crc_end = chunk_end + 4
        if crc_end > len(data):
            raise AssertionError(f"{path}: truncated {chunk_type!r} chunk")
        expected_crc = unpack(">I", data[chunk_end:crc_end])[0]
        actual_crc = crc32(chunk_type + data[chunk_start:chunk_end]) & 0xFFFFFFFF
        if actual_crc != expected_crc:
            raise AssertionError(f"{path}: bad CRC in {chunk_type!r}")
        yield chunk_type, data[chunk_start:chunk_end]
        offset = crc_end
        if chunk_type == b"IEND":
            saw_iend = True
            break

    if not saw_iend:
        raise AssertionError(f"{path}: missing IEND")


def png_size(path: Path) -> tuple[int, int]:
    for chunk_type, payload in read_png_chunks(path):
        if chunk_type == b"IHDR":
            width, height = unpack(">II", payload[:8])
            return width, height
    raise AssertionError(f"{path}: missing IHDR")


def main() -> None:
    for relative_path in FILES:
        path = ROOT / relative_path
        width, height = png_size(path)
        size_mb = path.stat().st_size / (1024 * 1024)
        if width * 9 != height * 16:
            raise AssertionError(f"{relative_path}: expected 16:9, got {width}x{height}")
        if size_mb > 6:
            raise AssertionError(f"{relative_path}: suspiciously large at {size_mb:.2f} MB")
        print(f"{relative_path}: {width}x{height}, {size_mb:.2f} MB")


if __name__ == "__main__":
    main()
