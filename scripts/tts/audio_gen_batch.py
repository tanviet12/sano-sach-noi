#!/usr/bin/env python3
"""
Đọc giọng nhiều file văn bản — nạp VieNeu v3 Turbo 1 lần, đọc tuần tự từng file.
Đây là script mà cmd/sano-docx2tts gọi khi --tts-mode vieneu.

Mỗi file <stem>.txt → <stem>_full.wav (48 kHz) ở thư mục hiện tại (CWD).
Dòng trắng / xuống dòng = ranh giới đoạn → v3 Turbo tự nghỉ dài hơn giữa câu.

Usage (chạy bằng python của venv VieNeu-TTS v3, PYTHONPATH chứa thư mục này):
    python audio_gen_batch.py chuong1.txt chuong2.txt chuong3.txt
    python audio_gen_batch.py --voice "Hải Đăng" chuong*.txt
    python audio_gen_batch.py --list-voices
"""

import argparse
import time
from pathlib import Path

# Import helper từ audio_gen.py (cũng cấu hình stdout UTF-8)
from audio_gen import DEFAULT_VOICE, load_tts, pick_voice, print_voices, synth_file


def process_file(tts, voice, input_file: str) -> dict:
    """Đọc giọng 1 file, trả về thống kê."""
    input_path = Path(input_file)
    if not input_path.exists():
        print(f"\n❌ File không tồn tại: {input_file}")
        return {"file": input_file, "ok": False, "error": "Not found"}

    output = Path(input_path.stem + "_full.wav")
    print(f"\n{'='*70}\n  📄 {input_file}\n{'='*70}")
    try:
        duration, compute = synth_file(tts, voice, input_path, output)
    except Exception as e:  # 1 file lỗi không làm hỏng cả lượt
        print(f"  ❌ Lỗi: {e}")
        return {"file": input_file, "ok": False, "error": str(e)[:60]}

    print(f"  ✨ {output.name}")
    print(f"     Audio: {duration:.1f}s ({duration/60:.1f} phút)")
    print(f"     Compute: {compute:.1f}s ({compute/60:.1f} phút)")
    print(f"     Speed ratio: {compute/duration:.2f}x realtime")
    return {"file": input_file, "ok": True, "duration_sec": duration, "compute_sec": compute}


def main():
    parser = argparse.ArgumentParser(description="Đọc giọng nhiều file văn bản bằng VieNeu-TTS v3 Turbo")
    parser.add_argument("files", nargs='*', help="Các file txt input")
    parser.add_argument("--voice", default=DEFAULT_VOICE, help=f'Tên giọng preset (mặc định: "{DEFAULT_VOICE}")')
    parser.add_argument("--list-voices", action="store_true", help="In danh sách giọng rồi thoát")
    args = parser.parse_args()

    print(f"🎤 Loading VieNeu v3 Turbo (1 lần duy nhất)...")
    load_start = time.time()
    tts = load_tts()
    if args.list_voices:
        print_voices(tts)
        return
    if not args.files:
        parser.error("cần ít nhất 1 file txt")
    voice = pick_voice(tts, args.voice)
    load_time = time.time() - load_start
    print(f"   Loaded trong {load_time:.0f}s")

    files = sorted(args.files)
    total_start = time.time()
    results = [process_file(tts, voice, f) for f in files]
    total_time = time.time() - total_start

    print(f"\n\n{'='*70}\n  📊 TỔNG KẾT\n{'='*70}")
    print(f"  Giọng: {voice} · Tổng thời gian: {total_time/60:.1f} phút · Load model: {load_time:.0f}s")
    print(f"\n  {'File':<30} {'Status':<10} {'Audio':<10} {'Compute':<10}")
    for r in results:
        if not r["ok"]:
            print(f"  {r['file']:<30} ❌ {r['error']}")
            continue
        print(f"  {r['file']:<30} {'✅ OK':<10} {r['duration_sec']/60:.1f}m{'':<6} {r['compute_sec']/60:.1f}m")

    total_audio = sum(r.get("duration_sec", 0) for r in results)
    total_compute = sum(r.get("compute_sec", 0) for r in results)
    print(f"\n  TỔNG: {total_audio/60:.1f} phút audio, {total_compute/60:.1f} phút compute")
    print(f"  Speed ratio: {total_compute/max(total_audio, 1):.2f}x realtime")
    failed = [r for r in results if not r["ok"]]
    if failed:
        print(f"  ⚠️  {len(failed)} file lỗi")
    print(f"{'='*70}\n")
    if failed:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
