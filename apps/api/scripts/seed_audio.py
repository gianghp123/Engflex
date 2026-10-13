"""Generate lesson listening/speaking audio with edge-tts and upload to MinIO.

The database seed is the single source of truth: this script reads
lesson_activities.config, synthesizes each ScriptLine / SpeakingItem with its
declared voice, joins MP3 frames, and PUTs to the object key the config
already references. Re-runs overwrite (idempotent).

Usage: make seed-audio   (uv run --with edge-tts --with minio --with "psycopg[binary]" --with python-dotenv)
       add --dry-run to validate configs and list planned uploads only.
"""

import argparse
import asyncio
import io
import json
import os
import sys
from pathlib import Path

import edge_tts
from dotenv import load_dotenv
from minio import Minio


def load_env() -> None:
    load_dotenv(Path(__file__).resolve().parents[1] / ".env")


def db_url() -> str:
    url = os.environ.get("DATABASE_URL")
    if not url:
        sys.exit("DATABASE_URL is not set (apps/api/.env or environment)")
    return url


def minio_client() -> Minio:
    endpoint = os.environ.get("MINIO_ENDPOINT", "http://localhost:9000")
    return Minio(
        endpoint.replace("http://", "").replace("https://", ""),
        access_key=os.environ["MINIO_ROOT_USER"],
        secret_key=os.environ["MINIO_ROOT_PASSWORD"],
        secure=endpoint.startswith("https://"),
    )


async def mp3_bytes(voice: str, text: str) -> bytes:
    out = io.BytesIO()
    communicate = edge_tts.Communicate(text, voice)
    async for chunk in communicate.stream():
        if chunk["type"] == "audio":
            out.write(chunk["data"])
    data = out.getvalue()
    if not data:
        raise ValueError(f"edge-tts produced no audio for voice={voice!r}")
    return data


def join(items: list[bytes]) -> bytes:
    """Concatenate MP3 frame streams (same codec/rate per edge-tts run)."""
    return b"".join(items)


def validate_listening(cfg: dict) -> list[str]:
    errs = []
    if not cfg.get("audioKey"):
        errs.append("listening: missing audioKey")
    script = cfg.get("script") or []
    if not script:
        errs.append("listening: empty script")
    for i, line in enumerate(script):
        if not line.get("text") or not line.get("voice"):
            errs.append(f"listening: script[{i}] needs text and voice")
    if not (3 <= len(cfg.get("questions") or []) <= 5):
        errs.append("listening: need 3-5 questions")
    return errs


def validate_speaking(cfg: dict) -> list[str]:
    errs = []
    items = cfg.get("items") or []
    if not (3 <= len(items) <= 5):
        errs.append("speaking: need 3-5 items")
    for i, item in enumerate(items):
        if not item.get("text") or not item.get("modelAudioKey") or not item.get("modelVoice"):
            errs.append(f"speaking: item[{i}] needs text/modelAudioKey/modelVoice")
    return errs


async def main(dry_run: bool, activity_id: str | None) -> None:
    load_env()
    import psycopg

    with psycopg.connect(db_url()) as conn:
        if activity_id:
            rows = conn.execute(
                "SELECT id, type, config FROM lesson_activities "
                "WHERE type IN ('listening','speaking') AND id = %s",
                (activity_id,),
            ).fetchall()
        else:
            rows = conn.execute(
                "SELECT id, type, config FROM lesson_activities "
                "WHERE type IN ('listening','speaking') ORDER BY id"
            ).fetchall()

    if not rows:
        if activity_id:
            sys.exit(f"activity {activity_id}: not found, or not type listening/speaking")
        sys.exit("no listening/speaking activities found — run migrations first")

    uploads: list[tuple[str, bytes]] = []
    planned: list[str] = []
    for activity_id, kind, raw in rows:
        # psycopg v3 parses jsonb to dict already; older drivers return str.
        cfg = raw if isinstance(raw, dict) else json.loads(raw)
        errs = validate_listening(cfg) if kind == "listening" else validate_speaking(cfg)
        if errs:
            sys.exit(f"activity {activity_id}: " + "; ".join(errs))
        if dry_run:
            if kind == "listening":
                planned.append(cfg["audioKey"])
            else:
                planned.extend(item["modelAudioKey"] for item in cfg["items"])
            continue
        if kind == "listening":
            parts = [await mp3_bytes(line["voice"], line["text"]) for line in cfg["script"]]
            uploads.append((cfg["audioKey"], join(parts)))
        else:
            for item in cfg["items"]:
                uploads.append((item["modelAudioKey"], await mp3_bytes(item["modelVoice"], item["text"])))

    if dry_run:
        print(f"{len(planned)} object(s) planned from {len(rows)} activities")
        for key in planned:
            print(f"  {key}")
        return

    print(f"{len(uploads)} object(s) planned from {len(rows)} activities")
    for key, data in uploads:
        print(f"  {key} ({len(data)} bytes)")

    client = minio_client()
    bucket = os.environ.get("MINIO_BUCKET", "lesson-audio")
    if not client.bucket_exists(bucket):
        sys.exit(f"bucket {bucket!r} does not exist — run `docker compose up -d minio-init`")
    for key, data in uploads:
        client.put_object(bucket, key, io.BytesIO(data), length=len(data), content_type="audio/mpeg")
    print(f"uploaded {len(uploads)} object(s) to {bucket}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--dry-run", action="store_true", help="validate configs, list uploads, no synthesis")
    parser.add_argument("--activity-id", default=None, help="synthesize only this lesson_activities.id (default: all)")
    args = parser.parse_args()
    asyncio.run(main(args.dry_run, args.activity_id))
