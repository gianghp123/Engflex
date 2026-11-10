"""TOEIC evidence pipeline. Run from any directory through uv --project."""

import argparse
import html
import json
import re
import sys
from pathlib import Path

from extractors import answer_key, digital, inventory, ocr, select_sources
from pipeline_io import BASE, ROOT, digest, read, test_dir, write
from schemas import Ref, Test
from structure import build, rc_answers
from validation import validate


def book_parts(book, number):
    pages = book["pages"]

    def pattern(n):
        return rf"(?:TEST|Test)\s*0?{n}\b"

    start = next((i for i, p in enumerate(pages) if re.search(pattern(number), p["text"])), None)
    if start is None:
        return {}, None
    end = next(
        (i for i in range(start + 1, len(pages)) if re.search(pattern(number + 1), pages[i]["text"])),
        len(pages),
    )
    groups = {}
    current = None
    for p in pages[start:end]:
        headings = list(re.finditer(r"PART\s*([1-7])\b", p["text"], re.I))
        # Transition pages belong to both adjacent parts; exact spans remain evidence.
        if current is not None:
            groups.setdefault(current, []).append(p)
        for m in headings:
            current = int(m[1])
            if p not in groups.setdefault(current, []):
                groups[current].append(p)
    return groups, {
        "sourceId": book["sourceId"],
        "startPage": start + 1,
        "endPage": end,
        "status": "needsReview",
    }


def exam_for_part(artifact, part):
    # Page ranges are discovered from OCR headings; include transition pages.
    current = None
    selected = []
    for p in artifact["pages"]:
        headings = [
            int(m[1]) for line in p["lines"] for m in re.finditer(r"\bPART\s*([1-7])\b", line["text"], re.I)
        ]
        if current == part or part in headings:
            selected.append(p)
        if headings:
            current = max(headings)
    return selected or artifact["pages"]


def run(args):
    manifest = inventory(args.test)
    sources = select_sources(manifest, args.test)
    out = test_dir(args.test)
    write(out / "sources.json", sources)
    if args.dry_run:
        print(
            json.dumps(
                {
                    "test": args.test,
                    "sourceIds": {k: v["sourceId"] for k, v in sources.items()},
                    "stages": ["extract", "structure", "audio", "validate", "review"],
                },
                indent=2,
            )
        )
        return 0
    books = {k: digital(sources[k]) for k in ("key", "lcBook", "rcBook")}
    for role, book in books.items():
        write(out / "raw" / (role + ".json"), book)
    keys = answer_key(books["key"], args.test)
    rc, mapping = rc_answers(books["rcBook"], args.test)
    keys += rc
    write(out / "answer_candidates.json", keys)
    book_groups = {}
    mappings = {"rcAnswerRows": mapping}
    for role in ("lcBook", "rcBook"):
        groups, bounds = book_parts(books[role], args.test)
        book_groups.update(
            {part: {"sourceId": books[role]["sourceId"], "pages": pages} for part, pages in groups.items()}
        )
        mappings[role] = bounds
    write(out / "mapping_candidates.json", mappings)
    ocrs = {}
    for section in ("LC", "RC"):
        ocrs[section] = ocr(sources[section], out, scale=args.scale, save_previews=args.page_previews)
        write(out / "raw" / (section + "_ocr.json"), ocrs[section])
    test = build(args.test, sources, ocrs, keys)
    stage_errors = []
    if args.stage != "extract":
        from llm import load_api_env, structure_part

        if load_api_env(args.env_file):
            for index in range(7):
                part = index + 1
                print(f"Structure Part {part}", flush=True)
                try:
                    structured = structure_part(
                        args.test,
                        part,
                        test.parts[index].model_dump(),
                        exam_for_part(ocrs["LC" if part <= 4 else "RC"], part),
                        book_groups.get(part),
                        [
                            k
                            for k in keys
                            if (1, 7, 32, 71, 101, 131, 147)[index]
                            <= k["questionNumber"]
                            <= (6, 31, 70, 100, 130, 146, 200)[index]
                        ],
                    )
                    # Authoritative extracted answer candidates override model reasoning.
                    key_map = {}
                    for k in keys:
                        key_map.setdefault(k["questionNumber"], []).append(k)
                    for item in structured.items:
                        for q in item.questions:
                            candidates = key_map.get(q.questionNumber, [])
                            if len(candidates) == 1:
                                q.correctKey = candidates[0]["key"]
                                q.answerRefs = [Ref.model_validate(r) for r in candidates[0]["sourceRefs"]]
                            elif q.correctKey is not None:
                                # Book answers must retain source evidence, else null.
                                if not q.answerRefs:
                                    q.correctKey = None
                    test.parts[index] = structured
                except Exception as error:
                    stage_errors.append(
                        {
                            "stage": "structure",
                            "part": part,
                            "errorType": type(error).__name__,
                            "message": "Part failed; inspect retained raw response/cache",
                        }
                    )
                write(out / "draft.json", test.model_dump())
            if args.repair:
                from repair import repair

                try:
                    repair(test, sources, ocrs, book_groups, keys)
                except Exception as error:
                    stage_errors.append(
                        {
                            "stage": "repair",
                            "errorType": type(error).__name__,
                            "message": "Targeted source repair failed; unresolved content remains blocked",
                        }
                    )
        else:
            stage_errors.append({"stage": "structure", "message": "DEEPSEEK_API_KEY missing"})
        if not args.no_audio:
            from audio import align, transcribe

            try:
                asr = transcribe(sources["audio"], args.asr_model, args.device, args.compute)
                write(out / "raw/asr.json", asr)
                matches = align(
                    [i for p in test.parts[:4] for i in p.items], asr["segments"], test.audioAssetId
                )
                # Validate assigned ranges (Pydantic assignment is explicit here).
                test = Test.model_validate(test.model_dump())
                write(out / "alignment_candidates.json", matches)
            except Exception as error:
                stage_errors.append(
                    {
                        "stage": "audio",
                        "errorType": type(error).__name__,
                        "message": "ASR/alignment failed; rerun with cache",
                    }
                )
    from book_tables import enrich, rc_solutions
    from media import crop_media

    if mapping:
        solutions = rc_solutions(sources["rcBook"], mapping["startPage"], mapping["endPage"])
        write(out / "raw/rc_solutions.json", solutions)
        enrich(test, solutions)

    for part in test.parts:
        if part.partNumber == 6:
            for item in part.items:
                for q in item.questions:
                    nodes = [
                        node
                        for p in item.passages
                        for b in p.blocks
                        for node in b.inline
                        if node.type == "blank" and node.questionId == q.id
                    ]
                    if len(nodes) == 1:
                        q.targetIds = [nodes[0].id]

    crop_media(test, out)
    write(out / "draft.json", test.model_dump())
    write(out / "stage_errors.json", stage_errors)
    report = validate(test, sources)
    write(out / "validation.json", report)
    review(args.test)
    print(
        json.dumps(
            {
                "draft": str(out / "draft.json"),
                "answers": report["answerCount"],
                "questions": report["questionCount"],
                "issues": report["issueCounts"],
                "stageErrors": stage_errors,
            },
            ensure_ascii=False,
            indent=2,
        )
    )
    return 1 if stage_errors else 0


