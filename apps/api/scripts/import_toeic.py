"""Import one reviewed TOEIC release. Run from apps/api with uv and psycopg.

Media manifest: JSON object mapping each export sourceId to an accessible HTTP(S)
URL. The URLs must serve the exact bytes bound by the release asset hashes.
Requires ffprobe on PATH and configured MinIO. Verified media is published to
MinIO before inserting the test; source URLs remain manifest identities only.
"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import subprocess
import tempfile
from urllib.parse import urlparse
from urllib.request import urlopen


def digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True, allow_nan=False).encode()).hexdigest()


def validate_content(content):
    subprocess.run(["go", "run", "./cmd/validate-toeic"], input=json.dumps(content, ensure_ascii=False).encode(), check=True)


def verify_release(content, provenance, report):
    if provenance.get("status") != "release" or provenance.get("exportSchemaVersion") != "0.1.0" or provenance.get("contentHash") != digest(content):
        raise ValueError("Release provenance missing, stale, or not reviewed")
    if not provenance.get("draftHash") or report.get("payloadHash") != provenance["draftHash"] or report.get("releaseReady") is not True or report.get("issueCounts") or report.get("stageErrors") or not isinstance(report.get("issues"), list) or any(issue.get("status") != "verified" for issue in report["issues"]):
        raise ValueError("Release review report missing, stale, or incomplete")
    validate_content(content)


def prepare(content, provenance, report, media):
    verify_release(content, provenance, report)
    assets = {asset["sourceId"]: asset["sha256"].lower() for asset in content["assets"]}
    if set(media) != set(assets) or len(set(media.values())) != len(media):
        raise ValueError("Media manifest must map each asset to a distinct URL")
    audio_duration = None
    for source, expected in assets.items():
        url = media[source]
        if not isinstance(url, str) or urlparse(url).scheme not in ("http", "https") or not urlparse(url).netloc:
            raise ValueError(f"Invalid media URL for {source}")
        with urlopen(url, timeout=30) as response, tempfile.TemporaryDirectory() as directory:
            if urlparse(response.geturl()).scheme not in ("http", "https"):
                raise ValueError(f"Invalid redirect for {source}")
            h = hashlib.sha256()
            path = Path(directory) / "asset"
            with path.open("wb") as output:
                while chunk := response.read(1024 * 1024):
                    h.update(chunk)
                    output.write(chunk)
            if h.hexdigest() != expected:
                raise ValueError(f"Media hash mismatch for {source}")
            probe = subprocess.run(["ffprobe", "-v", "error", "-show_entries", "format=duration:stream=codec_type", "-of", "json", str(path)], capture_output=True, text=True, check=True)
            info = json.loads(probe.stdout)
            stream_type = "audio" if source == content["audioAssetId"] else "video"
            if not any(stream.get("codec_type") == stream_type for stream in info.get("streams", [])):
                raise ValueError(f"Invalid {stream_type} media for {source}")
            if stream_type == "audio":
                audio_duration = float(info.get("format", {}).get("duration", "nan"))
                if not math.isfinite(audio_duration) or audio_duration <= 0:
                    raise ValueError("Invalid audio duration")
    for part in content["parts"]:
        for item in part["items"]:
            for key in ("itemAudio", "stimulusAudio"):
                audio = item.get(key)
                if audio and audio["endTime"] > audio_duration:
                    raise ValueError(f"Audio range exceeds duration: {item['id']}")
    # Rewrite identities consistently so the stored graph resolves directly to
    # verified media URLs. The original release hash was checked first.
    def rewrite(value):
        if isinstance(value, dict):
            return {k: media[v] if k in ("sourceId", "assetId", "audioAssetId") and isinstance(v, str) and v in media else rewrite(v) for k, v in value.items()}
        if isinstance(value, list):
            return [rewrite(v) for v in value]
        return value
    result = rewrite(content)
    validate_content(result)
    return result


def prepare_package(original, provenance, report, packaged):
    verify_release(original, provenance, report)
    from toeic_storage import relocate_content
    by_hash = {asset["sha256"].lower(): asset["sourceId"] for asset in packaged["assets"]}
    try:
        mapping = {asset["sourceId"]: by_hash[asset["sha256"].lower()] for asset in original["assets"]}
    except KeyError as error:
        raise ValueError("Package is missing a release asset") from error
    if relocate_content(original, mapping) != packaged:
        raise ValueError("Package differs from the reviewed release")
    validate_content(packaged)
    return packaged


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--release", type=Path, required=True, help="Directory containing content.json and provenance.json")
    inputs = parser.add_mutually_exclusive_group(required=True)
    inputs.add_argument("--media", type=Path, help="JSON sourceId to ready URL map")
    inputs.add_argument("--package-root", type=Path, help="Local output root containing toeic/<category>/<test>/")
    parser.add_argument("--category-slug", required=True)
    parser.add_argument("--category-title", required=True)
    parser.add_argument("--test-slug", required=True)
    args = parser.parse_args()
    read = lambda p: json.loads(p.read_text(encoding="utf-8"))
    original = read(args.release / "content.json")
    provenance = read(args.release / "provenance.json")
    report = read(args.release.parent / "validation.json")
    if args.media:
        content = prepare(original, provenance, report, read(args.media))
    else:
        from toeic_storage import test_prefix
        packaged = read(args.package_root / "toeic" / test_prefix(args.category_slug, args.test_slug) / "content.json")
        content = prepare_package(original, provenance, report, packaged)
    from publish_toeic_media import publish_content
    content, _ = publish_content(content, source_root=args.package_root, category_slug=args.category_slug, test_slug=args.test_slug)
    import psycopg
    from psycopg.types.json import Jsonb
    from psycopg.errors import UniqueViolation
    try:
        with psycopg.connect(os.environ["DATABASE_URL"]) as conn, conn.cursor() as cur:
            cur.execute("INSERT INTO exam_categories (slug, title, type) VALUES (%s, %s, 'toeic') ON CONFLICT (slug) DO NOTHING", (args.category_slug, args.category_title))
            cur.execute("SELECT id, title, type FROM exam_categories WHERE slug = %s FOR UPDATE", (args.category_slug,))
            category_id, title, kind = cur.fetchone()
            if title != args.category_title or kind != "toeic":
                raise ValueError("Existing category does not match title/type")
            cur.execute("INSERT INTO exam_tests (category_id, source_key, slug, title, content) VALUES (%s, %s, %s, %s, %s) RETURNING id", (category_id, original["testId"], args.test_slug, original["title"], Jsonb(content)))
            test_id = cur.fetchone()[0]
        print(f"Imported test {test_id}")
    except UniqueViolation as exc:
        raise SystemExit(f"Duplicate category/test identity: {exc.diag.constraint_name}") from exc


if __name__ == "__main__":
    main()
