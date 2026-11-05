"""Move existing TOEIC storage into category/test folders, with a local backup.

Default: inventory only. --apply backs up, verifies copies and updates media
identities atomically in PostgreSQL. --prune removes only the backed-up legacy
keys after checking the live database and every destination again.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath

from dotenv import load_dotenv
from PIL import Image
import psycopg
from psycopg import sql
from psycopg.types.json import Jsonb

from publish_toeic_media import media_assets
from toeic_storage import BUCKET, ensure_bucket, media_key, media_labels, minio_client, publish_json, relocate_content, test_prefix, verify_object

LEGACY_PREFIX = "exams/toeic/"


def connect():
    # Match the Go API: DB_* settings win over the convenience DATABASE_URL.
    if os.getenv("DB_HOST") or os.getenv("DB_NAME"):
        return psycopg.connect(host=os.getenv("DB_HOST", "localhost"), port=os.getenv("DB_PORT", "5777"), user=os.getenv("DB_USER", "postgres"), password=os.getenv("DB_PASSWORD", "postgres"), dbname=os.getenv("DB_NAME", "engflex"))
    return psycopg.connect(os.environ["DATABASE_URL"])


def write_state(directory, state):
    (directory / "backup.json").write_text(json.dumps(state, ensure_ascii=False, indent=2), encoding="utf-8")


def legacy_references(db, *, skip_exam_content=False):
    with db.cursor() as cursor:
        cursor.execute("SELECT table_name,column_name FROM information_schema.columns WHERE table_schema='public' AND data_type IN ('text','character varying','json','jsonb')")
        columns = cursor.fetchall()
        found = []
        for table, column in columns:
            if skip_exam_content and (table, column) == ("exam_tests", "content"):
                continue
            cursor.execute(sql.SQL("SELECT count(*) FROM {} WHERE {}::text LIKE %s").format(sql.Identifier(table), sql.Identifier(column)), (f"%{LEGACY_PREFIX}%",))
            count = cursor.fetchone()[0]
            if count:
                found.append({"table": table, "column": column, "rows": count})
        return found


def inventory(client, db):
    bucket = os.environ.get("MINIO_BUCKET") or "lesson-audio"
    objects = [{"key": o.object_name, "size": o.size, "etag": o.etag} for o in client.list_objects(bucket, prefix=LEGACY_PREFIX, recursive=True)]
    with db.cursor() as cursor:
        cursor.execute("SELECT t.id::text,c.slug,t.slug,t.content FROM exam_tests t JOIN exam_categories c ON c.id=t.category_id WHERE c.type='toeic' ORDER BY c.slug,t.slug")
        tests = [{"id": tid, "category": category, "slug": slug, "original": content} for tid, category, slug, content in cursor.fetchall()]
    if not objects or not tests:
        raise ValueError("No legacy TOEIC objects or tests found; refusing an empty migration")
    for test in tests:
        test_prefix(test["category"], test["slug"])
    extra = legacy_references(db, skip_exam_content=True)
    if extra:
        raise ValueError(f"Legacy objects are referenced outside TOEIC content: {extra}")
    return {"sourceBucket": bucket, "destinationBucket": BUCKET, "objects": objects, "tests": tests, "applied": False}


def apply(client, db, directory, state):
    directory.mkdir(parents=True, exist_ok=False)
    write_state(directory, state)
    for obj in state["objects"]:
        parts = PurePosixPath(obj["key"]).parts
        if not obj["key"].startswith(LEGACY_PREFIX) or any(p in ("..", ".") or "\\" in p for p in parts):
            raise ValueError("Unsafe legacy object key")
        path = directory.joinpath("objects", *parts)
        path.parent.mkdir(parents=True, exist_ok=True)
        stat = client.stat_object(state["sourceBucket"], obj["key"])
        if stat.etag != obj["etag"] or stat.size != obj["size"]:
            raise ValueError("Legacy inventory changed before backup")
        client.fget_object(state["sourceBucket"], obj["key"], str(path))
        if path.stat().st_size != obj["size"]:
            raise ValueError("Incomplete object backup")
        with path.open("rb") as stream:
            obj["sha256"] = hashlib.file_digest(stream, "sha256").hexdigest()
        obj["contentType"] = stat.content_type
    write_state(directory, state)
    objects = {obj["key"]: obj for obj in state["objects"]}
    for test in state["tests"]:
        mapping = {}
        labels = media_labels(test["original"])
        test["media"] = []
        for asset in media_assets(test["original"]):
            old_key = LEGACY_PREFIX + asset["sha256"].lower()
            obj = objects.get(old_key)
            if not obj or obj["sha256"] != asset["sha256"].lower():
                raise ValueError(f"Verified source missing: {old_key}")
            if asset["sourceId"] == test["original"]["audioAssetId"]:
                content_type = "audio/mpeg"
            else:
                with Image.open(directory / "objects" / old_key) as image:
                    content_type = Image.MIME.get(image.format)
                    image.verify()
            key = media_key(test["category"], test["slug"], asset, content_type, label=labels[asset["sourceId"]])
            mapping[asset["sourceId"]] = f"{BUCKET}/{key}"
            test["media"].append({"key": key, "source": old_key, "sha256": asset["sha256"].lower(), "contentType": content_type})
        test["content"] = relocate_content(test["original"], mapping)
        if relocate_content(test["content"], {v: k for k, v in mapping.items()}) != test["original"]:
            raise ValueError("Content relocation changed more than media identities")
    write_state(directory, state)
    ensure_bucket(client)
    for test in state["tests"]:
        for media in test["media"]:
            client.fput_object(BUCKET, media["key"], str(directory / "objects" / media["source"]), content_type=media["contentType"])
            verify_object(client, BUCKET, media["key"], media["sha256"])
        test["jsonKey"] = publish_json(client, test["category"], test["slug"], test["content"])
        print(f"Verified {BUCKET}/{test['category']}/{test['slug']}: {len(test['media'])} media + content.json", flush=True)
    # The connection context commits only when every row still matches backup.
    with db.transaction(), db.cursor() as cursor:
        for test in state["tests"]:
            cursor.execute("UPDATE exam_tests SET content=%s WHERE id=%s AND content=%s", (Jsonb(test["content"]), test["id"], Jsonb(test["original"])))
            if cursor.rowcount != 1:
                raise ValueError("Exam changed during copy; database changes rolled back")
    db.commit()
    state["applied"] = True
    write_state(directory, state)
    print("Copied and verified all destinations; database updated. Legacy objects retained until --prune.", flush=True)


def prune(client, db, directory):
    state = json.loads((directory / "backup.json").read_text(encoding="utf-8"))
    if not state["applied"] or state["destinationBucket"] != BUCKET:
        raise ValueError("An applied, verified migration backup is required")
    references = legacy_references(db)
    if references:
        raise ValueError(f"Legacy media is still referenced: {references}")
    with db.cursor() as cursor:
        for test in state["tests"]:
            cursor.execute("SELECT content FROM exam_tests WHERE id=%s", (test["id"],))
            row = cursor.fetchone()
            if not row or row[0] != test["content"]:
                raise ValueError("Exam changed since migration; prune aborted")
            for media in test["media"]:
                verify_object(client, BUCKET, media["key"], media["sha256"])
            response = client.get_object(BUCKET, test["jsonKey"])
            try:
                if json.loads(response.read()) != test["content"]:
                    raise ValueError("Stored exam JSON differs from database")
            finally:
                response.close(); response.release_conn()
    # Validate every source and every backup before the first deletion.
    for obj in state["objects"]:
        if not obj["key"].startswith(LEGACY_PREFIX):
            raise ValueError("Deletion outside the legacy TOEIC prefix refused")
        stat = client.stat_object(state["sourceBucket"], obj["key"])
        if stat.etag != obj["etag"] or stat.size != obj["size"]:
            raise ValueError("Legacy object changed since backup; prune aborted")
        with (directory / "objects" / obj["key"]).open("rb") as stream:
            if hashlib.file_digest(stream, "sha256").hexdigest() != obj["sha256"]:
                raise ValueError("Backup integrity failure")
    for obj in state["objects"]:
        client.remove_object(state["sourceBucket"], obj["key"])
    state["pruned"] = True
    write_state(directory, state)
    remaining = list(client.list_objects(state["sourceBucket"], prefix=LEGACY_PREFIX, recursive=True))
    if remaining:
        raise ValueError("Unexpected objects remain in the legacy prefix")
    print(json.dumps({"removedLegacyObjects": len(state["objects"]), "removedLegacyBytes": sum(o["size"] for o in state["objects"]), "mediaObjects": sum(len(t["media"]) for t in state["tests"]), "examJsonFiles": len(state["tests"]), "backup": str(directory)}), flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--backup-dir", type=Path, required=True)
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument("--apply", action="store_true")
    mode.add_argument("--prune", action="store_true")
    args = parser.parse_args()
    load_dotenv()
    client = minio_client()
    with connect() as db:
        if args.prune:
            prune(client, db, args.backup_dir)
        else:
            state = inventory(client, db)
            print(json.dumps({"tests": len(state["tests"]), "legacyObjects": len(state["objects"]), "legacyBytes": sum(o["size"] for o in state["objects"]), "layout": [f"{BUCKET}/{t['category']}/{t['slug']}" for t in state["tests"]]}), flush=True)
            if args.apply:
                apply(client, db, args.backup_dir, state)


if __name__ == "__main__":
    main()
