#!/usr/bin/env python3
"""
Đọc giọng 1 file văn bản bằng VieNeu-TTS v3 Turbo (ONNX/CPU, 48 kHz, giọng dựng sẵn).

v3 Turbo tự chia câu, tự chống "nói thêm" (babble guard) và tự chèn khoảng nghỉ
theo ranh giới: ngắt đoạn (dòng trắng / xuống dòng) 0,70s > hết câu 0,50s > ngắt
trong câu 0,30s. Vì vậy script chỉ đưa NGUYÊN văn bản vào `tts.infer()`.

Model nạp theo revision ghim trong versions.env (xem models.py) và chạy offline
sau lần tải đầu.

Usage (chạy bằng python của venv VieNeu-TTS v3):
    python audio_gen.py <input.txt> [voice_name] [output.wav]

Example:
    python audio_gen.py chuong1.txt "Thiện Minh" chuong1_full.wav
"""

import re
import sys
import time
from pathlib import Path

# In tiếng Việt + emoji an toàn trên console Windows (mặc định cp1252).
for _stream in (sys.stdout, sys.stderr):
    try:
        _stream.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, ValueError):
        pass

import models

DEFAULT_VOICE = "Thiện Minh"


def clean_text(text):
    """Bỏ dòng tiêu đề markdown (# ...) — không đọc."""
    return re.sub(r'^#+\s.*$', '', text, flags=re.MULTILINE).strip()


def load_tts():
    """Nạp VieNeu v3 Turbo với model ghim revision, không cần mạng sau lần tải đầu."""
    models.activate_offline()
    from vieneu import Vieneu
    return Vieneu(mode="v3turbo")


def voice_names(tts):
    return list(tts._preset_voices)


def print_voices(tts):
    print("\n📋 Giọng có sẵn:")
    for label, _ in tts.list_preset_voices():
        print(f"   • {label}")


def pick_voice(tts, voice_query=None):
    """Trả TÊN giọng preset: khớp đúng tên/bí danh, rồi khớp một phần (không phân
    biệt hoa thường). Không thấy → báo lỗi kèm danh sách (không tự đổi giọng)."""
    query = (voice_query or DEFAULT_VOICE).strip()
    name = tts.resolve_voice_name(query)
    if name is None:
        q = query.casefold()
        name = next((n for n in voice_names(tts) if q in n.casefold()), None)
    if name is None:
        print_voices(tts)
        raise SystemExit(f"❌ Không có giọng '{query}'.")
    print(f"✅ Giọng: {name}")
    return name


def synth_file(tts, voice, input_path, output_path):
    """Đọc cả file → WAV 48 kHz. Trả (thời lượng audio giây, thời gian đọc giây)."""
    import soundfile as sf

    text = clean_text(Path(input_path).read_text(encoding="utf-8"))
    if not text:
        raise ValueError(f"File rỗng: {input_path}")
    start = time.time()
    wav = tts.infer(text=text, voice=voice)
    compute = time.time() - start
    if wav is None or len(wav) == 0:
        raise RuntimeError(f"VieNeu không sinh được audio cho {input_path}")
    sf.write(str(output_path), wav, tts.sample_rate)
    return len(wav) / tts.sample_rate, compute


def main():
    if len(sys.argv) < 2:
        print('Usage: python audio_gen.py <input.txt> [voice_name] [output.wav]')
        print('Example: python audio_gen.py chuong1.txt "Thiện Minh" chuong1_full.wav')
        sys.exit(1)

    input_file = Path(sys.argv[1])
    voice_query = sys.argv[2] if len(sys.argv) > 2 else DEFAULT_VOICE
    output = Path(sys.argv[3]) if len(sys.argv) > 3 else Path(input_file.stem + "_full.wav")

    print("🎤 Loading VieNeu v3 Turbo...")
    t0 = time.time()
    tts = load_tts()
    voice = pick_voice(tts, voice_query)
    print(f"   Loaded trong {time.time() - t0:.0f}s")

    duration, compute = synth_file(tts, voice, input_file, output)
    print(f"\n✨ Done! Output: {output}")
    print(f"   Audio: {duration:.1f}s — đọc mất {compute:.1f}s ({compute / duration:.2f}x realtime)")


if __name__ == "__main__":
    main()
