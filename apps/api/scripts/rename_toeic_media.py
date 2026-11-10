"""Rename TOEIC media by test, part and question group, with verified backups.

--apply copies new names, exports the local package, and updates JSON/database.
--prune removes the backed-up old names only after all references are updated.
"""

import argparse
import hashlib
import json
from pathlib import Path

from dotenv import load_dotenv
from minio.commonconfig import CopySource
from psycopg import sql
from psycopg.types.json import Jsonb

from organize_toeic_storage import connect, write_state
from toeic_storage import BUCKET, export_package, media_key, media_labels, minio_client, publish_json, relocate_content, test_prefix, verify_object


def plan(client, db):
    with db.cursor() as cursor:
        cursor.execute("SELECT t.id::text,c.slug,t.slug,t.content FROM exam_tests t JOIN exam_categories c ON c.id=t.category_id WHERE c.type='toeic' ORDER BY c.slug,t.slug")
        rows = cursor.fetchall()
    tests = []
    for tid, category, slug, content in rows:
        prefix = test_prefix(category, slug)
        labels = media_labels(content)
        mapping, media = {}, []
        for asset in content["assets"]:
            source = asset["sourceId"]
            if not source.startswith(f"{BUCKET}/{prefix}/") or ".." in source or "\\" in source:
                raise ValueError("Media identity is outside its test folder")
            old = source.removeprefix(f"{BUCKET}/")
            stat = client.stat_object(BUCKET, old)
            key = media_key(category, slug, asset, stat.content_type, label=labels[source])
            mapping[source] = f"{BUCKET}/{key}"
            media.append({"old":old,"key":key,"sha256":asset["sha256"].lower(),"size":stat.size,"etag":stat.etag,"contentType":stat.content_type})
        stored = relocate_content(content, mapping)
        if relocate_content(stored, {v:k for k,v in mapping.items()}) != content:
            raise ValueError("Relocation changed exam content")
        tests.append({"id":tid,"category":category,"slug":slug,"original":content,"content":stored,"media":media,"jsonKey":f"{prefix}/content.json"})
    return {"tests":tests,"applied":False}


def apply(client, db, state, backup, output_root):
    backup.mkdir(parents=True, exist_ok=False)
    write_state(backup, state)
    # Back up all old media and the original full JSON before the first write.
    for test in state["tests"]:
        for media in test["media"]:
            path = backup / "objects" / media["old"]
            path.parent.mkdir(parents=True, exist_ok=True)
            client.fget_object(BUCKET, media["old"], str(path))
            with path.open("rb") as stream:
                if hashlib.file_digest(stream,"sha256").hexdigest() != media["sha256"]:
                    raise ValueError("Backup media hash mismatch")
        path = backup / "objects" / test["jsonKey"]
        path.parent.mkdir(parents=True, exist_ok=True)
        client.fget_object(BUCKET, test["jsonKey"], str(path))
        if json.loads(path.read_text(encoding="utf-8")) != test["original"]:
            raise ValueError("Stored JSON differs from the database")
    for test in state["tests"]:
        prepared = []
        for asset, media in zip(test["original"]["assets"], test["media"], strict=True):
            if media["old"] != media["key"]:
                client.copy_object(BUCKET, media["key"], CopySource(BUCKET, media["old"]))
            verify_object(client, BUCKET, media["key"], media["sha256"])
            prepared.append((asset, backup / "objects" / media["old"], media["key"], media["contentType"]))
        export_package(test["content"], prepared, output_root, test["category"], test["slug"])
    try:
        with db.transaction(), db.cursor() as cursor:
            for test in state["tests"]:
                cursor.execute("UPDATE exam_tests SET content=%s WHERE id=%s AND content=%s", (Jsonb(test["content"]),test["id"],Jsonb(test["original"])))
                if cursor.rowcount != 1:
                    raise ValueError("Exam changed during rename")
            for test in state["tests"]:
                publish_json(client,test["category"],test["slug"],test["content"],expected_content=test["original"])
        db.commit()
    except Exception:
        db.rollback()
        for test in state["tests"]:
            client.fput_object(BUCKET,test["jsonKey"],str(backup / "objects" / test["jsonKey"]),content_type="application/json")
        raise
    state["applied"] = True
    write_state(backup,state)
    print("Verified copies, exported local packages, and updated media references. Old names retained until --prune.",flush=True)


def prune(client,db,state,backup):
    if not state["applied"]:
        raise ValueError("Applied backup required")
    old = [m for t in state["tests"] for m in t["media"] if m["old"] != m["key"]]
    if not old:
        raise ValueError("No renamed objects to prune")
    with db.cursor() as cursor:
        cursor.execute("SELECT table_name,column_name FROM information_schema.columns WHERE table_schema='public' AND data_type IN ('text','character varying','json','jsonb')")
        columns = cursor.fetchall()
        for table,column in columns:
            cursor.execute(sql.SQL("SELECT count(*) FROM {} WHERE {}::text LIKE ANY(%s)").format(sql.Identifier(table),sql.Identifier(column)),([f"%{BUCKET}/{m['old']}%" for m in old],))
            if cursor.fetchone()[0]:
                raise ValueError(f"Old media still referenced in {table}.{column}")
        for test in state["tests"]:
            cursor.execute("SELECT content FROM exam_tests WHERE id=%s",(test["id"],))
            if cursor.fetchone()[0] != test["content"]:
                raise ValueError("Exam changed after rename")
            response = client.get_object(BUCKET,test["jsonKey"])
            try:
                if json.loads(response.read()) != test["content"]:
                    raise ValueError("Stored JSON does not match database")
            finally:
                response.close();response.release_conn()
            for media in test["media"]:
                verify_object(client,BUCKET,media["key"],media["sha256"])
    for media in old:
        stat = client.stat_object(BUCKET,media["old"])
        if stat.etag != media["etag"] or stat.size != media["size"]:
            raise ValueError("Old media changed after backup")
        with (backup / "objects" / media["old"]).open("rb") as stream:
            if hashlib.file_digest(stream,"sha256").hexdigest() != media["sha256"]:
                raise ValueError("Backup integrity failure")
    for media in old:
        client.remove_object(BUCKET,media["old"])
    state["pruned"] = True
    write_state(backup,state)
    print(f"Removed {len(old)} verified old media names; all named media and JSON retained.",flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--backup-dir",type=Path,required=True)
    parser.add_argument("--output-root",type=Path,default=Path("../../scripts/toeic_pipeline/output"))
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument("--apply",action="store_true")
    mode.add_argument("--prune",action="store_true")
    args = parser.parse_args()
    load_dotenv()
    client = minio_client()
    with connect() as db:
        if args.prune:
            state = json.loads((args.backup_dir / "backup.json").read_text(encoding="utf-8"))
            prune(client,db,state,args.backup_dir)
        else:
            state = plan(client,db)
            print(json.dumps({"tests":len(state["tests"]),"media":sum(len(t["media"]) for t in state["tests"]),"examples":[m["key"] for t in state["tests"] for m in t["media"]][:5]}),flush=True)
            if args.apply:
                apply(client,db,state,args.backup_dir,args.output_root)


if __name__ == "__main__":
    main()
