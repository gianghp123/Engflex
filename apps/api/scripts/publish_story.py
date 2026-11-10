"""Publish a licensed local story package to the Engflex stories bucket.

Run from apps/api:
    uv run --with minio --with python-dotenv python scripts/publish_story.py manifest.json

The manifest uses the Stories response field names. audioUrl, thumbnailUrl,
and sentence imageUrl values are paths relative to the manifest directory.
The command validates and prints the published manifest unless --publish is set.
"""

import argparse
import io
import json
import math
import mimetypes
import re
from pathlib import Path

from dotenv import load_dotenv

from toeic_storage import minio_client

BUCKET = "stories"


def build_package(manifest_path):
    story = json.loads(manifest_path.read_text(encoding="utf-8"))
    slug = story["slug"]
    if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", slug):
        raise ValueError("invalid story slug")
    duration = story["duration"]
    if not math.isfinite(duration) or duration <= 0:
        raise ValueError("duration must be positive and finite")
    if not story["title"].strip() or not story["sentences"]:
        raise ValueError("title and sentences are required")

    directory = manifest_path.resolve().parent
    prefix = slug
    assets = {}
    scene_names = {}

    def media_url(relative_path, expected_type, name=None):
        path = (directory / relative_path).resolve()
        if not path.is_relative_to(directory) or not path.is_file():
            raise ValueError(f"media must be a local file inside the package: {relative_path}")
        content_type, _ = mimetypes.guess_type(path)
        if not content_type or not content_type.startswith(f"{expected_type}/"):
            raise ValueError(f"expected a supported {expected_type} file: {relative_path}")
        if name is None:
            name = scene_names.setdefault(path, f"scene-{len(scene_names) + 1:02d}")
        key = f"{prefix}/assets/{name}{path.suffix.lower()}"
        assets[key] = (path, content_type)
        return f"/api/media/{BUCKET}/{key}"

    story["audioUrl"] = media_url(story["audioUrl"], "audio", "narration")
    if story.get("thumbnailUrl"):
        story["thumbnailUrl"] = media_url(story["thumbnailUrl"], "image", "cover")

    previous_end = 0
    for sequence, sentence in enumerate(story["sentences"], start=1):
        start = sentence["startTimestamp"]
        end = sentence["endTimestamp"]
        if not previous_end <= start < end <= duration:
            raise ValueError(f"invalid timestamps for sentence {sequence}")
        if not sentence["englishText"].strip():
            raise ValueError(f"empty English text for sentence {sequence}")
        sentence["sequence"] = sequence
        sentence.setdefault("vietnameseText", "")
        if sentence.get("imageUrl"):
            sentence["imageUrl"] = media_url(sentence["imageUrl"], "image")
        previous_end = end

    data = json.dumps(story, ensure_ascii=False, indent=2, allow_nan=False).encode("utf-8")
    return f"{prefix}/story.json", data, assets


def publish_package(key, data, assets):
    load_dotenv(Path(__file__).resolve().parents[1] / ".env")
    client = minio_client()
    for asset_key, (path, content_type) in assets.items():
        client.fput_object(BUCKET, asset_key, str(path), content_type=content_type)
    client.put_object(BUCKET, key, io.BytesIO(data), len(data), content_type="application/json")
    print(f"Published {BUCKET}/{key} with {len(assets)} media objects")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("manifest", type=Path)
    parser.add_argument("--publish", action="store_true", help="Upload assets and JSON to MinIO")
    args = parser.parse_args()
    key, data, assets = build_package(args.manifest)
    if args.publish:
        publish_package(key, data, assets)
    else:
        print(data.decode("utf-8"))


if __name__ == "__main__":
    main()
