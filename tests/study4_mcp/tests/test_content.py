import copy
import json
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[3] / "scripts"))

from align_study4_audio import align_listening_items, get_listening_items_from_dom
from export_study4_content import export_bundle, validate_content
from fixtures import element, full_export, listening_segments, paragraph
from study4_content import build_content, content_blocks, digest, exam_items, nodes


def alignment(data):
    return {
        "status": "complete",
        "runId": "synthetic",
        "audioSha256": data["assets"][0]["sha256"],
        "exportHash": digest(data["content"]),
        "audioDuration": 2000,
        "asrHash": "a" * 64,
        "items": align_listening_items(
            get_listening_items_from_dom(data["content"]), listening_segments(data), "audio.mp3"
        ),
    }


class ContentTests(unittest.TestCase):
    def test_bare_transcript_text_preserves_translation_boundary(self):
        from study4_content import transcript_lines
        for paragraph_wrapped in (True,False):
            with self.subTest(paragraph_wrapped=paragraph_wrapped):
                english={"text":"W: Welcome to the show."}
                translated={"text":"W: Synthetic translation."}
                children=[paragraph(english["text"]) if paragraph_wrapped else english,paragraph("Dịch nghĩa:"),paragraph(translated["text"]) if paragraph_wrapped else translated]
                wrapper=element(cls="context-transcript",children=[element(cls="collapse",children=children)])
                self.assertEqual(([english["text"]],[translated["text"]]),transcript_lines(wrapper))

    def test_explicit_translation_document_labels_keep_passages_separate(self):
        from study4_content import passage_sources,text
        for order in ([1,2],[2,1],[1]):
            with self.subTest(order=order):
                children=[]
                for number in order:
                    children.extend([paragraph(f"TÀI LIỆU {number}:"),element(children=[paragraph(f"Document {number} body")])])
                source=element(children=[element(children=children)])
                if order==[1,2]:
                    result=passage_sources(source,2)
                    self.assertEqual(["Document 1 body","Document 2 body"],[text(p) for p in result])
                else:
                    with self.assertRaises(ValueError):passage_sources(source,2)

    def test_reading_translation_keeps_passage_and_block_references(self):
        data = full_export()
        part = nodes(data["content"][0], "test-questions-wrapper")[6]
        for wrapper in part["children"]:
            context = nodes(wrapper, "context-content")[0]
            translation = element(
                cls="context-content context-transcript",
                children=[
                    element(cls="collapse", children=[paragraph("Dịch nghĩa:"), copy.deepcopy(context)])
                ],
            )
            nodes(wrapper, "context-wrapper")[0]["children"].append(translation)
        content = build_content(data, collection_id="synthetic", test_number=1, alignment=alignment(data))
        blocks = content["parts"][6]["items"][-1]["passages"][0]["blocks"]
        self.assertEqual(blocks[0]["vietnameseText"], blocks[0]["text"])
        self.assertEqual(blocks[1]["vietnameseText"], "Cell")

    def test_marked_inline_text_does_not_merge_words(self):
        source = element(
            "p",
            children=[
                {"text": "The"},
                element("u", children=[{"text": "marked word"}]),
                {"text": "remains."},
            ],
        )
        blocks = content_blocks(source, "paragraph", {}, [], 7)
        self.assertEqual("".join(n["text"] for n in blocks[0]["inline"]), "The marked word remains.")
        self.assertEqual(blocks[0]["inline"][1]["type"], "markedText")

    def test_nested_graph_passes_real_go_parser(self):
        data = full_export()
        content = build_content(data, collection_id="synthetic", test_number=1, alignment=alignment(data))
        validate_content(content)
        self.assertEqual([len(p["items"]) for p in content["parts"]], [6, 25, 13, 10, 30, 4, 15])
        part6 = content["parts"][5]["items"][0]
        for q in part6["questions"]:
            self.assertEqual(len(q["targetIds"]), 1)
            blank = next(
                n for b in part6["passages"][0]["blocks"] for n in b["inline"] if n["id"] == q["targetIds"][0]
            )
            self.assertEqual(blank["questionId"], q["id"])
        part7 = content["parts"][6]["items"]
        self.assertEqual([len(i["passages"]) for i in part7[-5:]], [2, 2, 3, 3, 3])
        self.assertEqual(part7[-1]["passages"][0]["blocks"][-1]["rows"], [["Cell"]])
        self.assertEqual(content["parts"][2]["items"][0]["transcript"][0]["text"], "Source utterance.")
        self.assertEqual(content["parts"][2]["items"][0]["transcript"][0]["vietnameseText"], "Bản dịch.")
        self.assertNotIn("content", content)
        self.assertNotIn("source", content)

    def test_reject_missing_duplicate_reordered_questions_and_landing(self):
        for fault in ("missing", "duplicate", "order", "landing"):
            with self.subTest(fault=fault):
                data = full_export()
                part = nodes(data["content"][0], "test-questions-wrapper")[0]
                if fault == "missing":
                    part["children"].pop()
                if fault == "duplicate":
                    part["children"].append(copy.deepcopy(part["children"][0]))
                if fault == "order":
                    part["children"].reverse()
                if fault == "landing":
                    data["page_kind"] = "landing"
                with self.assertRaises(ValueError):
                    exam_items(data)

    def test_reject_unresolved_or_conflicting_answers_and_blank_layout(self):
        for fault in ("missing_answer", "conflicting_answer", "blank", "passages", "merged_table"):
            with self.subTest(fault=fault):
                data = full_export()
                root = data["content"][0]
                if fault == "missing_answer":
                    nodes(root, "text-success")[0]["children"] = []
                if fault == "conflicting_answer":
                    choice = nodes(root, "form-check")[1]["children"][0]
                    choice["class"] += " correct"
                if fault == "blank":
                    part = nodes(root, "test-questions-wrapper")[5]
                    context = nodes(part, "context-content")[0]
                    context["children"] = [{"text": "No source blanks"}]
                if fault == "passages":
                    part = nodes(root, "test-questions-wrapper")[6]
                    context = nodes(part["children"][-1], "context-content")[0]
                    context["children"] = [{"text": "No source separators"}]
                if fault == "merged_table":
                    from study4_content import walk

                    cell = next(n for n in walk(root) if n.get("tag") == "td")
                    cell["colspan"] = "2"
                with self.assertRaises(ValueError):
                    build_content(data, collection_id="synthetic", test_number=1, alignment=alignment(data))

    def test_stale_alignment_is_rejected(self):
        data = full_export()
        bound = alignment(data)
        for key in ("audioSha256", "exportHash", "status"):
            with self.subTest(key=key):
                stale = {**bound, key: "stale"}
                with self.assertRaises(ValueError):
                    build_content(data, collection_id="synthetic", test_number=1, alignment=stale)

    def test_preview_and_explicit_reviewed_release(self):
        data = full_export()
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "export.json").write_text(json.dumps(data), encoding="utf-8")
            (root / "audio-alignment.json").write_text(json.dumps(alignment(data)), encoding="utf-8")
            from unittest.mock import patch

            with patch("export_study4_content.validate_content"):
                for reviewer in (None, "source-reviewer"):
                    output = root / ("reviewed" if reviewer else "draft")
                    result = export_bundle(
                        root / "export.json",
                        output,
                        collection_id="synthetic",
                        test_number=1,
                        reviewed_by=reviewer,
                    )
                    provenance = json.loads(result.with_name("provenance.json").read_text(encoding="utf-8"))
                    report = json.loads((output / "validation.json").read_text(encoding="utf-8"))
                    content = json.loads(result.read_text(encoding="utf-8"))
                    self.assertEqual(report["releaseReady"], bool(reviewer))
                    self.assertEqual(provenance["status"], "release" if reviewer else "preview-needs-review")
                    self.assertEqual(provenance["draftHash"], report["payloadHash"])
                    self.assertEqual(provenance["contentHash"], digest(content))
                    with self.assertRaises(FileExistsError):
                        export_bundle(root / "export.json", output, collection_id="synthetic", test_number=1)


if __name__ == "__main__":
    unittest.main()
