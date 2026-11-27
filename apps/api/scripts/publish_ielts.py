"""Build and optionally publish a complete IELTS archive. Never writes to the DB.

Run via the uv-managed scripts/toeic_pipeline environment with --group study4.
Without --publish this only validates and exports a local package.
"""

import argparse
import hashlib
import io
import json
import shutil
import subprocess
import sys
from pathlib import Path

from dotenv import load_dotenv
from minio.error import S3Error

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))

from study4_ielts_content import build_content
from study4_mcp_export import media_format
from toeic_storage import minio_client, verify_object

BUCKET = "ielts"


def encode(content):
    return (json.dumps(content, ensure_ascii=False, indent=2, allow_nan=False) + "\n").encode("utf-8")


def prepare(export_path, *, category_slug, test_slug, sub_type):
    export_path = Path(export_path).resolve()
    data = json.loads(export_path.read_text(encoding="utf-8"))
    test_root = (export_path.parent / data["asset_root"]).resolve()
    content, media = build_content(data, test_root, category_slug=category_slug, test_slug=test_slug, sub_type=sub_type)
    keys = [item["key"] for item in media]
    if len(keys) != len(set(keys)):
        raise ValueError("Duplicate destination media key")
    for item in media:
        actual_type, extension = media_format(item["path"], "audio" if item["contentType"].startswith("audio/") else "image")
        if actual_type != item["contentType"] or not item["key"].endswith(extension):
            raise ValueError("Media bytes, declared content type and destination extension differ")
    subprocess.run(["go", "run", "./cmd/validate-ielts"], cwd=ROOT / "apps/api", input=encode(content), check=True)
    return content, media


def export_package(content, media, output_root, category_slug, test_slug):
    root = Path(output_root).resolve() / BUCKET
    target = root / category_slug / test_slug / "content.json"
    if target.exists() and json.loads(target.read_text(encoding="utf-8")) != content:
        raise ValueError("Existing local content differs; choose another output path")
    for item in media:
        destination = root / item["key"]
        destination.parent.mkdir(parents=True, exist_ok=True)
        if destination.exists():
            with destination.open("rb") as stream:
                if hashlib.file_digest(stream, "sha256").hexdigest() != item["sha256"]:
                    raise ValueError("Existing local media differs")
        else:
            shutil.copyfile(item["path"], destination)
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(encode(content))
    return target


def ensure_bucket(client):
    if not client.bucket_exists(BUCKET):
        client.make_bucket(BUCKET)
    desired = {"Version": "2012-10-17", "Statement": [{"Sid": "IeltsMediaRead", "Effect": "Allow",
        "Principal": {"AWS": ["*"]}, "Action": ["s3:GetObject"],
        "Resource": [f"arn:aws:s3:::{BUCKET}/*/*/audio/*", f"arn:aws:s3:::{BUCKET}/*/*/images/*"]}]}
    try:
        existing = json.loads(client.get_bucket_policy(BUCKET))
    except S3Error as error:
        if error.code != "NoSuchBucketPolicy":
            raise
        client.set_bucket_policy(BUCKET, json.dumps(desired))
        return
    for document in (existing, desired):
        for statement in document.get("Statement", []):
            for field in ("Action", "Resource"):
                statement[field] = sorted(statement[field])
    if existing != desired:
        raise ValueError("Existing IELTS bucket policy differs; refusing to overwrite it")


def publish(client, content, media, category_slug, test_slug):
    key = f"{category_slug}/{test_slug}/content.json"
    # Check all existing objects before writing any test content or media.
    if client.bucket_exists(BUCKET):
        try:
            response = client.get_object(BUCKET, key)
        except S3Error as error:
            if error.code != "NoSuchKey":
                raise
        else:
            try:
                if json.loads(response.read()) != content:
                    raise ValueError("Existing IELTS test content differs; refusing to overwrite")
            finally:
                response.close()
                response.release_conn()
        for item in media:
            try:
                client.stat_object(BUCKET, item["key"])
            except S3Error as error:
                if error.code != "NoSuchKey":
                    raise
            else:
                verify_object(client, BUCKET, item["key"], item["sha256"])
    ensure_bucket(client)
    for item in media:
        # Revalidate source bytes immediately before upload, not only at prepare time.
        with item["path"].open("rb") as stream:
            if hashlib.file_digest(stream, "sha256").hexdigest() != item["sha256"]:
                raise ValueError("Source media changed after validation")
            stream.seek(0)
            client.put_object(BUCKET, item["key"], stream, item["path"].stat().st_size, content_type=item["contentType"])
        verify_object(client, BUCKET, item["key"], item["sha256"])
    raw = encode(content)
    client.put_object(BUCKET, key, io.BytesIO(raw), len(raw), content_type="application/json")
    verify_object(client, BUCKET, key, hashlib.sha256(raw).hexdigest())
    return key


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--export", type=Path, required=True)
    parser.add_argument("--category-slug", required=True)
    parser.add_argument("--test-slug", required=True)
    parser.add_argument("--sub-type", choices=["academic", "general_training"], required=True)
    parser.add_argument("--output-root", type=Path, default=ROOT / "scripts/ielts_pipeline/output")
    parser.add_argument("--publish", action="store_true")
    args = parser.parse_args()
    content, media = prepare(args.export, category_slug=args.category_slug, test_slug=args.test_slug, sub_type=args.sub_type)
    target = export_package(content, media, args.output_root, args.category_slug, args.test_slug)
    receipt = {"status": "validated-local", "sourceExport": str(args.export.resolve()), "package": str(target.parent),
               "skill": content["skill"], "sections": len(content["sections"]),
               "questions": sum(len(g["questions"]) for s in content["sections"] for g in s["questionGroups"]),
               "media": len(media), "dbWritten": False, "sourceReview": "not-attested"}
    if args.publish:
        load_dotenv(ROOT / "apps/api/.env")
        receipt["object"] = BUCKET + "/" + publish(minio_client(), content, media, args.category_slug, args.test_slug)
        receipt["status"] = "published-verified"
    target.parent.with_suffix(".receipt.json").write_bytes(encode(receipt))
    print(json.dumps(receipt, ensure_ascii=True))


if __name__ == "__main__":
    main()
