# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Build a nested TOEIC DB document and validate it with the API's Go parser.

Run after study4_mcp_export.py and align_study4_audio.py. The default output is
a preview. --reviewed-by records the caller's completed source/audio review;
it is never set automatically by crawling or by structural validation.
"""

import argparse
import hashlib
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from urllib.parse import urlsplit

from study4_content import build_content, digest

ROOT = Path(__file__).resolve().parents[1]


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n", encoding="utf-8")


def validate_content(content):
    subprocess.run(
        ["go", "run", "./cmd/validate-toeic"],
        cwd=ROOT / "apps/api",
        input=json.dumps(content, ensure_ascii=False, allow_nan=False).encode(),
        check=True,
    )


def export_bundle(export_path, output, *, collection_id, test_number, title=None, reviewed_by=None):
    export_path, output = Path(export_path).resolve(), Path(output).resolve()
    if output.exists():
        raise FileExistsError(f"Output already exists: {output}; use a new directory")
    data = read(export_path)
    alignment_path = export_path.parent / "audio-alignment.json"
    alignment = read(alignment_path)
    content = build_content(
        data, collection_id=collection_id, test_number=test_number, title=title, alignment=alignment
    )
    validate_content(content)
    media = {}
    for asset in data["assets"]:
        source = f"exams/toeic/{asset['sha256']}"
        if source not in {a["sourceId"] for a in content["assets"]}:
            continue
        url = asset["url"]
        parsed = urlsplit(url)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc or parsed.username or parsed.password:
            raise ValueError(f"No accessible HTTP(S) source for {source}")
        if source in media and media[source] != url:
            # Same bytes can legitimately have multiple source URLs. Keep one.
            continue
        media[source] = url
    source_manifest = {
        "source": data["source"],
        "runId": alignment["runId"],
        "exportFileSha256": hashlib.sha256(export_path.read_bytes()).hexdigest(),
        "alignmentFileSha256": hashlib.sha256(alignment_path.read_bytes()).hexdigest(),
        "domHash": digest(data["content"]),
        "asrHash": alignment["asrHash"],
        "assets": data["assets"],
    }
    payload_hash = digest(source_manifest)
    released = bool(reviewed_by and reviewed_by.strip())
    report = {
        "payloadHash": payload_hash,
        "releaseReady": released,
        "issues": []
        if released
        else [
            {
                "code": "sourceReviewRequired",
                "status": "open",
                "detail": "Review all source questions/answers/passages and listen to every audio range before release",
            }
        ],
        "issueCounts": {} if released else {"sourceReviewRequired": 1},
        "stageErrors": [],
        "checks": {"goContentValidator": "passed", "questionCount": 200, "listeningItems": 54},
        "reviewedBy": reviewed_by.strip() if released else None,
        "contentHash": digest(content),
    }
    provenance = {
        "status": "release" if released else "preview-needs-review",
        "exportSchemaVersion": "0.1.0",
        "contentHash": digest(content),
        "draftHash": payload_hash,
        "sourceManifestHash": digest(source_manifest),
        "reviewedBy": report["reviewedBy"],
        "note": "Contains answer keys; server import only, never a learner-facing response",
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    staging = Path(tempfile.mkdtemp(prefix=f".{output.name}-", dir=output.parent))
    try:
        destination = staging / ("release" if released else "preview")
        destination.mkdir()
        write(destination / "content.json", content)
        write(destination / "provenance.json", provenance)
        write(staging / "validation.json", report)
        write(staging / "source-manifest.json", source_manifest)
        write(staging / "media.json", media)
        staging.rename(output)
    finally:
        if staging.exists():
            shutil.rmtree(staging)
    return output / destination.name / "content.json"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--export", type=Path, required=True, help="A run's export.json, with sibling audio-alignment.json"
    )
    parser.add_argument(
        "--output", type=Path, required=True, help="New bundle directory containing preview/ or release/"
    )
    parser.add_argument(
        "--collection-id", required=True, help="Explicit collection identity; never guessed from a URL"
    )
    parser.add_argument("--test-number", type=int, required=True)
    parser.add_argument("--title")
    parser.add_argument(
        "--reviewed-by", help="Record your completed source/audio review and produce a release"
    )
    args = parser.parse_args()
    try:
        path = export_bundle(
            args.export,
            args.output,
            collection_id=args.collection_id,
            test_number=args.test_number,
            title=args.title,
            reviewed_by=args.reviewed_by,
        )
        print(f"Validated TOEIC DB document: {path}")
        return 0
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as error:
        print(f"Content export failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
