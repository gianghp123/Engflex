"""Validate IELTS source coverage without converting or discarding its DOM graph.

This validates a crawl archive, not release readiness or the learner DTO.
Question answers come only from explicit source labels, never learner inputs.
"""

import re
from urllib.parse import urlsplit

from study4_content import has_class, nodes, question_number, text, walk


def source_skill(data):
    source = urlsplit(data.get("source", {}).get("url", ""))
    if source.scheme != "https" or source.hostname not in {"study4.com", "www.study4.com"}:
        return None
    match = re.match(r"^/tests/\d+/ielts-[^/]*-(listening|reading)-[^/]+/", source.path)
    return match[1] if match else None


def validate_ielts_export(data):
    skill = source_skill(data)
    if skill is None or data.get("page_kind") != "exam":
        raise ValueError("A full IELTS Listening or Reading solution page is required")
    parts = [part for root in data.get("content", []) for part in nodes(root, "test-questions-wrapper")]
    expected_parts = 4 if skill == "listening" else 3
    if len(parts) != expected_parts:
        raise ValueError(f"IELTS {skill}: expected {expected_parts} sections, found {len(parts)}")
    numbers, answer_count, audio_ids = [], 0, set()
    asset_map = {asset["id"]: asset for asset in data.get("assets", [])}
    if len(asset_map) != len(data.get("assets", [])):
        raise ValueError("IELTS: duplicate asset identities")
    for index, part in enumerate(parts, 1):
        questions = nodes(part, "question-wrapper")
        section_numbers = [question_number(question) for question in questions]
        if not section_numbers:
            raise ValueError(f"IELTS section {index}: no questions")
        if skill == "listening" and len(section_numbers) != 10:
            raise ValueError(f"IELTS Listening section {index}: expected 10 questions")
        numbers.extend(section_numbers)
        for question in questions:
            labels = [text(node).strip() for node in walk(question) if has_class(node, "text-success")]
            answers = [label.split(":", 1)[1].strip() for label in labels if label.startswith("Đáp án đúng:")]
            if len(answers) != 1 or not answers[0]:
                raise ValueError(f"IELTS question {question_number(question)}: missing or conflicting source answer")
            answer_count += 1
        section_audio = {
            node["asset"] for node in walk(part)
            if node.get("asset") and asset_map.get(node["asset"], {}).get("kind") == "audio"
        }
        if skill == "listening" and len(section_audio) != 1:
            raise ValueError(f"IELTS Listening section {index}: expected one audio asset")
        audio_ids.update(section_audio)
    if numbers != list(range(1, 41)):
        raise ValueError("IELTS: expected questions 1-40 exactly once in order")
    used_assets = {node["asset"] for root in data["content"] for node in walk(root) if node.get("asset")}
    if used_assets != set(asset_map):
        raise ValueError("IELTS: missing or unreferenced asset identities")
    if data.get("media_download_complete") is False:
        raise ValueError("IELTS: incomplete media download")
    if any(asset.get("status") == "failed" for asset in asset_map.values()):
        raise ValueError("IELTS: failed media asset")
    return {
        "skill": skill,
        "sections": len(parts),
        "questions": len(numbers),
        "sourceAnswers": answer_count,
        "audioAssets": len(audio_ids),
        "imageAssets": sum(asset.get("kind") == "image" for asset in asset_map.values()),
        "status": "source-validated",
        "releaseReady": False,
    }
