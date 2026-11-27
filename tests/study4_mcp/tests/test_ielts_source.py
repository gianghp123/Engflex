"""Coverage checks use synthetic source content, not copyrighted live fixtures."""

import copy
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[3] / "scripts"))

from study4_ielts_source import source_skill, validate_ielts_export


def fixture(skill):
    sections = []
    offset = 1
    counts = [10] * 4 if skill == "listening" else [13, 13, 14]
    assets = []
    for index, count in enumerate(counts, 1):
        children = []
        if skill == "listening":
            assets.append({"id": f"audio-{index}", "kind": "audio"})
            children.append({"asset": f"audio-{index}"})
        for number in range(offset, offset + count):
            children.append({"class": "question-wrapper", "children": [
                {"class": "question-number", "children": [{"text": str(number)}]},
                {"class": "text-success", "children": [{"text": "Đáp án đúng:example"}]},
            ]})
        sections.append({"class": "test-questions-wrapper", "children": children})
        offset += count
    return {"page_kind": "exam", "source": {"url": f"https://study4.com/tests/9999/ielts-collection-1-{skill}-test-1/results/1/details/"}, "content": sections, "assets": assets}


class IeltsSourceTests(unittest.TestCase):
    def test_both_skills_keep_source_unchanged(self):
        for skill in ("reading", "listening"):
            data = fixture(skill)
            before = copy.deepcopy(data)
            report = validate_ielts_export(data)
            self.assertEqual(report["questions"], 40)
            self.assertEqual(report["sourceAnswers"], 40)
            self.assertFalse(report["releaseReady"])
            self.assertEqual(data, before)

    def test_missing_duplicate_reordered_or_unanswered_question_fails(self):
        for fault in ("missing", "duplicate", "reordered", "unanswered"):
            data = fixture("reading")
            questions = data["content"][0]["children"]
            if fault == "missing":
                questions.pop()
            elif fault == "duplicate":
                questions[1] = copy.deepcopy(questions[0])
            elif fault == "reordered":
                questions.reverse()
            else:
                questions[0]["children"].pop()
            with self.assertRaises(ValueError):
                validate_ielts_export(data)

    def test_missing_section_audio_or_asset_fails(self):
        for fault in ("section", "audio", "asset"):
            data = fixture("listening")
            if fault == "section":
                data["content"].pop()
            elif fault == "audio":
                data["content"][0]["children"].pop(0)
            else:
                data["assets"].pop()
            with self.assertRaises(ValueError):
                validate_ielts_export(data)

    def test_toeic_url_not_classified_as_ielts(self):
        self.assertIsNone(source_skill({"source": {"url": "https://study4.com/tests/9999/toeic-test-1/"}}))

    def test_selected_input_is_not_an_answer(self):
        data = fixture("reading")
        data["content"][0]["children"][0]["children"][-1] = {"tag": "input", "class": "correct", "value": "A"}
        with self.assertRaises(ValueError):
            validate_ielts_export(data)
