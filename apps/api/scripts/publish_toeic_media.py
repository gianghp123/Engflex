"""Publish all hash-verified media for an existing TOEIC test (Part 1-7).

Publishes to toeic/<category>/<test>/{content.json,images/,audio/} and updates
media identities consistently. Review status, questions, answers and attempts
are preserved. Run from apps/api with uv.
"""

import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import re
import tempfile
from urllib.parse import urlsplit
from urllib.request import urlopen

from toeic_storage import BUCKET, assert_json_available, ensure_bucket, export_package, media_key, media_labels, minio_client, publish_json, relocate_content, test_prefix, verify_object


def media_assets(content):
    """Validate every graph reference and publish the entire asset manifest."""
    references = {content["audioAssetId"]}

    def visit(value):
        if isinstance(value, dict):
            for key, child in value.items():
                if key == "assetId" and isinstance(child, str):
                    references.add(child)
                else:
                    visit(child)
        elif isinstance(value, list):
            for child in value:
                visit(child)

    visit(content["parts"])
    assets = {asset["sourceId"]: asset for asset in content["assets"]}
    if len(assets) != len(content["assets"]):
        raise ValueError("Duplicate asset identity in manifest")
    missing = references - assets.keys()
    if missing:
        raise ValueError(f"Media references missing from manifest: {sorted(missing)}")
    return [assets[source] for source in sorted(assets)]


def local_asset(asset, source_root, crop_root):
    source = asset["sourceId"]
    digest = asset["sha256"].lower()
    if not re.fullmatch(r"[a-f0-9]{64}", digest):
        raise ValueError(f"Invalid SHA-256 for {source}")
    if urlsplit(source).scheme or source.startswith(("/", "\\")) or "\\" in source:
        raise ValueError(f"Expected a relative export asset identity: {source}")
    root = crop_root if source.startswith("crops/") else source_root
    if root is None:
        raise ValueError(f"Source root required for local asset: {source}")
    root = root.resolve()
    path = (root / source).resolve()
    if not path.is_relative_to(root) or not path.is_file():
        raise ValueError(f"Asset missing or outside its source root: {source}")
    with path.open("rb") as stream:
        actual = hashlib.file_digest(stream, "sha256").hexdigest()
    if actual != digest:
        raise ValueError(f"Source bytes do not match stored SHA-256: {source}")
    return path, f"exams/toeic/{digest}"


def remote_asset(asset, download_root):
    source = asset["sourceId"]
    digest = asset["sha256"].lower()
    url = urlsplit(source)
    if not re.fullmatch(r"[a-f0-9]{64}", digest):
        raise ValueError(f"Invalid SHA-256 for {source}")
    if url.scheme not in ("http", "https") or not url.netloc:
        raise ValueError(f"Invalid media URL: {source}")
    path = download_root / digest
    with urlopen(source, timeout=30) as response, path.open("wb") as output:
        if urlsplit(response.geturl()).scheme not in ("http", "https"):
            raise ValueError(f"Invalid redirect for {source}")
        actual = hashlib.sha256()
        while chunk := response.read(1024 * 1024):
            actual.update(chunk)
            output.write(chunk)
    if actual.hexdigest() != digest:
        raise ValueError(f"Downloaded bytes do not match stored SHA-256: {source}")
    return path, f"exams/toeic/{digest}"


