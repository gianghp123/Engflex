import json
import sys
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch

from fixtures import element, full_export, listening_segments, paragraph

sys.path.insert(0, str(Path(__file__).resolve().parents[3] / "scripts"))

from align_study4_audio import (
    align_listening_items,
    get_listening_items_from_dom,
    normalize_text,
    process_test_alignment,
    spoken_number,
)


class AlignAudioTests(unittest.TestCase):
    def test_explicit_markers_without_asr_number_prefix_use_word_timestamps(self):
        for first,last,label,next_label in [(4,4,"4. Look at the picture","5. Look at the picture"),(71,73,"71. Through 73. Refer to the following announcement","74. 76. Refer to the following talk")]:
            with self.subTest(label=label):
                following=5 if first==4 else 74
                items=[{"id":"first","questions":list(range(first,last+1)),"transcript":"Source utterance"},{"id":"next","questions":[following] if following==5 else [74,75,76],"transcript":"Next source utterance"}]
                segments=[]
                for start,line in [(10,label),(40,next_label)]:
                    words=[{"word":word,"start":start+index,"end":start+index+0.5} for index,word in enumerate(line.split())]
                    segments.append({"start":start,"end":start+len(words),"text":line,"words":words})
                aligned=align_listening_items(items,segments,"audio")
                self.assertEqual(10,aligned[0]["audioRange"]["startTime"])
                self.assertEqual(40,aligned[1]["audioRange"]["startTime"])

    def test_cli_failed_alignment_returns_nonzero(self):
        from align_study4_audio import main

        with (
            patch.object(sys, "argv", ["align_study4_audio.py", "--test-id", "missing"]),
            patch("align_study4_audio.process_test_alignment", side_effect=ValueError("missing test")),
        ):
            self.assertEqual(main(), 1)

    def test_asr_cache_is_bound_to_audio_and_model(self):
        import hashlib

        from align_study4_audio import transcribe_audio

        with tempfile.TemporaryDirectory() as directory:
            audio = Path(directory) / "audio.mp3"
            audio.write_bytes(b"audio")
            cache = Path(directory) / "asr.json"
            identity = {
                "audioSha256": hashlib.sha256(b"audio").hexdigest(),
                "model": "base.en",
                "language": "en",
                "wordTimestamps": True,
                "vadFilter": True,
                "beamSize": 5,
            }
            cache.write_text(
                json.dumps({"identity": identity, "segments": [{"text": "cached"}]}), encoding="utf-8"
            )
            self.assertEqual(transcribe_audio(audio, cache)[0]["text"], "cached")
            for stale in ("model", "audio"):
                kwargs = {"model_name": "small.en"} if stale == "model" else {}
                if stale == "audio":
                    audio.write_bytes(b"changed")
                with (
                    patch("faster_whisper.WhisperModel", side_effect=RuntimeError("cache rejected")),
                    self.assertRaisesRegex(RuntimeError, "cache rejected"),
                ):
                    transcribe_audio(audio, cache, device="cpu", **kwargs)

    def test_inline_part_boundary_uses_word_timestamps(self):
        items = [
            {"id": "q68", "questions": [68, 69, 70], "transcript": "source"},
            {"id": "q71", "questions": [71, 72, 73], "transcript": "source"},
        ]
        segments = [
            {"start": 1, "end": 2, "text": "Questions 68 through 70."},
            {
                "start": 3,
                "end": 10,
                "text": "Number 70. What next? Part 4.",
                "words": [
                    {"word": "Number", "start": 3, "end": 4},
                    {"word": "70.", "start": 4, "end": 5},
                    {"word": "What", "start": 5, "end": 6},
                    {"word": "next?", "start": 6, "end": 7},
                    {"word": "Part", "start": 8, "end": 9},
                    {"word": "4.", "start": 9, "end": 10},
                ],
            },
            {"start": 11, "end": 20, "text": "Directions."},
            {"start": 22, "end": 23, "text": "Questions 71 through 73."},
            {"start": 24, "end": 30, "text": "Speech."},
        ]
        result = align_listening_items(items, segments, "audio")
        self.assertEqual(result[0]["audioRange"]["endTime"], 7)
        del segments[1]["words"]
        self.assertIsNone(align_listening_items(items, segments, "audio")[0]["audioRange"]["endTime"])

    def test_missing_next_anchor_does_not_swallow_next_item(self):
        items = [{"id": str(n), "questions": [n], "transcript": "source"} for n in [1, 2, 3]]
        segments = [
            {"start": 1, "end": 2, "text": "Number 1."},
            {"start": 3, "end": 8, "text": "First and second speech."},
            {"start": 10, "end": 12, "text": "Number 3."},
            {"start": 13, "end": 18, "text": "Third speech."},
        ]
        result = align_listening_items(items, segments, "audio")
        self.assertIsNone(result[0]["audioRange"]["endTime"])
        self.assertIsNone(result[1]["audioRange"]["startTime"])

    def test_directions_are_not_question_seven_anchor(self):
        result = align_listening_items(
            [{"id": "q7", "questions": [7], "transcript": "source"}],
            [
                {"start": 1, "end": 2, "text": "Now let us begin with question number 7."},
                {"start": 5, "end": 6, "text": "Number 7."},
                {"start": 7, "end": 10, "text": "Speech."},
            ],
            "audio",
        )
        self.assertEqual(result[0]["audioRange"]["startTime"], 5)

    def test_spoken_number(self):
        self.assertEqual(spoken_number(1), "one")
        self.assertEqual(spoken_number(12), "twelve")
        self.assertEqual(spoken_number(20), "twenty")
        self.assertEqual(spoken_number(32), "thirty two")
        self.assertEqual(spoken_number(71), "seventy one")
        self.assertEqual(spoken_number(99), "ninety nine")

    def test_normalize_text(self):
        self.assertEqual(normalize_text("Number 1-2, look at Photo!"), "number 1 2 look at photo")

    def test_get_listening_items_from_dom(self):
        content = [
            {
                "tag": "div",
                "class": "question-item-wrapper",
                "children": [
                    {"tag": "div", "class": "question-number", "children": [{"text": "1"}]},
                    {
                        "tag": "div",
                        "class": "context-transcript",
                        "children": [
                            element(cls="collapse", children=[paragraph("(A) Waiting. (B) Walking.")])
                        ],
                    },
                ],
            },
            {
                "tag": "div",
                "class": "question-group-wrapper",
                "children": [
                    {"tag": "div", "class": "question-number", "children": [{"text": "32"}]},
                    {"tag": "div", "class": "question-number", "children": [{"text": "33"}]},
                    {"tag": "div", "class": "question-number", "children": [{"text": "34"}]},
                    {
                        "tag": "div",
                        "class": "context-transcript",
                        "children": [
                            element(
                                cls="collapse",
                                children=[paragraph("M: Hello, can I help you? W: Yes, please.")],
                            )
                        ],
                    },
                ],
            },
            # Reading item > 100 should be ignored
            {
                "tag": "div",
                "class": "question-item-wrapper",
                "children": [
                    {"tag": "div", "class": "question-number", "children": [{"text": "101"}]},
                ],
            },
        ]

        items = get_listening_items_from_dom(content)
        self.assertEqual(len(items), 2)
        self.assertEqual(items[0]["id"], "item_q01")
        self.assertEqual(items[0]["questions"], [1])
        self.assertEqual(items[0]["transcript"], "(A) Waiting. (B) Walking.")

        self.assertEqual(items[1]["id"], "item_q32_34")
        self.assertEqual(items[1]["questions"], [32, 33, 34])
        self.assertEqual(items[1]["transcript"], "M: Hello, can I help you? W: Yes, please.")

    def test_align_listening_items(self):
        items = [
            {"id": "item_q01", "questions": [1], "transcript": "(A) Waiting."},
            {"id": "item_q02", "questions": [2], "transcript": "(A) Walking."},
        ]
        segments = [
            {"start": 10.0, "end": 12.0, "text": "Part 1 directions."},
            {"start": 15.0, "end": 17.0, "text": "Number 1."},
            {"start": 18.0, "end": 25.0, "text": "Look at the picture. A, waiting."},
            {"start": 30.0, "end": 32.0, "text": "Number 2."},
            {"start": 33.0, "end": 40.0, "text": "Look at the picture. A, walking."},
            {"start": 45.0, "end": 47.0, "text": "Go on to the next page."},
        ]

        aligned = align_listening_items(items, segments, "exams/toeic/1001/audio/audio_full.mp3")
        self.assertEqual(len(aligned), 2)

        # Item 1
        self.assertEqual(aligned[0]["audioRange"]["startTime"], 15.0)
        self.assertEqual(aligned[0]["audioRange"]["endTime"], 25.0)

        # Item 2
        self.assertEqual(aligned[1]["audioRange"]["startTime"], 30.0)
        self.assertEqual(aligned[1]["audioRange"]["endTime"], 40.0)

    def test_process_test_alignment_flow(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            test_dir = Path(temp_dir) / "study4" / "tests" / "9999"
            run_dir = test_dir / "runs" / "run_01"
            run_dir.mkdir(parents=True, exist_ok=True)

            cache_dir = test_dir / "cache"
            cache_dir.mkdir(parents=True, exist_ok=True)
            local_audio = cache_dir / "audio_full.mp3"
            local_audio.write_bytes(b"mock_audio_bytes")

            export_data = full_export()
            with (run_dir / "export.json").open("w", encoding="utf-8") as f:
                json.dump(export_data, f)

            mock_segments = listening_segments(export_data)

            with (
                patch("align_study4_audio.transcribe_audio", return_value=mock_segments),
                patch("mutagen.mp3.MP3", return_value=SimpleNamespace(info=SimpleNamespace(length=2000))),
            ):
                summary = process_test_alignment("9999", crawl_root=temp_dir)

            self.assertEqual(summary["testId"], "9999")
            self.assertEqual(summary["itemCount"], 54)
            self.assertEqual(summary["alignedCount"], 54)

            # Check that audio-alignment.json was saved
            align_file = run_dir / "audio-alignment.json"
            self.assertTrue(align_file.exists())
            align_data = json.loads(align_file.read_text(encoding="utf-8"))
            self.assertEqual(align_data["items"][0]["audioRange"]["startTime"], 5.0)

            # Check that export.json was updated
            updated_export = json.loads((run_dir / "export.json").read_text(encoding="utf-8"))
            self.assertIn("audio_ranges", updated_export)
            self.assertEqual(updated_export["audio_ranges"]["item_q01"]["startTime"], 5.0)


if __name__ == "__main__":
    unittest.main()
