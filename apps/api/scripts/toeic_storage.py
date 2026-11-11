"""TOEIC bucket layout shared by publishing and storage maintenance."""

import hashlib
import io
import json
import os
import re
import shutil
from urllib.parse import urlsplit

BUCKET = "toeic"
SLUG = re.compile(r"[a-z0-9]+(?:-[a-z0-9]+)*")
EXTENSIONS = {"audio/mpeg": "mp3", "image/jpeg": "jpg", "image/png": "png", "image/webp": "webp", "image/gif": "gif"}


def test_prefix(category_slug, test_slug):
    if not all(isinstance(s, str) and SLUG.fullmatch(s) for s in (category_slug, test_slug)):
        raise ValueError("Category and test slugs must be lowercase hyphenated slugs")
    return f"{category_slug}/{test_slug}"


def media_labels(content):
    """Name assets by their first question-group use, independently of old keys."""
    labels = {content["audioAssetId"]: "audio"}
    for part in content["parts"]:
        for item in part["items"]:
            references = []
            def collect(value):
                if isinstance(value, dict):
                    for key, child in sorted(value.items()):
                        if key == "assetId" and isinstance(child, str):
                            if child != content["audioAssetId"] and child not in references:
                                references.append(child)
                        else:
                            collect(child)
                elif isinstance(value, list):
                    for child in value:
                        collect(child)
            collect(item)
            numbers = sorted(q["questionNumber"] for q in item["questions"])
            scope = f"part-{part['partNumber']}-q{numbers[0]:02d}"
            if len(numbers) > 1:
                scope += f"-{numbers[-1]:02d}"
            for index, source in enumerate(references, 1):
                label = scope + (f"-image-{index:02d}" if len(references) > 1 else "")
                labels.setdefault(source, label)
    for index, asset in enumerate(content["assets"], 1):
        labels.setdefault(asset["sourceId"], f"image-{index:02d}")
    if len(set(labels.values())) != len(labels):
        raise ValueError("Media names collide; check question-group references")
    return labels


def media_key(category_slug, test_slug, asset, content_type, *, label):
    digest = asset["sha256"].lower()
    if not re.fullmatch(r"[a-f0-9]{64}", digest) or content_type not in EXTENSIONS or not SLUG.fullmatch(label):
        raise ValueError("Invalid media digest or content type")
    folder = "audio" if content_type.startswith("audio/") else "images"
    return f"{test_prefix(category_slug, test_slug)}/{folder}/{test_slug}-{label}.{EXTENSIONS[content_type]}"


def export_package(content, prepared, output_root, category_slug, test_slug):
    """Write the same bucket/key tree locally, using only verified source bytes."""
    root = output_root.resolve() / BUCKET
    for asset, path, key, _ in prepared:
        destination = (root / key).resolve()
        if not destination.is_relative_to(root):
            raise ValueError("Package path is outside output root")
        with path.open("rb") as stream:
            if hashlib.file_digest(stream, "sha256").hexdigest() != asset["sha256"].lower():
                raise ValueError("Package source hash mismatch")
        destination.parent.mkdir(parents=True, exist_ok=True)
        if path.resolve() != destination:
            shutil.copyfile(path, destination)
    target = root / test_prefix(category_slug, test_slug) / "content.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(content, ensure_ascii=False, indent=2, allow_nan=False), encoding="utf-8")
    return target.parent


def relocate_content(content, source_keys):
    """Replace media identities consistently; questions and answers stay intact."""
    def visit(value):
        if isinstance(value, dict):
            return {key: source_keys.get(child, child) if key in ("sourceId", "assetId", "audioAssetId") and isinstance(child, str) else visit(child) for key, child in value.items()}
        if isinstance(value, list):
            return [visit(child) for child in value]
        return value
    return visit(content)


def minio_client():
    from minio import Minio
    endpoint = urlsplit(os.environ.get("MINIO_ENDPOINT") or "http://localhost:9000")
    if endpoint.scheme not in ("http", "https") or not endpoint.netloc:
        raise ValueError("MINIO_ENDPOINT must be an HTTP(S) URL")
    return Minio(endpoint.netloc, access_key=os.environ["MINIO_ROOT_USER"], secret_key=os.environ["MINIO_ROOT_PASSWORD"], secure=endpoint.scheme == "https")


def ensure_bucket(client):
    from minio.error import S3Error
    if not client.bucket_exists(BUCKET):
        client.make_bucket(BUCKET)
    policy = {"Version": "2012-10-17", "Statement": [{"Sid": "ToeicMediaRead", "Effect": "Allow", "Principal": {"AWS": ["*"]}, "Action": ["s3:GetObject"], "Resource": [f"arn:aws:s3:::{BUCKET}/*/*/images/*", f"arn:aws:s3:::{BUCKET}/*/*/audio/*"]}]}
    try:
        existing = json.loads(client.get_bucket_policy(BUCKET))
    except S3Error as error:
        if error.code != "NoSuchBucketPolicy":
            raise
    else:
        # MinIO serializes action/resource sets in its own order.
        for document in (existing, policy):
            for statement in document.get("Statement", []):
                for key in ("Action", "Resource"):
                    if isinstance(statement.get(key), list):
                        statement[key] = sorted(statement[key])
        if existing != policy:
            raise ValueError("TOEIC bucket has a different policy; refusing to overwrite it")
        return
    # JSON exams remain private; only images and audio are anonymously readable.
    client.set_bucket_policy(BUCKET, json.dumps(policy))


def verify_object(client, bucket, key, expected):
    response = client.get_object(bucket, key)
    try:
        digest = hashlib.sha256()
        for chunk in response.stream(1024 * 1024):
            digest.update(chunk)
        if digest.hexdigest() != expected.lower():
            raise ValueError(f"Object hash mismatch: {bucket}/{key}")
    finally:
        response.close()
        response.release_conn()


def publish_json(client, category_slug, test_slug, content, *, expected_content=None):
    assert_json_available(client, category_slug, test_slug, content, expected_content=expected_content)
    data = json.dumps(content, ensure_ascii=False, indent=2, allow_nan=False).encode("utf-8")
    key = f"{test_prefix(category_slug, test_slug)}/content.json"
    client.put_object(BUCKET, key, io.BytesIO(data), len(data), content_type="application/json")
    verify_object(client, BUCKET, key, hashlib.sha256(data).hexdigest())
    return key


def assert_json_available(client, category_slug, test_slug, content, *, expected_content=None):
    from minio.error import S3Error
    key = f"{test_prefix(category_slug, test_slug)}/content.json"
    try:
        response = client.get_object(BUCKET, key)
    except S3Error as error:
        if error.code == "NoSuchKey":
            return
        raise
    try:
        current = json.loads(response.read())
        if current != content and current != expected_content:
            raise ValueError("Existing test JSON differs; use a different category/test slug")
    finally:
        response.close()
        response.release_conn()