def prepare(content, source_root, crop_root, download_root, category_slug, test_slug):
    # Validate every source before performing the first upload.
    from mutagen.mp3 import MP3
    from PIL import Image

    prepared = []
    labels = media_labels(content)
    duration = None
    for asset in media_assets(content):
        source = asset["sourceId"]
        if source.startswith("toeic/") and source_root is not None:
            path, key = local_asset(asset, source_root, crop_root)
        elif source.startswith(("toeic/", "exams/toeic/")):
            client = minio_client()
            bucket, key = (BUCKET, source.removeprefix("toeic/")) if source.startswith("toeic/") else (os.environ.get("MINIO_BUCKET") or "lesson-audio", source)
            if ".." in key or "\\" in key:
                raise ValueError("Unsafe stored media key")
            path = download_root / asset["sha256"].lower()
            client.fget_object(bucket, key, str(path))
            with path.open("rb") as stream:
                if hashlib.file_digest(stream, "sha256").hexdigest() != asset["sha256"].lower():
                    raise ValueError(f"Stored media hash mismatch: {source}")
        elif urlsplit(source).scheme in ("http", "https"):
            path, key = remote_asset(asset, download_root)
        else:
            path, key = local_asset(asset, source_root, crop_root)
        if source == content["audioAssetId"]:
            duration = MP3(path).info.length
            if not math.isfinite(duration) or duration <= 0:
                raise ValueError("Invalid MP3 duration")
            content_type = "audio/mpeg"
        else:
            with Image.open(path) as image:
                content_type = Image.MIME.get(image.format)
                image.verify()
            if content_type is None:
                raise ValueError(f"Unsupported image format: {source}")
        key = media_key(category_slug, test_slug, asset, content_type, label=labels[source])
        prepared.append((asset, path, key, content_type))
    for part in content["parts"]:
        for item in part["items"]:
            for field in ("itemAudio", "stimulusAudio"):
                audio = item.get(field)
                if audio and not (0 <= audio["startTime"] < audio["endTime"] <= duration):
                    raise ValueError(f"Audio range outside source duration: {item['id']}")
    return prepared


def publish_content(content, source_root=None, crop_root=None, dry_run=False, *, category_slug, test_slug, output_root=None):
    from dotenv import load_dotenv

    load_dotenv()
    with tempfile.TemporaryDirectory() as directory:
        test_prefix(category_slug, test_slug)
        prepared = prepare(content, source_root, crop_root, Path(directory), category_slug, test_slug)
        stored = relocate_content(content, {asset["sourceId"]: f"{BUCKET}/{key}" for asset, _, key, _ in prepared})
        summary = {"validatedAssets": len(prepared), "bytes": sum(path.stat().st_size for _, path, _, _ in prepared)}
        if output_root is not None:
            summary["outputDirectory"] = str(export_package(stored, prepared, output_root, category_slug, test_slug))
        if dry_run:
            return stored, summary
        client = minio_client()
        ensure_bucket(client)
        assert_json_available(client, category_slug, test_slug, stored, expected_content=content)
        for asset, path, key, content_type in prepared:
            client.fput_object(BUCKET, key, str(path), content_type=content_type)
            verify_object(client, BUCKET, key, asset["sha256"])
            print(f"Verified {asset['sourceId']} -> {key}")
        summary["contentKey"] = publish_json(client, category_slug, test_slug, stored, expected_content=content)
        return stored, summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--test-id", required=True)
    parser.add_argument("--source-root", type=Path, help="Required for relative non-crop assets")
    parser.add_argument("--crop-root", type=Path, help="Required for relative crop assets")
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--output-root", type=Path, help="Also export the MinIO bucket/key tree locally, including with --dry-run")
    args = parser.parse_args()
    from dotenv import load_dotenv
    import psycopg

    load_dotenv()
    with psycopg.connect(os.environ["DATABASE_URL"]) as connection:
        with connection.cursor() as cursor:
            cursor.execute("SELECT t.content,c.slug,t.slug FROM exam_tests t JOIN exam_categories c ON c.id=t.category_id WHERE t.id = %s AND c.type='toeic'", (args.test_id,))
            row = cursor.fetchone()
    if row is None:
        raise SystemExit("Exam test not found")
    stored, summary = publish_content(row[0], args.source_root, args.crop_root, args.dry_run, category_slug=row[1], test_slug=row[2], output_root=args.output_root)
    if not args.dry_run:
        from psycopg.types.json import Jsonb
        with psycopg.connect(os.environ["DATABASE_URL"]) as connection, connection.cursor() as cursor:
            cursor.execute("UPDATE exam_tests SET content=%s WHERE id=%s AND content=%s", (Jsonb(stored), args.test_id, Jsonb(row[0])))
            if cursor.rowcount != 1:
                raise ValueError("Exam content changed during publication; database left unchanged")
    print(json.dumps(summary))
    if not args.dry_run:
        print("Published TOEIC JSON/media and updated media identities; questions, answers and attempts unchanged")


if __name__ == "__main__":
    main()