def review(number):
    out = test_dir(number)
    test = Test.model_validate(read(out / "draft.json"))
    report = read(out / "validation.json")
    rows = []
    for part in test.parts:
        rows.append(f"<h2>Part {part.partNumber}</h2>")
        for item in part.items:
            import os

            from extractors import source_path

            refs = set()

            def collect_refs(value):
                if isinstance(value, dict):
                    if value.get("sourceId", "").endswith(".pdf") and value.get("pageNumber"):
                        refs.add((value["sourceId"], value["pageNumber"]))
                    for v in value.values():
                        collect_refs(v)
                elif isinstance(value, list):
                    for v in value:
                        collect_refs(v)

            collect_refs(item.model_dump())
            source_links = " ".join(
                f'<a target="_blank" href="{html.escape(os.path.relpath(source_path(s), out).replace(chr(92), chr(47)))}#page={p}">{html.escape(s)} p{p}</a>'
                for s, p in sorted(refs)
            )
            audio_button = ""
            if item.itemAudio:
                audio_button = f'<button onclick="playRange({item.itemAudio.startTime},{item.itemAudio.endTime})">Listen to this item</button>'
            crops = "".join(
                f'<img loading="lazy" src="{html.escape(b.assetId)}">' for b in item.media if b.assetId
            )
            rows.append(
                f"<section><h3>{html.escape(item.id)}</h3>{source_links}{audio_button}{crops}<pre>{html.escape(json.dumps(item.model_dump(), ensure_ascii=False, indent=2))}</pre></section>"
            )
    pages = sorted((out / "pages").glob("*.png"))
    images = (
        "".join(
            f'<figure><figcaption>{html.escape(p.name)}</figcaption><img loading="lazy" src="{p.relative_to(out).as_posix()}"></figure>'
            for p in pages
        )
        if pages
        else "<p style='padding:20px;color:#888'>Page previews disabled (run with --page-previews to generate scan images).</p>"
    )
    text = '<!doctype html><meta charset="utf-8"><title>TOEIC source review</title><style>body{font:16px system-ui;margin:24px}main{display:grid;grid-template-columns:1fr 1fr;gap:20px}article{height:85vh;overflow:auto}pre{white-space:pre-wrap}img{max-width:100%}section{border-bottom:1px solid #aaa}audio{width:100%}</style>'
    # Audio stays inside workspace; no remote publication or copies needed.
    import os

    from extractors import source_path

    audio = os.path.relpath(source_path(test.audioAssetId), out).replace("\\", "/")
    text += (
        f'<h1>Test {number:02d} — DRAFT / needs review</h1><p>Payload hash: {report["payloadHash"]}</p><audio controls src="{html.escape(audio)}"></audio><main><article>'
        + "".join(rows)
        + "</article><article>"
        + images
        + "</article></main>"
    )
    text += """<script>let stopAt=null;const player=document.querySelector('audio');function playRange(start,end){stopAt=end;player.currentTime=start;player.play()}player.addEventListener('timeupdate',()=>{if(stopAt!==null && player.currentTime>=stopAt){player.pause();stopAt=null}});</script>"""
    (out / "review.html").write_text(text, encoding="utf-8")
    print(f"Review: {out / 'review.html'}", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "command", choices=["inventory", "run", "review", "validate", "release", "schema", "correct"]
    )
    parser.add_argument("--corrections", type=Path)
    parser.add_argument("--test", type=int, choices=range(1, 11), default=2)
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--stage", choices=["extract", "all"], default="all")
    parser.add_argument("--no-audio", action="store_true")
    parser.add_argument(
        "--repair",
        action="store_true",
        help="Second full-page OCR and source repair for unresolved layout/blanks",
    )
    parser.add_argument("--env-file", type=Path)
    parser.add_argument("--category-slug", default="ets-2024", help="MinIO category folder")
    parser.add_argument("--test-slug", help="MinIO test folder (default: <category>-test-NN)")
    parser.add_argument(
        "--output-root",
        type=Path,
        default=BASE / "output",
        help="Root of the final toeic/<category>/<test> package",
    )
    parser.add_argument(
        "--page-previews",
        action="store_true",
        help="Render full-page PNG previews (off by default to save disk and speed up execution)",
    )
    parser.add_argument("--scale", type=float, default=1.8)
    parser.add_argument("--asr-model", default="base.en")
    parser.add_argument("--device", choices=["cpu", "cuda"], default="cuda")
    parser.add_argument("--compute", default="float16")
    args = parser.parse_args()
    if args.command == "inventory":
        print(f"Inventoried {len(inventory()['files'])} files")
        return 0
    if args.command == "schema":
        write(BASE / "schema.json", Test.model_json_schema())
        return 0
    if args.command == "run":
        from filelock import FileLock

        test_dir(args.test).mkdir(parents=True, exist_ok=True)
        with FileLock(test_dir(args.test) / "run.lock", timeout=0):
            return run(args)
    if args.command == "review":
        review(args.test)
        return 0
    out = test_dir(args.test)
    test = Test.model_validate(read(out / "draft.json"))
    if args.command == "correct":
        from corrections import apply

        if not args.corrections:
            parser.error("correct requires --corrections FILE")
        overlay = read(args.corrections)
        test = apply(test.model_dump(), overlay)
        write(out / "corrections" / (digest(overlay) + ".json"), overlay)
        write(out / "draft.json", test.model_dump())
    # Re-read sources from disk: a changed input must invalidate validation.
    sources = select_sources(inventory(args.test), args.test)
    reviews = read(out / "reviews.json") if (out / "reviews.json").exists() else None
    report = validate(test, sources, reviews)
    errors = read(out / "stage_errors.json") if (out / "stage_errors.json").exists() else []
    if errors:
        report["releaseReady"] = False
        report["stageErrors"] = errors
    write(out / "validation.json", report)
    if args.command == "release":
        if not report["releaseReady"]:
            print(f"Release blocked: see {out / 'validation.json'}")
            return 2
        from content_export import export_content, package_content

        content = export_content(test, out / "release", release=True)
        package = package_content(
            content,
            args.output_root,
            args.category_slug,
            args.test_slug or f"{args.category_slug}-test-{args.test:02d}",
            ROOT / "Topic Test" / "Toeic 2024",
            out,
        )
        print(f"MinIO-ready package: {package}")
    return 0 if report["releaseReady"] else 2


if __name__ == "__main__":
    sys.exit(main())
