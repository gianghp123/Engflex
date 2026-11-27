import importlib.util
import hashlib
import json
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "scripts"))
sys.path.insert(0, str(ROOT / "apps/api/scripts"))

from study4_ielts_content import answer_variants, context_block, group_type

spec = importlib.util.spec_from_file_location("publish_ielts", ROOT / "apps/api/scripts/publish_ielts.py")
publisher = importlib.util.module_from_spec(spec)
spec.loader.exec_module(publisher)


class IeltsContentTests(unittest.TestCase):
    def test_explicit_answer_alternatives(self):
        for source, expected in {
            "monkey(s)": ["monkey", "monkeys"],
            "(the) royal antelope": ["royal antelope", "the royal antelope"],
            "66/sixty-six Lake": ["66 Lake", "sixty-six Lake"],
            "splayed hooves [OR] long, splayed hooves": ["splayed hooves", "long, splayed hooves"],
        }.items():
            self.assertEqual(answer_variants(source), expected)
        with self.assertRaises(ValueError):
            answer_variants("one/two three/four")

    def test_context_preserves_merged_cells_and_image_identity(self):
        context = {"tag": "table", "children": [{"tag": "tr", "children": [
            {"tag": "td", "colspan": "2", "rowspan": "3", "children": [
                {"text": "Question (1)______"}, {"tag": "img", "asset": "image-1", "alt": "Diagram"},
            ]},
        ]}]}
        block = context_block(context, {"image-1": "ielts/academic/test/images/diagram.png"})
        cell = block["children"][0]["children"][0]
        self.assertEqual((cell["colSpan"], cell["rowSpan"]), (2, 3))
        self.assertEqual(cell["children"][1]["assetId"], "ielts/academic/test/images/diagram.png")
        with self.assertRaises(ValueError):
            context_block({"tag": "script", "children": [{"text": "untrusted()"}]}, {})
        with self.assertRaises(KeyError):
            context_block({"tag": "img", "asset": "missing"}, {})

    def test_unknown_question_type_fails_instead_of_guessing(self):
        with self.assertRaises(ValueError):
            group_type("An unsupported exercise")

    def test_publish_does_not_overwrite_different_content(self):
        client = MagicMock()
        client.bucket_exists.return_value = True
        response = MagicMock()
        response.read.return_value = json.dumps({"testId": "existing"}).encode()
        client.get_object.return_value = response
        with self.assertRaisesRegex(ValueError, "refusing to overwrite"):
            publisher.publish(client, {"testId": "different"}, [], "academic", "test")
        client.put_object.assert_not_called()
        client.set_bucket_policy.assert_not_called()
        response.release_conn.assert_called_once()

    def test_media_upload_precedes_private_json(self):
        client = MagicMock()
        client.bucket_exists.return_value = False
        from minio.error import S3Error
        client.get_bucket_policy.side_effect = S3Error(response=None, code="NoSuchBucketPolicy", message="missing", resource="", request_id="", host_id="")
        content = {"testId": "test"}
        raw = publisher.encode(content)
        image = b"verified test bytes"
        responses = []
        for value in (image, raw):
            response = MagicMock()
            response.stream.return_value = [value]
            responses.append(response)
        client.get_object.side_effect = responses
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "diagram.png"
            source.write_bytes(image)
            media = [{"path": source, "key": "academic/test/images/diagram.png", "sha256": hashlib.sha256(image).hexdigest(), "contentType": "image/png"}]
            self.assertEqual(publisher.publish(client, content, media, "academic", "test"), "academic/test/content.json")
        self.assertEqual([call.args[1] for call in client.put_object.call_args_list], ["academic/test/images/diagram.png", "academic/test/content.json"])
        policy = json.loads(client.set_bucket_policy.call_args.args[1])
        resources = policy["Statement"][0]["Resource"]
        self.assertTrue(all("/audio/" in r or "/images/" in r for r in resources))
        self.assertFalse(any("content.json" in r for r in resources))
