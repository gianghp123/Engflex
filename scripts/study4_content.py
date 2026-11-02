"""Deterministic Study4 DOM adapter for the existing TOEIC 0.1.0 DB document.

Only source answer labels/explicit grading classes are accepted. Never infer answers from checked inputs,
translations, explanations, ASR, or question numbers. Unsupported layout fails
closed so it remains available in raw.json for review.
"""

import hashlib
import json
import math
import re

BOUNDS = ((1, 6), (7, 31), (32, 70), (71, 100), (101, 130), (131, 146), (147, 200))


def digest(value):
    return hashlib.sha256(
        json.dumps(value, ensure_ascii=False, sort_keys=True, allow_nan=False).encode()
    ).hexdigest()


def walk(node):
    yield node
    for child in node.get("children", []):
        yield from walk(child)


def has_class(node, name):
    return name in node.get("class", "").split()


def nodes(node, name):
    return [child for child in walk(node) if has_class(child, name)]


def context_sources(node):
    """Keep English source contexts separate from their translated subtrees."""
    if has_class(node, "context-transcript"):
        return []
    if has_class(node, "context-content"):
        return [node]
    return [source for child in node.get("children", []) for source in context_sources(child)]


def text(node):
    if "text" in node:
        return node["text"]
    if node.get("tag") == "br":
        return "\n"
    separator = "\n" if node.get("tag") in {"div", "ul", "ol", "table", "tbody", "tr"} else " "
    return separator.join(filter(None, (text(child) for child in node.get("children", [])))).strip()


def one(values, label):
    if len(values) != 1:
        raise ValueError(f"Expected exactly one {label}; found {len(values)}")
    return values[0]


def question_number(node):
    number = text(one(nodes(node, "question-number"), "question number")).strip()
    if not number.isdigit():
        raise ValueError(f"Invalid question number: {number!r}")
    return int(number)


def exam_items(data):
    if data.get("page_kind") != "exam":
        raise ValueError("A complete question/solution page is required; landing pages are not exams")
    parts = []
    for root in data.get("content", []):
        parts.extend(nodes(root, "test-questions-wrapper"))
    if len(parts) != 7:
        raise ValueError(f"Expected seven source parts, found {len(parts)}")
    result = []
    all_numbers = []
    for index, part in enumerate(parts):
        wrappers = []

        def collect(node):
            if has_class(node, "question-group-wrapper") or has_class(node, "question-item-wrapper"):
                wrappers.append(node)
                return
            for child in node.get("children", []):
                collect(child)

        collect(part)
        numbers = [question_number(q) for wrapper in wrappers for q in nodes(wrapper, "question-wrapper")]
        lo, hi = BOUNDS[index]
        if numbers != list(range(lo, hi + 1)):
            raise ValueError(f"Part {index + 1}: missing, duplicate or unordered question numbers")
        all_numbers.extend(numbers)
        result.append(wrappers)
    if all_numbers != list(range(1, 201)):
        raise ValueError("Expected questions 1–200 exactly once")
    return result


def collapsed(node):
    return one([child for child in walk(node) if has_class(child, "collapse")], "collapsed source content")


def transcript_lines(wrapper):
    contexts = nodes(wrapper, "context-transcript")
    if not contexts:
        return [], []
    source = collapsed(one(contexts, "transcript"))
    # Both observed source layouts put English before the explicit translation
    # heading. Paragraphs are the source's utterance/option boundaries.
    english, vietnamese = [], []
    translated = False
    for child in walk(source):
        if child.get("tag") not in {"p", "li"}:
            continue
        line = text(child).strip()
        if not line:
            continue
        if line.startswith("Dịch nghĩa:"):
            translated = True
            remainder = line.removeprefix("Dịch nghĩa:").strip()
            if remainder:
                vietnamese.append(remainder)
        else:
            (vietnamese if translated else english).append(line)
    return english, vietnamese


