"""Import verified IELTS packages from MinIO. Dry run unless --apply is supplied."""

import argparse
import json
import os
import re
import subprocess
from pathlib import Path

from dotenv import load_dotenv
import psycopg
from psycopg.types.json import Jsonb

from toeic_storage import minio_client, verify_object

ROOT = Path(__file__).resolve().parents[3]


def read_package(client, source_id, key):
    if not re.fullmatch(r"\d+", source_id):
        raise ValueError("Expected numeric Study4 source ID")
    match = re.fullmatch(r"([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)/content\.json", key)
    if not match:
        raise ValueError("Invalid IELTS content key")
    response = client.get_object("ielts", key)
    try:
        raw = response.read()
    finally:
        response.close()
        response.release_conn()
    subprocess.run(["go", "run", "./cmd/validate-ielts"], cwd=ROOT / "apps/api", input=raw, check=True)
    content = json.loads(raw)
    category, slug = match.groups()
    if content["testId"] != slug or content["skill"] not in {"reading", "listening"}:
        raise ValueError("IELTS package identity/skill mismatch")
    if [q["questionNumber"] for section in content["sections"] for group in section["questionGroups"] for q in group["questions"]] != list(range(1, 41)):
        raise ValueError("Expected all 40 IELTS questions")
    for asset in content.get("assets", []):
        source = asset["sourceId"]
        if not source.startswith(f"ielts/{category}/{slug}/") or ".." in source:
            raise ValueError("Media reference is outside this IELTS package")
        verify_object(client, "ielts", source.removeprefix("ielts/"), asset["sha256"])
    return {"category": category, "slug": slug, "sourceKey": f"study4_{source_id}", "content": content}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--test", nargs=2, action="append", metavar=("SOURCE_ID", "CONTENT_KEY"), required=True)
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    load_dotenv(ROOT / "apps/api/.env")
    client = minio_client()
    packages = [read_package(client, source_id, key) for source_id, key in args.test]
    results = []
    with psycopg.connect(os.environ["DATABASE_URL"]) as connection:
        with connection.transaction():
            for package in packages:
                content = package["content"]
                category = connection.execute("SELECT id, type, sub_type FROM exam_categories WHERE slug = %s FOR SHARE", (package["category"],)).fetchone()
                if not category or category[1:] != ("ielts", content["subType"]):
                    raise ValueError("Missing or mismatched IELTS category")
                existing = connection.execute("SELECT id, slug, source_key, content, category_id, skill, sub_type FROM exam_tests WHERE slug = %s OR source_key = %s", (package["slug"], package["sourceKey"])).fetchall()
                if existing:
                    if len(existing) != 1 or existing[0][1:] != (package["slug"], package["sourceKey"], content, category[0], content["skill"], content["subType"]):
                        raise ValueError("Existing exam differs; refusing to overwrite")
                    results.append({"id": str(existing[0][0]), "slug": package["slug"], "status": "already-imported"})
                elif args.apply:
                    row = connection.execute("INSERT INTO exam_tests (category_id, source_key, slug, title, content, skill, sub_type) VALUES (%s,%s,%s,%s,%s,%s,%s) RETURNING id", (category[0], package["sourceKey"], package["slug"], content["title"], Jsonb(content), content["skill"], content["subType"])).fetchone()
                    results.append({"id": str(row[0]), "slug": package["slug"], "status": "imported"})
                else:
                    results.append({"slug": package["slug"], "status": "validated-dry-run"})
    print(json.dumps(results))


if __name__ == "__main__":
    main()
