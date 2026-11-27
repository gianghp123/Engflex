"""Convert a validated Study4 IELTS archive to the nested Go IELTS document."""

import hashlib
import re
from pathlib import Path

from study4_content import collapsed, has_class, nodes, one, question_number, text, transcript_lines, walk
from study4_ielts_source import source_skill, validate_ielts_export

SAFE_TAGS = set("div p span b strong i em u br ul ol li table thead tbody tfoot tr td th img h1 h2 h3 h4 h5 h6 sup sub".split())


def context_block(node, asset_ids):
    if "text" in node:
        return {"tag": "text", "text": node["text"]}
    tag = node.get("tag")
    if tag not in SAFE_TAGS:
        raise ValueError(f"Unsupported source context tag: {tag}")
    block = {"tag": tag}
    if tag == "img":
        block["assetId"] = asset_ids[node["asset"]]
        if node.get("alt"):
            block["alt"] = node["alt"]
    for source, target in (("colspan", "colSpan"), ("rowspan", "rowSpan")):
        if source in node:
            span = int(node[source])
            if span < 1:
                raise ValueError("Invalid table cell span")
            block[target] = span
    if node.get("children"):
        block["children"] = [context_block(child, asset_ids) for child in node["children"]]
    return block


def answer_variants(source):
    """Expand only explicit optional fragments and source alternative notation."""
    alternatives = re.split(r"\s*\[OR\]\s*", source, flags=re.I)
    values = []
    for alternative in alternatives:
        if "/" in alternative:
            match = re.fullmatch(r"([^/\s]+)/([^/\s]+)(\s+.*)?", alternative)
            if not match:
                raise ValueError(f"Ambiguous answer alternative: {source}")
            pending = [match[1] + (match[3] or ""), match[2] + (match[3] or "")]
        else:
            pending = [alternative]
        while pending:
            candidate = pending.pop(0)
            optional = re.search(r"\(([^()]+)\)", candidate)
            if optional:
                pending.extend([candidate[:optional.start()] + suffix + candidate[optional.end():] for suffix in ("", optional[1])])
                continue
            value = " ".join(candidate.split())
            if not value or any(c in value for c in "()[]/"):
                raise ValueError(f"Unresolved source answer notation: {source}")
            if value not in values:
                values.append(value)
    return values


def group_type(instruction):
    value = instruction.lower()
    if "yes" in value and "not given" in value and "views" in value:
        return "yes_no_not_given"
    if "true" in value and "not given" in value:
        return "true_false_not_given"
    for phrase, kind in (("label the diagram", "diagram_labeling"), ("correct heading", "matching_headings"),
                         ("complete the table", "table_completion"), ("complete the notes", "note_completion"),
                         ("complete the summary", "summary_completion"), ("answer the questions", "short_answer"),
                         ("choose the correct letter", "multiple_choice"), ("which paragraph", "matching_information")):
        if phrase in value:
            return kind
    if any(phrase in value for phrase in ("match each", "choose four answers", "choose six answers", "who will be responsible")):
        return "matching"
    raise ValueError(f"Unrecognized IELTS question group: {instruction[:100]}")


def shared_options(contexts, kind):
    options = []
    for ctx in contexts:
        for node in walk(ctx):
            if node.get("tag") not in {"p", "li"}:
                continue
            match = re.fullmatch(r"([A-Z]|[ivx]+)\.\s+(.+)", text(node), re.S)
            if match:
                options.append({"key": match[1], "text": match[2].strip()})
    if not options and kind in {"diagram_labeling", "matching_information"}:
        match = re.search(r"\b([A-Z])\s*[-–]\s*([A-Z])\b", "\n".join(text(c) for c in contexts))
        if not match:
            raise ValueError("Missing explicit letter range in source instructions")
        options = [{"key": chr(code), "text": chr(code)} for code in range(ord(match[1]), ord(match[2]) + 1)]
    if kind in {"yes_no_not_given", "true_false_not_given"}:
        values = ["YES", "NO", "NOT GIVEN"] if kind.startswith("yes") else ["TRUE", "FALSE", "NOT GIVEN"]
        options = [{"key": value, "text": value} for value in values]
    if len({option["key"] for option in options}) != len(options):
        raise ValueError("Duplicate shared option key")
    return options


def make_question(wrapper, prefix, kind, options):
    source = one(nodes(wrapper, "question-wrapper"), "IELTS question")
    number = question_number(source)
    answer = one([text(n).split(":", 1)[1].strip() for n in walk(source)
                  if has_class(n, "text-success") and text(n).startswith("Đáp án đúng:")], "explicit IELTS answer")
    choices = []
    for choice in nodes(source, "form-check"):
        control = one([n for n in walk(choice) if n.get("tag") == "input"], "option input")
        label = text(one([n for n in walk(choice) if n.get("tag") == "label"], "option label"))
        key = control["value"]
        if label == key and kind in {"yes_no_not_given", "true_false_not_given"}:
            choices.append({"key": key, "text": label})
            continue
        if not label.startswith(key + "."):
            raise ValueError("Source option label/key mismatch")
        choices.append({"key": key, "text": label[len(key) + 1:].strip()})
    choices = choices or options
    question = {"id": f"{prefix}_q{number}", "questionNumber": number, "sourceAnswer": answer}
    stems = nodes(source, "question-text")
    if stems:
        question["stem"] = text(one(stems, "question stem"))
    if kind in {"short_answer", "note_completion", "table_completion", "summary_completion"}:
        question["correctAnswers"] = answer_variants(answer)
    else:
        if answer not in {option["key"] for option in choices}:
            raise ValueError(f"Question {number}: correct key missing from options")
        question.update({"options": choices, "correctKey": answer})
    explanations = nodes(wrapper, "question-explanation-wrapper")
    if explanations:
        question["explanation"] = text(collapsed(one(explanations, "question explanation")))
    return question


