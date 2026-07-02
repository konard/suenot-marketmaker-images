from pathlib import Path
import struct


ROOT = Path(__file__).resolve().parents[1]
FILES = [
    "ipc-tax-backtest-engine.png",
    "ipc-tax-backtest-engine-the-boundary.png",
    "ipc-tax-backtest-engine-serialization-tax.png",
    "ipc-tax-backtest-engine-chatty-vs-chunky.png",
    "ipc-tax-backtest-engine-spawn-cost.png",
    "ipc-tax-backtest-engine-break-even.png",
    "ipc-tax-backtest-engine-transport-floor.png",
    "ipc-tax-backtest-engine-the-verdict.png",
]


PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def read_png_size(path: Path) -> tuple[int, int]:
    with path.open("rb") as image:
        assert image.read(8) == PNG_SIGNATURE, f"{path.name}: invalid PNG signature"
        length = struct.unpack(">I", image.read(4))[0]
        chunk_type = image.read(4)
        assert chunk_type == b"IHDR", f"{path.name}: first chunk is not IHDR"
        assert length >= 8, f"{path.name}: invalid IHDR length {length}"
        width, height = struct.unpack(">II", image.read(8))
        return width, height


def main() -> None:
    for name in FILES:
        path = ROOT / "blog" / name
        width, height = read_png_size(path)
        assert width * 9 == height * 16, f"{name}: not 16:9"
        size = path.stat().st_size
        assert 500_000 <= size <= 3_000_000, f"{name}: suspicious file size {size}"
        print(f"OK {name}: {width}x{height}, {size:,} bytes")


if __name__ == "__main__":
    main()
