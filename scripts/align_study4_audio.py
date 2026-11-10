"""Align Study4 TOEIC listening audio (Part 1-4) with question transcripts.

Extracts timestamps (startTime, endTime) for all 100 listening questions
using Faster-Whisper ASR and anchor-based alignment.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import re
import sys
from pathlib import Path

from study4_content import digest, exam_items, transcript_lines

# Add scripts directory to sys.path so we can import helpers
SCRIPTS_DIR = Path(__file__).resolve().parent
if str(SCRIPTS_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPTS_DIR))


def setup_cuda_dll_paths():
    """Register and preload nvidia CUDA/cuDNN DLL directories with Windows DLL loader."""
    if sys.platform != "win32":
        return
    import ctypes
    import site

    search_dirs = []
    try:
        for base in site.getsitepackages():
            search_dirs.append(Path(base))
    except Exception:
        pass
    # Also check local venv site-packages
    venv_site = SCRIPTS_DIR / "toeic_pipeline" / ".venv" / "Lib" / "site-packages"
    if venv_site.is_dir() and venv_site not in search_dirs:
        search_dirs.append(venv_site)

    for base in search_dirs:
        nvidia_dir = base / "nvidia"
        if nvidia_dir.is_dir():
            for sub in nvidia_dir.iterdir():
                bin_dir = sub / "bin"
                if bin_dir.is_dir():
                    os.environ["PATH"] = str(bin_dir) + os.pathsep + os.environ.get("PATH", "")
                    try:
                        os.add_dll_directory(str(bin_dir))
                    except Exception:
                        pass
                    # Preload DLLs so CTranslate2 LoadLibrary resolves them immediately
                    for dll_path in sorted(bin_dir.glob("*.dll")):
                        try:
                            ctypes.CDLL(str(dll_path))
                        except Exception:
                            pass


setup_cuda_dll_paths()

ONES = "zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen seventeen eighteen nineteen".split()
TENS = "zero ten twenty thirty forty fifty sixty seventy eighty ninety".split()


def spoken_number(n: int) -> str:
    if n == 100:
        return "one hundred"
    if n < 20:
        return ONES[n]
    tens_part = TENS[n // 10]
    ones_part = (" " + ONES[n % 10]) if (n % 10) else ""
    return tens_part + ones_part


def normalize_text(text: str) -> str:
    return re.sub(r"[^a-z0-9 ]", "", text.lower().replace("-", " "))


def find_q_numbers(node: dict) -> list[int]:
    if not isinstance(node, dict):
        return []
    res = []
    cls = node.get("class", "")
    if "question-number" in cls:
        for c in node.get("children", []):
            if "text" in c and c["text"].strip().isdigit():
                res.append(int(c["text"].strip()))
            for sc in c.get("children", []):
                if "text" in sc and sc["text"].strip().isdigit():
                    res.append(int(sc["text"].strip()))
    for c in node.get("children", []):
        res.extend(find_q_numbers(c))
    return res


def extract_all_text(node: dict) -> list[str]:
    if not isinstance(node, dict):
        return []
    res = []
    if "text" in node:
        res.append(node["text"])
    for c in node.get("children", []):
        res.extend(extract_all_text(c))
    return res


def find_transcripts(node: dict) -> list[str]:
    if not isinstance(node, dict):
        return []
    res = []
    cls = node.get("class", "")
    if "context-transcript" in cls:
        english, _ = transcript_lines(node)
        return [" ".join(english)] if english else []
    for c in node.get("children", []):
        res.extend(find_transcripts(c))
    return res


def get_listening_items_from_dom(content: list[dict]) -> list[dict]:
    items = []

    def scan_items(node: dict):
        if not isinstance(node, dict):
            return
        cls = node.get("class", "")
        if "question-item-wrapper" in cls or "question-group-wrapper" in cls:
            qs = find_q_numbers(node)
            trans = find_transcripts(node)
            if qs and max(qs) <= 100:
                item_id = f"item_q{qs[0]:02d}" if len(qs) == 1 else f"item_q{qs[0]}_{qs[-1]}"
                items.append(
                    {
                        "id": item_id,
                        "questions": sorted(list(set(qs))),
                        "transcript": trans[0] if trans else "",
                        "wrapper_class": cls,
                    }
                )
                return
        for c in node.get("children", []):
            scan_items(c)

    for root in content:
        scan_items(root)
    return items


def transcribe_audio(
    audio_path: Path,
    cache_path: Path | None = None,
    model_name: str = "base.en",
    device: str = "auto",
) -> list[dict]:
    with audio_path.open("rb") as stream:
        audio_hash = hashlib.file_digest(stream, "sha256").hexdigest()
    identity = {
        "audioSha256": audio_hash,
        "model": model_name,
        "language": "en",
        "wordTimestamps": True,
        "vadFilter": True,
        "beamSize": 5,
    }
    if cache_path and cache_path.exists():
        with cache_path.open("r", encoding="utf-8") as f:
            cached = json.load(f)
        if isinstance(cached, dict) and cached.get("identity") == identity:
            return cached["segments"]

    setup_cuda_dll_paths()
    from faster_whisper import WhisperModel

    # Determine device
    if device == "auto":
        use_cuda = False
        try:
            import numpy as np

            test_m = WhisperModel("tiny.en", device="cuda", compute_type="float16")
            segs, _ = test_m.transcribe(np.zeros(16000, dtype=np.float32))
            list(segs)
            use_cuda = True
        except Exception:
            use_cuda = False

        if use_cuda:
            device = "cuda"
            compute_type = "float16"
        else:
            device = "cpu"
            compute_type = "int8"
    else:
        compute_type = "float16" if device == "cuda" else "int8"

    print(
        f"Transcribing {audio_path.name} with Faster-Whisper ({model_name}, device={device}, compute_type={compute_type}, beam_size=5, word_timestamps=True)..."
    )
    model = WhisperModel(model_name, device=device, compute_type=compute_type, cpu_threads=8)
    seg_iter, info = model.transcribe(
        str(audio_path),
        language="en",
        word_timestamps=True,
        vad_filter=True,
        beam_size=5,
    )

    segments = []
    for s in seg_iter:
        seg = {
            "start": round(s.start, 2),
            "end": round(s.end, 2),
            "text": s.text.strip(),
        }
        if s.words:
            seg["words"] = [
                {
                    "word": w.word,
                    "start": round(w.start, 2),
                    "end": round(w.end, 2),
                    "probability": round(w.probability, 2),
                }
                for w in s.words
            ]
        segments.append(seg)

    print(f"ASR complete: {len(segments)} segments transcribed (audio duration: {info.duration:.1f}s).")
    if cache_path:
        cache_path.parent.mkdir(parents=True, exist_ok=True)
        with cache_path.open("w", encoding="utf-8") as f:
            json.dump({"identity": identity, "segments": segments}, f, ensure_ascii=False, indent=2)

    return segments


def align_listening_items(
    items: list[dict],
    segments: list[dict],
    audio_asset_id: str,
) -> list[dict]:
    # Search a word timeline so markers inside a segment have exact boundaries.
    # Segment-only caches may locate markers only at the start of a segment.
    units = []
    for segment in segments:
        words = segment.get("words") or [{"word": segment["text"], "start": segment["start"], "end": segment["end"]}]
        units.extend({**word, "segmentFirst": index == 0} for index, word in enumerate(words))
    timeline, spans = "", []
    for unit in units:
        value = normalize_text(unit["word"]).strip()
        if not value:
            continue
        offset = len(timeline)
        timeline += value + " "
        spans.append((offset, len(timeline), unit))

    def timestamp(position):
        for start, end, unit in spans:
            if start <= position < end:
                # Never guess a sub-segment timestamp without word timings.
                return unit["start"] if position == start else None
        return None

    anchors = {}
    cursor = 0
    for item in items:
        number = item["questions"][0]
        token = "number" if number <= 31 else r"questions?"
        alternatives = "|".join(re.escape(value) for value in (str(number), spoken_number(number)))
        pattern = rf"(?<!question )(?<!marked )\b{token}\s+(?:{alternatives})\b"
        if number > 31:
            last = item["questions"][-1]
            ends = "|".join(re.escape(value) for value in (str(last), spoken_number(last)))
            # Whisper sometimes omits 'Questions', but retains both group
            # numbers and the explicit 'refer to the following' announcement.
            pattern += rf"|\b(?:{alternatives})\s+(?:(?:through|to)\s+)?(?:{ends})\s+refer to the following\b"
        candidates = [(cursor + match.start(), cursor + match.end()) for match in re.finditer(pattern, timeline[cursor:])]
        if number <= 31:
            # An isolated numbered ASR segment is also an explicit marker.
            # Only accept punctuation-marked segment starts, not numbers
            # occurring inside spoken options or picture instructions.
            for start, end, unit in spans:
                if start >= cursor and unit.get("segmentFirst") and re.fullmatch(rf"\s*(?:{alternatives})[.:]\s*", unit["word"], re.I):
                    candidates.append((start, end))
        for position, end in sorted(candidates):
            time = timestamp(position)
            if time is not None:
                anchors[item["id"]] = (position, time)
                cursor = end
                break

    aligned = []
    for idx, item in enumerate(items):
        anchor = anchors.get(item["id"])
        next_anchor = anchors.get(items[idx + 1]["id"]) if idx + 1 < len(items) else None
        start_time, end_time = None, None
        if (
            anchor
            and anchor[1] is not None
            and (idx + 1 == len(items) or next_anchor and next_anchor[1] is not None)
        ):
            start_time = anchor[1]
            limit = next_anchor[0] if next_anchor else len(timeline)
            section = timeline[anchor[0] : limit]
            boundary = re.search(
                r"\b(?:this is the end|end of the listening|part (?:[1-4]|one|two|three|four)\b|go on to the next page)",
                section,
            )
            if boundary:
                limit = anchor[0] + boundary.start()
                if timestamp(limit) is None:
                    start_time = None
            previous = [unit for _, end, unit in spans if end <= limit and unit["end"] > (start_time or 0)]
            if start_time is not None and previous:
                end_time = previous[-1]["end"]
            if end_time is None or end_time <= (start_time or 0):
                start_time, end_time = None, None
        entry = {
            "id": item["id"],
            "questions": item["questions"],
            "transcript": item["transcript"],
            "audioRange": {
                "assetId": audio_asset_id,
                "startTime": start_time,
                "endTime": end_time,
            },
        }
        aligned.append(entry)

    return aligned


def process_test_alignment(
    test_id: str,
    crawl_root: Path | str = "docs/crawl",
    model_name: str = "base.en",
    run_id: str | None = None,
    device: str = "auto",
) -> dict:
    test_dir = Path(crawl_root).resolve() / "study4" / "tests" / test_id
    if not test_dir.exists():
        raise FileNotFoundError(f"Test directory not found: {test_dir}")

    runs_dir = test_dir / "runs"
    runs = sorted([r for r in runs_dir.iterdir() if r.is_dir()])
    if not runs:
        raise FileNotFoundError(f"No run folders found under {runs_dir}")
    latest_run = next((r for r in runs if r.name == run_id), None) if run_id else runs[-1]
    if latest_run is None:
        raise ValueError("Selected run does not exist")
    export_path = latest_run / "export.json"

    with export_path.open("r", encoding="utf-8") as f:
        export_data = json.load(f)

    exam_items(export_data)

    items = get_listening_items_from_dom(export_data.get("content", []))
    if not items:
        raise ValueError(f"No listening items found in {export_path}")

    audio_assets = [a for a in export_data["assets"] if a.get("kind") == "audio"]
    if len(audio_assets) != 1 or not re.fullmatch(r"[a-f0-9]{64}", audio_assets[0].get("sha256", "")):
        raise ValueError("Expected one hash-verified full listening audio asset")
    audio = audio_assets[0]
    audio_key = audio["sourceId"]
    cache_dir = test_dir / "cache"
    cache_dir.mkdir(parents=True, exist_ok=True)
    local_audio = cache_dir / f"{audio['sha256']}.mp3"
    asr_cache = cache_dir / f"{audio['sha256']}.{digest(model_name)[:12]}.asr.json"

    def verify(path):
        with path.open("rb") as stream:
            return hashlib.file_digest(stream, "sha256").hexdigest() == audio["sha256"]

    if not local_audio.exists():
        import shutil

        candidates = [cache_dir / "audio_full.mp3"]
        if audio.get("localPath"):
            candidate = (test_dir / audio["localPath"]).resolve()
            if not candidate.is_relative_to(test_dir):
                raise ValueError("Audio localPath escapes test directory")
            candidates.insert(0, candidate)
        existing = next((p for p in candidates if p.is_file() and verify(p)), None)
        if existing:
            shutil.copyfile(existing, local_audio)
        else:
            from study4_mcp_export import get_minio_client_and_bucket

            if audio.get("status") != "uploaded" or not audio.get("minioBucket") or not audio.get("minioKey"):
                raise ValueError("No verified local audio or uploaded object available")
            client, bucket = get_minio_client_and_bucket(
                argparse.Namespace(minio_bucket=audio["minioBucket"])
            )
            print(f"Fetching {audio['minioKey']} from MinIO [{bucket}]...")
            temporary = local_audio.with_suffix(".download")
            try:
                client.fget_object(bucket, audio["minioKey"], str(temporary))
                if not verify(temporary):
                    raise ValueError("Downloaded audio SHA-256 does not match manifest")
                temporary.replace(local_audio)
            finally:
                temporary.unlink(missing_ok=True)
    if not verify(local_audio):
        raise ValueError("Cached audio SHA-256 does not match manifest")

    from mutagen.mp3 import MP3

    duration = MP3(local_audio).info.length
    if not math.isfinite(duration) or duration <= 0:
        raise ValueError("Invalid listening audio duration")

    segments = transcribe_audio(local_audio, cache_path=asr_cache, model_name=model_name, device=device)
    aligned_items = align_listening_items(items, segments, audio_key)

    # Save alignment artifact in run folder
    alignment_file = latest_run / "audio-alignment.json"
    alignment_summary = {
        "testId": test_id,
        "runId": latest_run.name,
        "audioKey": audio_key,
        "audioSha256": audio["sha256"],
        "audioDuration": duration,
        "exportHash": digest(export_data["content"]),
        "asrHash": digest(segments),
        "model": model_name,
        "itemCount": len(aligned_items),
        "alignedCount": sum(1 for it in aligned_items if it["audioRange"]["startTime"] is not None),
        "items": aligned_items,
    }
    summary_errors = []
    previous_end = 0
    if len(aligned_items) != 54:
        summary_errors.append("Expected 54 listening items")
    for item in aligned_items:
        start, end = item["audioRange"]["startTime"], item["audioRange"]["endTime"]
        if start is None or end is None or not 0 <= previous_end <= start < end <= duration:
            summary_errors.append(f"Invalid or unresolved audio range: {item['id']}")
        else:
            previous_end = end
    alignment_summary["issues"] = summary_errors
    alignment_summary["status"] = "incomplete" if summary_errors else "complete"
    with alignment_file.open("w", encoding="utf-8") as f:
        json.dump(alignment_summary, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(
        f"Saved audio alignment to {alignment_file} ({alignment_summary['alignedCount']}/{alignment_summary['itemCount']} aligned)"
    )

    # Inject into export.json
    export_data["listening_alignments"] = "audio-alignment.json"
    export_data["audio_ranges"] = {item["id"]: item["audioRange"] for item in aligned_items}
    with export_path.open("w", encoding="utf-8") as f:
        json.dump(export_data, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"Updated {export_path} with listening audio ranges.")

    if summary_errors:
        raise ValueError("Alignment incomplete; see audio-alignment.json: " + "; ".join(summary_errors))
    return alignment_summary


def main():
    parser = argparse.ArgumentParser(description="Align Study4 listening audio with question transcripts")
    parser.add_argument("--test-id", help="Single test ID to align (e.g. 6852)")
    parser.add_argument("--all", action="store_true", help="Align all crawled tests under crawl_root")
    parser.add_argument(
        "--crawl-root", default="docs/crawl", help="Root crawl directory (default: docs/crawl)"
    )
    parser.add_argument("--model", default="base.en", help="Whisper model name (default: base.en)")
    parser.add_argument("--run-id", help="Select an exact crawl run instead of the latest run")
    parser.add_argument("--device", choices=["auto", "cpu", "cuda"], default="auto")
    args = parser.parse_args()

    crawl_root = Path(args.crawl_root).resolve()
    tests_dir = crawl_root / "study4" / "tests"

    if args.all:
        test_ids = [d.name for d in tests_dir.iterdir() if d.is_dir()]
    elif args.test_id:
        test_ids = [args.test_id]
    else:
        parser.print_help()
        sys.exit(1)

    failed = False
    for tid in test_ids:
        print(f"\n=================== Aligning Test {tid} ===================")
        try:
            summary = process_test_alignment(
                tid, crawl_root=crawl_root, model_name=args.model, run_id=args.run_id, device=args.device
            )
            print(f"Test {tid} alignment successful: {summary['alignedCount']}/{summary['itemCount']} items.")
        except Exception as e:
            failed = True
            print(f"Test {tid} alignment failed: {e}", file=sys.stderr)
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
