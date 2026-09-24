#!/usr/bin/env python3
"""
Model VieNeu-TTS v3 Turbo ghim revision (đọc từ versions.env cùng thư mục).

Vì sao cần: VieNeu gọi `hf_hub_download` không kèm revision → mỗi lần nạp đều hỏi
Hugging Face bản "main" mới nhất (cần mạng, và model đổi âm thầm nếu upstream cập
nhật). Module này:

1. `fetch()`  — tải đúng revision ghim vào cache HF chuẩn, trỏ refs/main về
                revision đó. Cần mạng, chạy 1 lần. Không cần HF_TOKEN.
2. `activate_offline()` — gọi TRƯỚC khi import vieneu: nếu cache đã đủ file ghim
                thì bật HF_HUB_OFFLINE=1 (đọc không cần mạng, luôn đúng bản ghim);
                thiếu thì tự fetch trước.

3. `verify()` — kiểm SHA256 nội dung từng file ghim khớp models.sha256 (cùng thư
                mục). `fetch()` tự gọi sau khi tải; lệch là báo lỗi.

Dùng dòng lệnh:
    python models.py fetch    # tải model ghim + kiểm SHA256, in dung lượng
    python models.py check    # kiểm tra cache đã đủ file chưa (exit 1 nếu thiếu)
    python models.py verify   # kiểm SHA256 từng file (exit 1 nếu lệch/thiếu)
"""

import hashlib
import os
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
VERSIONS_FILE = HERE / "versions.env"
SHA256_FILE = HERE / "models.sha256"


def read_pins(path=VERSIONS_FILE):
    pins = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        pins[k.strip()] = v.strip()
    return pins


def model_specs(pins=None):
    """[(repo_id, revision, [files])] cho backbone + codec."""
    pins = pins or read_pins()
    specs = []
    for key in ("BACKBONE", "CODEC"):
        repo = pins[f"HF_{key}_REPO"]
        rev = pins[f"HF_{key}_REVISION"]
        files = [f.strip() for f in pins[f"HF_{key}_FILES"].split(",") if f.strip()]
        specs.append((repo, rev, files))
    return specs


def hub_cache_dir():
    """Thư mục cache HF (cùng quy tắc huggingface_hub, không cần import nó)."""
    if os.environ.get("HF_HUB_CACHE"):
        return Path(os.environ["HF_HUB_CACHE"])
    if os.environ.get("HF_HOME"):
        return Path(os.environ["HF_HOME"]) / "hub"
    xdg = os.environ.get("XDG_CACHE_HOME")
    base = Path(xdg) if xdg else Path.home() / ".cache"
    return base / "huggingface" / "hub"


def repo_dir(repo_id):
    return hub_cache_dir() / ("models--" + repo_id.replace("/", "--"))


def missing_files(specs=None):
    specs = specs or model_specs()
    missing = []
    for repo, rev, files in specs:
        snap = repo_dir(repo) / "snapshots" / rev
        for f in files:
            if not (snap / f).exists():
                missing.append(f"{repo}@{rev[:8]}/{f}")
    return missing


def read_sha256(path=SHA256_FILE):
    """{"<repo>@<revision>/<file>": sha256} từ models.sha256 (định dạng sha256sum)."""
    out = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        digest, name = line.split(None, 1)
        out[name.strip()] = digest.lower()
    return out


def _sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def verify(specs=None):
    """Trả danh sách file lệch/thiếu so với models.sha256 (rỗng = đúng hết)."""
    specs = specs or model_specs()
    want = read_sha256()
    bad = []
    for repo, rev, files in specs:
        snap = repo_dir(repo) / "snapshots" / rev
        for f in files:
            key = f"{repo}@{rev}/{f}"
            path = snap / f
            if key not in want:
                bad.append(f"{key} (models.sha256 không có dòng này)")
            elif not path.exists():
                bad.append(f"{key} (thiếu file)")
            elif _sha256_file(path) != want[key]:
                bad.append(f"{key} (SHA256 lệch)")
    return bad


def pin_refs(specs=None):
    """Trỏ refs/main của từng repo về revision ghim — để các lệnh hf_hub_download
    không kèm revision (bên trong VieNeu) chạy offline vẫn lấy đúng bản ghim."""
    specs = specs or model_specs()
    for repo, rev, _ in specs:
        refs = repo_dir(repo) / "refs"
        refs.mkdir(parents=True, exist_ok=True)
        (refs / "main").write_text(rev, encoding="utf-8")


def _dir_size(path):
    total = 0
    for p in Path(path).rglob("*"):
        try:
            if p.is_file() and not p.is_symlink():
                total += p.stat().st_size
        except OSError:
            pass
    return total


def fetch():
    """Tải model ghim (cần mạng). Trả tổng dung lượng cache của các repo (byte)."""
    from huggingface_hub import hf_hub_download

    specs = model_specs()
    for repo, rev, files in specs:
        for f in files:
            path = hf_hub_download(repo_id=repo, filename=f, revision=rev)
            print(f"   ✓ {repo}@{rev[:8]}/{f} → {path}", flush=True)
    pin_refs(specs)
    bad = verify(specs)
    if bad:
        raise RuntimeError("Model tải về không khớp SHA256 ghim:\n  " + "\n  ".join(bad))
    print("   ✓ SHA256 khớp models.sha256", flush=True)
    total = sum(_dir_size(repo_dir(repo)) for repo, _, _ in specs)
    print(f"   Model ghim: {total / 1024 / 1024:.1f} MB trong {hub_cache_dir()}")
    return total


def activate_offline():
    """Gọi TRƯỚC `import vieneu`. Đảm bảo model ghim có sẵn rồi bật chế độ offline."""
    specs = model_specs()
    if missing_files(specs):
        print("📥 Chưa có model ghim trong cache — tải lần đầu (cần mạng)...", flush=True)
        # Chạy tiến trình con: huggingface_hub đọc HF_HUB_OFFLINE lúc import, nên
        # tiến trình hiện tại chưa được import nó ở chế độ online.
        subprocess.run([sys.executable, str(Path(__file__).resolve()), "fetch"], check=True)
        still = missing_files(specs)
        if still:
            raise RuntimeError("Tải model ghim chưa đủ: " + ", ".join(still))
    pin_refs(specs)
    os.environ["HF_HUB_OFFLINE"] = "1"


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "check"
    if cmd == "fetch":
        fetch()
    elif cmd == "check":
        miss = missing_files()
        if miss:
            print("Thiếu:\n  " + "\n  ".join(miss))
            sys.exit(1)
        print(f"Đủ model ghim trong {hub_cache_dir()}")
    elif cmd == "verify":
        bad = verify()
        if bad:
            print("Lệch/thiếu:\n  " + "\n  ".join(bad))
            sys.exit(1)
        print(f"SHA256 khớp models.sha256 trong {hub_cache_dir()}")
    else:
        print("Usage: python models.py fetch|check|verify")
        sys.exit(2)


if __name__ == "__main__":
    main()