def make_transcript(wrapper, part, prefix):
    english, vietnamese = transcript_lines(wrapper)
    if not english:
        raise ValueError(f"{prefix}: missing English listening transcript")
    if vietnamese and len(vietnamese) != len(english):
        raise ValueError(f"{prefix}: transcript translation boundaries need review")
    result = []
    for index, line in enumerate(english):
        option = re.match(r"^\(([A-D])\)\s*(.+)$", line, re.S)
        speaker = re.match(r"^([MW]\d*):\s*(.+)$", line, re.S)
        translated = vietnamese[index] if vietnamese else None
        if option:
            role, key, body = "option", option[1], option[2]
            if translated:
                translated = re.sub(r"^\([A-D]\)\s*", "", translated)
        else:
            role, key, body = ("prompt" if part <= 2 else "utterance"), None, speaker[2] if speaker else line
            if translated and speaker:
                translated = re.sub(r"^[MW]\d*:\s*", "", translated)
        result.append(
            {
                "id": f"{prefix}_transcript_{index + 1}",
                "role": role,
                "speaker": speaker[1] if speaker else None,
                "optionKey": key,
                "text": body,
                "vietnameseText": translated,
            }
        )
    if part <= 2 and [t["optionKey"] for t in result if t["role"] == "option"] != list(
        "ABC" if part == 2 else "ABCD"
    ):
        raise ValueError(f"{prefix}: missing or ambiguous spoken options")
    return result


def make_question(wrapper, part, prefix):
    source = one(nodes(wrapper, "question-wrapper"), "question")
    number = question_number(source)
    options = []
    for choice in nodes(source, "form-check"):
        radio = one([n for n in walk(choice) if n.get("tag") == "input"], "choice input")
        key = radio.get("value")
        label = text(one([n for n in walk(choice) if n.get("tag") == "label"], "choice label"))
        match = re.fullmatch(r"([A-D])\.\s*(.*)", label, re.S)
        if not match or match[1] != key:
            raise ValueError(f"Question {number}: choice label does not match its key")
        options.append({"key": key, "text": match[2].strip() or None})
    if [o["key"] for o in options] != list("ABC" if part == 2 else "ABCD"):
        raise ValueError(f"Question {number}: missing/duplicate/out-of-order choices")
    correct = []
    for n in walk(source):
        if has_class(n, "text-success"):
            match = re.fullmatch(r"Đáp án đúng:\s*([A-D])", text(n))
            if match:
                correct.append(match[1])
    graded = [n["value"] for n in walk(source) if n.get("tag") == "input" and has_class(n, "correct")]
    if len(correct) > 1 or len(graded) > 1 or correct and graded and correct != graded:
        raise ValueError(f"Question {number}: conflicting source answer evidence")
    answer = one(correct or graded, f"source answer for question {number}")
    if answer not in [o["key"] for o in options]:
        raise ValueError(f"Question {number}: answer absent from options")
    stems = nodes(source, "question-text")
    stem = text(one(stems, "question stem")) if stems else None
    explanations = nodes(wrapper, "question-explanation-wrapper")
    explanation = text(collapsed(one(explanations, "explanation"))) if explanations else None
    return {
        "id": f"{prefix}_q{number}",
        "questionNumber": number,
        "stem": stem,
        "options": options,
        "correctKey": answer,
        "explanation": explanation,
        "vietnameseText": None,
        "vocabulary": [],
        "targetIds": [],
    }


