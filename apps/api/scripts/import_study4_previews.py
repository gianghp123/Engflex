"""Import technically validated Study4 previews only with explicit authorization.

Review status remains preview-needs-review in the import receipt. This command
does not manufacture release provenance or a human review attestation.
"""

import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import re
import sys

from dotenv import load_dotenv
from psycopg.types.json import Jsonb

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))
from export_study4_content import validate_content
from study4_content import build_content, digest
from organize_toeic_storage import connect
from publish_toeic_media import publish_content
from toeic_storage import relocate_content


def read(path):
    return path.read_text(encoding="utf-8")


def read_json(path):
    return json.loads(read(path))


def file_hash(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def prepare_bundle(test_id, args):
    root = args.crawl_root / "study4/tests" / test_id
    bundle = root / "database-named"
    content = read_json(bundle / "preview/content.json")
    provenance = read_json(bundle / "preview/provenance.json")
    report = read_json(bundle / "validation.json")
    source = read_json(bundle / "source-manifest.json")
    if content["testId"] != f"study4_{test_id}" or content["collectionId"] != args.collection_id:
        raise ValueError("Study4 test/collection identity mismatch")
    if provenance["status"] != "preview-needs-review" or report["releaseReady"]:
        raise ValueError("Use the reviewed-release importer for releases")
    if provenance["contentHash"] != report["contentHash"] or provenance["contentHash"] != digest(content):
        raise ValueError("Preview content hash mismatch")
    if provenance["draftHash"] != digest(source) or provenance["sourceManifestHash"] != digest(source) or report["payloadHash"] != digest(source):
        raise ValueError("Source provenance hash mismatch")
    if report["stageErrors"] or report["issueCounts"] != {"sourceReviewRequired": 1}:
        raise ValueError("Unresolved technical issues prevent preview import")
    run = root / "runs" / source["runId"]
    if file_hash(run / "export.json") != source["exportFileSha256"] or file_hash(run / "audio-alignment.json") != source["alignmentFileSha256"]:
        raise ValueError("Run changed since preview validation")
    data, alignment = read_json(run / "export.json"), read_json(run / "audio-alignment.json")
    rebuilt = build_content(data, collection_id=args.collection_id, test_number=content["testNumber"], title=content["title"], alignment=alignment)
    if rebuilt != content:
        raise ValueError("Preview does not rebuild exactly from source evidence")
    by_hash = {asset["sha256"].lower(): asset for asset in source["assets"]}
    mapping = {}
    for asset in content["assets"]:
        original = by_hash[asset["sha256"].lower()]
        if not original.get("localPath"):
            raise ValueError("Verified local media is required for preview import")
        path = (root / original["localPath"]).resolve()
        if not path.is_relative_to(root.resolve()) or file_hash(path) != asset["sha256"].lower():
            raise ValueError("Local media boundary/hash mismatch")
        mapping[asset["sourceId"]] = original["localPath"]
    local = relocate_content(content, mapping)
    slug = f"{args.category_slug}-test-{content['testNumber']}"
    stored, summary = publish_content(local, source_root=root, dry_run=True, category_slug=args.category_slug, test_slug=slug, output_root=args.output_root)
    validate_content(stored)
    print(f"Validated {slug}: 200 questions, {summary['validatedAssets']} assets, 54 listening ranges", flush=True)
    return {"root": root, "bundle": bundle, "original": content, "local": local, "stored": stored, "slug": slug}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--test-ids", nargs="+", required=True)
    parser.add_argument("--allow-unreviewed", action="store_true")
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--crawl-root", type=Path, default=ROOT / "docs/crawl")
    parser.add_argument("--output-root", type=Path, default=ROOT / "scripts/toeic_pipeline/output")
    parser.add_argument("--category-slug", default="toeic-lr-collection-1")
    parser.add_argument("--category-title", default="TOEIC LR Collection 1")
    parser.add_argument("--collection-id", default="study4_toeic_lr_collection_1")
    args = parser.parse_args()
    if not args.allow_unreviewed:
        parser.error("Explicit --allow-unreviewed authorization is required for previews")
    if not all(re.fullmatch(r"\d+", tid) for tid in args.test_ids):
        parser.error("Test IDs must be numeric")
    load_dotenv(ROOT / "apps/api/.env")
    bundles = [prepare_bundle(tid, args) for tid in args.test_ids]
    with connect() as db, db.cursor() as cursor:
        cursor.execute("SELECT source_key FROM exam_tests WHERE source_key = ANY(%s) OR slug = ANY(%s)", ([b["original"]["testId"] for b in bundles], [b["slug"] for b in bundles]))
        if cursor.fetchall():
            raise ValueError("A requested test already exists; refusing to overwrite")
        cursor.execute("SELECT id,title,type FROM exam_categories WHERE slug=%s", (args.category_slug,))
        category = cursor.fetchone()
        if not category or category[1:] != (args.category_title, "toeic"):
            raise ValueError("Existing category does not match the requested collection")
    if not args.apply:
        print("Dry run passed; no MinIO or database writes", flush=True)
        return
    for bundle in bundles:
        stored, _ = publish_content(bundle["local"], source_root=bundle["root"], category_slug=args.category_slug, test_slug=bundle["slug"])
        if stored != bundle["stored"]:
            raise ValueError("Published content differs from the validated package")
    records = []
    with connect() as db, db.cursor() as cursor:
        cursor.execute("SELECT id FROM exam_categories WHERE slug=%s FOR UPDATE", (args.category_slug,))
        category_id = cursor.fetchone()[0]
        for bundle in bundles:
            content = bundle["stored"]
            cursor.execute("INSERT INTO exam_tests (category_id,source_key,slug,title,content) VALUES (%s,%s,%s,%s,%s) RETURNING id", (category_id, content["testId"], bundle["slug"], content["title"], Jsonb(content)))
            records.append((bundle, str(cursor.fetchone()[0])))
    with connect() as db, db.cursor() as cursor:
        for bundle, test_id in records:
            cursor.execute("SELECT content FROM exam_tests WHERE id=%s", (test_id,))
            if cursor.fetchone()[0] != bundle["stored"]:
                raise ValueError("Committed database content mismatch")
            receipt = {"importedAt": datetime.now(timezone.utc).isoformat(), "id": test_id, "sourceKey": bundle["stored"]["testId"], "slug": bundle["slug"], "sourceReviewStatus": "preview-needs-review", "authorization": "User explicitly authorized importing these validated Study4 previews", "contentHash": digest(bundle["stored"]), "questionCount": 200, "assetCount": len(bundle["stored"]["assets"])}
            (bundle["bundle"] / "import-record.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2), encoding="utf-8")
            print(json.dumps(receipt, ensure_ascii=True), flush=True)
    print(f"Committed and verified {len(records)} previews; no existing tests or attempts changed", flush=True)


if __name__ == "__main__":
    main()