def outer_contexts(part):
    if has_class(part, "question-group-wrapper"):
        return []
    if has_class(part, "context-content"):
        return [part]
    return [ctx for child in part.get("children", []) for ctx in outer_contexts(child)]


def build_content(data, test_root, *, category_slug, test_slug, sub_type):
    validate_ielts_export(data)
    if sub_type not in {"academic", "general_training"}:
        raise ValueError("IELTS subtype must be explicitly selected")
    for slug in (category_slug, test_slug):
        if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", slug):
            raise ValueError("Expected lowercase hyphenated category/test slugs")
    test_root = Path(test_root).resolve()
    prefix = f"{category_slug}/{test_slug}"
    parts = [part for root in data["content"] for part in nodes(root, "test-questions-wrapper")]
    labels = {}
    for index, part in enumerate(parts, 1):
        for node in walk(part):
            if node.get("asset"):
                labels.setdefault(node["asset"], f"section-{index}")
    assets, prepared, asset_ids = [], [], {}
    for index, asset in enumerate(data["assets"], 1):
        if asset.get("status") != "downloaded" or not asset.get("localPath"):
            raise ValueError("IELTS publisher requires retained, downloaded local media")
        path = (test_root / asset["localPath"]).resolve()
        if not path.is_relative_to(test_root) or not path.is_file():
            raise ValueError("Media path is outside the source archive or missing")
        with path.open("rb") as stream:
            if hashlib.file_digest(stream, "sha256").hexdigest() != asset["sha256"]:
                raise ValueError("Source media SHA-256 mismatch")
        folder = "audio" if asset["kind"] == "audio" else "images"
        name = f"{test_slug}-{labels[asset['id']]}"
        if asset["kind"] == "image":
            name += f"-image-{index:02d}"
        key = f"{prefix}/{folder}/{name}{path.suffix.lower()}"
        identity = f"ielts/{key}"
        asset_ids[asset["id"]] = identity
        assets.append({"sourceId": identity, "type": asset["kind"], "sha256": asset["sha256"]})
        prepared.append({"path": path, "key": key, "sha256": asset["sha256"], "contentType": asset["contentType"]})
    content = {"schemaVersion": "1.0.0", "testId": test_slug,
               "title": data["source"]["title"].removeprefix("Đáp án chi tiết: ").removesuffix(" - STUDY4"),
               "skill": source_skill(data), "subType": sub_type, "sections": [], "assets": assets}
    for index, part in enumerate(parts, 1):
        section = {"sectionNumber": index, "title": f"Section {index}", "passages": [], "questionGroups": []}
        outer = outer_contexts(part)
        for ctx in outer:
            if has_class(ctx, "context-audio"):
                ids = {asset_ids[n["asset"]] for n in walk(ctx) if n.get("asset")}
                section["audioAssetId"] = one(list(ids), "section audio")
            elif not has_class(ctx, "context-transcript"):
                paragraphs = [n for n in walk(ctx) if n.get("tag") in {"p", "li"}]
                section["passages"].append({"id": f"{test_slug}_s{index}_passage",
                    "content": [{"id": f"s{index}_p{j}", "text": text(n)} for j, n in enumerate(paragraphs, 1)]})
        if content["skill"] == "listening":
            english, vietnamese = transcript_lines(part)
            if not english or vietnamese:
                raise ValueError("Missing or unsupported bilingual transcript; preserve source for review")
            section["transcript"] = [{"id": f"s{index}_t{j}", "text": line} for j, line in enumerate(english, 1)]
        for group_index, source in enumerate(nodes(part, "question-group-wrapper"), 1):
            contexts = nodes(source, "context-content")
            instruction = "\n".join(text(ctx) for ctx in contexts)
            kind = group_type(instruction)
            options = shared_options(contexts, kind)
            wrappers = nodes(source, "question-item-wrapper")
            if not wrappers:
                raise ValueError("Missing question item/explanation boundaries")
            group = {"id": f"{test_slug}_s{index}_g{group_index}", "type": kind,
                     "instructions": instruction.split("\n", 1)[0],
                     "context": [context_block(ctx, asset_ids) for ctx in contexts],
                     "questions": [make_question(w, test_slug, kind, options) for w in wrappers]}
            section["questionGroups"].append(group)
        content["sections"].append(section)
    questions = [q for s in content["sections"] for g in s["questionGroups"] for q in g["questions"]]
    if [q["questionNumber"] for q in questions] != list(range(1, 41)):
        raise ValueError("Conversion changed source question coverage")
    return content, prepared