def content_blocks(source, prefix, assets, questions, part):
    result = []
    blank_seen = []

    def add(kind, **fields):
        block = {"id": f"{prefix}_block_{len(result) + 1}", "type": kind, "inline": [], **fields}
        result.append(block)
        return block

    def paragraph(node):
        value = text(node)
        if not value:
            return
        block = add(
            "heading"
            if node.get("tag", "").startswith("h") and node.get("tag") in {"h1", "h2", "h3", "h4"}
            else "paragraph",
            text=value,
        )
        # Keep marked text and source blanks in reading order. A source marker
        # is required; never append fabricated blanks to satisfy the validator.
        if part == 6:
            pattern = r"[-–—]*\((1[34]\d)\)[-–—]*"
            position = 0
            for match in re.finditer(pattern, value):
                number = int(match[1])
                question = one([q for q in questions if q["questionNumber"] == number], "blank question")
                if value[position : match.start()]:
                    block["inline"].append(
                        {
                            "id": f"{block['id']}_text_{len(block['inline'])}",
                            "type": "text",
                            "text": value[position : match.start()],
                        }
                    )
                blank_id = f"{prefix}_blank_{number}"
                block["inline"].append({"id": blank_id, "type": "blank", "questionId": question["id"]})
                question["targetIds"].append(blank_id)
                blank_seen.append(number)
                position = match.end()
            if position:
                if value[position:]:
                    block["inline"].append(
                        {
                            "id": f"{block['id']}_text_{len(block['inline'])}",
                            "type": "text",
                            "text": value[position:],
                        }
                    )
                block.pop("text")
        elif any(n.get("tag") == "u" for n in walk(node)):
            # Preserve underline semantics without introducing HTML in JSONB.
            def inline(n, marked=False):
                if n.get("tag") == "br":
                    block["inline"].append(
                        {"id": f"{block['id']}_inline_{len(block['inline'])}", "type": "text", "text": "\n"}
                    )
                if "text" in n:
                    value = n["text"]
                    if block["inline"] and not block["inline"][-1]["text"].endswith("\n"):
                        value = " " + value
                    block["inline"].append(
                        {
                            "id": f"{block['id']}_inline_{len(block['inline'])}",
                            "type": "markedText" if marked else "text",
                            "text": value,
                        }
                    )
                for child in n.get("children", []):
                    inline(child, marked or n.get("tag") == "u")

            inline(node)
            block.pop("text")

    def visit(node):
        tag = node.get("tag")
        if tag == "table":
            rows = []
            for row in walk(node):
                if row.get("tag") == "tr":
                    cells = [cell for cell in row.get("children", []) if cell.get("tag") in {"td", "th"}]
                    if any(
                        str(cell.get(attr, "1")) != "1" for cell in cells for attr in ("colspan", "rowspan")
                    ):
                        raise ValueError(f"{prefix}: merged table cells need review")
                    rows.append([text(cell) for cell in cells])
            if not rows or not all(rows):
                raise ValueError(f"{prefix}: empty table")
            add("table", rows=rows)
        elif tag == "img":
            asset = assets.get(node.get("asset"))
            if asset is None:
                raise ValueError(f"{prefix}: unresolved image asset")
            add("image", assetId=asset["sourceId"])
        elif tag in {"p", "li", "h1", "h2", "h3", "h4"}:
            if any(n.get("tag") in {"img", "table"} for n in walk(node)):
                for child in node.get("children", []):
                    visit(child)
            else:
                paragraph(node)
        elif "text" in node:
            paragraph(node)
        else:
            for child in node.get("children", []):
                visit(child)

    visit(source)
    if part == 6 and sorted(blank_seen) != sorted(q["questionNumber"] for q in questions):
        raise ValueError(f"{prefix}: missing or duplicate source blanks")
    return result


def passage_sources(context, expected):
    groups = [[]]

    def split(node):
        children = node.get("children", [])
        separators = [
            i for i, child in enumerate(children) if re.fullmatch(r"[-–—]{3,}", text(child).strip())
        ]
        if separators:
            for child in children:
                if re.fullmatch(r"[-–—]{3,}", text(child).strip()):
                    groups.append([])
                else:
                    groups[-1].append(child)
            return True
        meaningful = [child for child in children if text(child) or child.get("asset")]
        if len(meaningful) == 1 and split(meaningful[0]):
            return True
        return False

    if expected == 1:
        return [context]
    if not split(context) or len(groups) != expected or not all(groups):
        raise ValueError(f"Expected {expected} explicit source passage boundaries")
    return [{"tag": "div", "children": children} for children in groups]


