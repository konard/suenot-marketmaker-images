import struct
from pathlib import Path

EXPECTED = [
    "blog/deeplob-deep-learning-order-book.png",
    "blog/deeplob-deep-learning-order-book-lob-structure.png",
    "blog/deeplob-deep-learning-order-book-architecture.png",
    "blog/deeplob-deep-learning-order-book-features.png",
    "blog/deeplob-deep-learning-order-book-replication.png",
    "blog/deeplob-deep-learning-order-book-landscape.png",
    "blog/spread-modeling-machine-learning.png",
    "blog/spread-modeling-machine-learning-components.png",
    "blog/spread-modeling-machine-learning-roll.png",
    "blog/spread-modeling-machine-learning-features.png",
    "blog/spread-modeling-machine-learning-quoting.png",
    "blog/spread-modeling-machine-learning-evaluation.png",
    "blog/temporal-fusion-transformer-trading.png",
    "blog/temporal-fusion-transformer-trading-architecture.png",
    "blog/temporal-fusion-transformer-trading-comparison.png",
    "blog/temporal-fusion-transformer-trading-interpretability.png",
    "blog/temporal-fusion-transformer-trading-pipeline.png",
    "blog/conformal-prediction-trading.png",
    "blog/conformal-prediction-trading-nonconformity.png",
    "blog/conformal-prediction-trading-split.png",
    "blog/conformal-prediction-trading-non-exchangeability.png",
    "blog/conformal-prediction-trading-adaptive.png",
    "blog/conformal-prediction-trading-position-sizing.png",
]

def png_size(path):
    data = Path(path).read_bytes()
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise AssertionError(f"{path}: not a PNG")
    if data[12:16] != b"IHDR":
        raise AssertionError(f"{path}: missing IHDR")
    return struct.unpack(">II", data[16:24])

for name in EXPECTED:
    p = Path(name)
    if not p.exists():
        raise AssertionError(f"missing {name}")
    width, height = png_size(p)
    ratio = width / height
    if abs(ratio - (16 / 9)) > 0.002:
        raise AssertionError(f"{name}: expected near 16:9, got {width}x{height}")
    print(f"OK {name} {width}x{height} {p.stat().st_size} bytes")
print(f"Validated {len(EXPECTED)} PNG files")