def build_content(data, *, collection_id, test_number, title=None, alignment=None):
    parts = exam_items(data)
    match = re.search(r"/tests/(\d+)/", data.get("source", {}).get("url", ""))
    if not match or not collection_id.strip() or test_number < 1:
        raise ValueError("Explicit collection/test identity and a Study4 test URL are required")
    test_id = f"study4_{match[1]}"
    assets = {}
    for asset in data.get("assets", []):
        if asset.get("status") not in {"uploaded", "downloaded"} or not re.fullmatch(
            r"[a-f0-9]{64}", asset.get("sha256", "")
        ):
            raise ValueError("Every source asset must have verified bytes and a SHA-256")
        if asset["id"] in assets:
            raise ValueError("Duplicate DOM asset ID")
        assets[asset["id"]] = {"sourceId": f"exams/toeic/{asset['sha256']}", "sha256": asset["sha256"]}
    audio = one([a for a in data.get("assets", []) if a.get("kind") == "audio"], "full listening audio")
    audio_id = assets[audio["id"]]["sourceId"]
    if (
        alignment is None
        or alignment.get("status") != "complete"
        or alignment.get("audioSha256") != audio["sha256"]
        or alignment.get("exportHash") != digest(data["content"])
    ):
        raise ValueError("A complete alignment bound to this DOM and audio hash is required")
    ranges = {tuple(item["questions"]): item["audioRange"] for item in alignment["items"]}
    content = {
        "schemaVersion": "0.1.0",
        "collectionId": collection_id,
        "testId": test_id,
        "testNumber": test_number,
        "title": title
        or data["source"].get("title", "").removeprefix("Đáp án chi tiết: ").removesuffix(" - STUDY4"),
        "audioAssetId": audio_id,
        "assets": [],
        "parts": [],
    }
    previous_end = 0
    for part_number, wrappers in enumerate(parts, 1):
        part = {
            "partNumber": part_number,
            "section": "listening" if part_number <= 4 else "reading",
            "items": [],
        }
        for wrapper in wrappers:
            question_wrappers = (
                nodes(wrapper, "question-item-wrapper")
                if has_class(wrapper, "question-group-wrapper")
                else [wrapper]
            )
            questions = [make_question(q, part_number, test_id) for q in question_wrappers]
            first = questions[0]["questionNumber"]
            prefix = f"{test_id}_item_{first}"
            item = {
                "id": prefix,
                "type": "group" if part_number in {3, 4, 6, 7} else "single",
                "questions": questions,
                "passages": [],
                "transcript": [],
                "itemAudio": None,
                "stimulusAudio": None,
                "media": [],
            }
            if part_number <= 4:
                item["transcript"] = make_transcript(wrapper, part_number, prefix)
                numbers = tuple(q["questionNumber"] for q in questions)
                audio_range = ranges.get(numbers)
                if not audio_range:
                    raise ValueError(f"{prefix}: missing audio range")
                start, end = audio_range["startTime"], audio_range["endTime"]
                if (
                    not all(isinstance(v, (int, float)) and math.isfinite(v) for v in (start, end))
                    or not previous_end <= start < end <= alignment["audioDuration"]
                ):
                    raise ValueError(f"{prefix}: invalid/overlapping audio range")
                item["itemAudio"] = {**audio_range, "assetId": audio_id}
                previous_end = end
            contexts = context_sources(wrapper)
            if part_number in {6, 7}:
                context = one(contexts, "reading context")
                expected = 1 if part_number == 6 or first <= 175 else 2 if first <= 185 else 3
                translations = nodes(wrapper, "context-transcript")
                translated_sources = None
                if translations:
                    translated = collapsed(one(translations, "reading translation"))
                    children = translated.get("children", [])
                    if not children or text(children[0]).strip() != "Dịch nghĩa:":
                        raise ValueError(f"{prefix}: unknown reading translation layout")
                    translated_sources = passage_sources({"tag": "div", "children": children[1:]}, expected)
                for index, source in enumerate(passage_sources(context, expected), 1):
                    passage_id = f"{prefix}_passage_{index}"
                    blocks = content_blocks(source, passage_id, assets, questions, part_number)
                    if translated_sources is not None:
                        translated_blocks = content_blocks(
                            translated_sources[index - 1], f"{passage_id}_vi", assets, [], 0
                        )
                        if [b["type"] for b in blocks] != [b["type"] for b in translated_blocks]:
                            raise ValueError(f"{passage_id}: translation block boundaries need review")
                        for block, translated_block in zip(blocks, translated_blocks, strict=True):
                            if block["type"] == "table":
                                block["vietnameseText"] = "\n".join(
                                    " | ".join(row) for row in translated_block["rows"]
                                )
                            elif block["type"] != "image":
                                block["vietnameseText"] = translated_block.get("text") or " ".join(
                                    n["text"] for n in translated_block["inline"]
                                )
                    item["passages"].append({"id": passage_id, "format": None, "blocks": blocks})
            else:
                for index, context in enumerate(contexts):
                    item["media"].extend(
                        content_blocks(context, f"{prefix}_media_{index}", assets, questions, part_number)
                    )
            part["items"].append(item)
        content["parts"].append(part)
    references = {audio_id}

    def collect(value):
        if isinstance(value, dict):
            if value.get("assetId"):
                references.add(value["assetId"])
            for child in value.values():
                collect(child)
        elif isinstance(value, list):
            for child in value:
                collect(child)

    collect(content["parts"])
    by_source = {a["sourceId"]: a for a in assets.values()}
    content["assets"] = [by_source[key] for key in sorted(references)]
    return content
