import copy
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import Mock, patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import toeic_storage as storage
import publish_toeic_media as publisher


class StorageTests(unittest.TestCase):
    def test_question_names_are_stable_across_existing_storage_paths(self):
        for prefix in ("crops/", "toeic/ets/test/images/"):
            with self.subTest(prefix=prefix):
                first, second = prefix + "old-a.png", prefix + "old-b.png"
                content = {"audioAssetId":"old.mp3", "assets":[{"sourceId":"old.mp3"},{"sourceId":first},{"sourceId":second}], "parts":[{"partNumber":3,"items":[{"questions":[{"questionNumber":62},{"questionNumber":63},{"questionNumber":64}],"media":[{"assetId":first},{"assetId":second}]}]}]}
                labels = storage.media_labels(content)
                self.assertEqual("audio", labels["old.mp3"])
                self.assertEqual("part-3-q62-64-image-01", labels[first])
                self.assertEqual("part-3-q62-64-image-02", labels[second])
                self.assertEqual("ets/test/images/test-part-3-q62-64-image-01.png", storage.media_key("ets","test",{"sha256":"a"*64},"image/png",label=labels[first]))

    def test_local_package_checks_hashes_and_matches_bucket_tree(self):
        import hashlib
        for corrupt in (False, True):
            with self.subTest(corrupt=corrupt), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                source = root / "source.mp3"; source.write_bytes(b"audio source")
                asset = {"sourceId":"source.mp3","sha256":hashlib.sha256(b"audio source").hexdigest()}
                key = storage.media_key("ets","test-1",asset,"audio/mpeg",label="audio")
                if corrupt: source.write_bytes(b"changed")
                content = {"audioAssetId":"toeic/"+key}
                if corrupt:
                    with self.assertRaises(ValueError): storage.export_package(content,[(asset,source,key,"audio/mpeg")],root/"output","ets","test-1")
                else:
                    target = storage.export_package(content,[(asset,source,key,"audio/mpeg")],root/"output","ets","test-1")
                    self.assertEqual(root/"output/toeic/ets/test-1",target)
                    self.assertEqual(content,json.loads((target/"content.json").read_text(encoding="utf-8")))
                    self.assertEqual(source.read_bytes(),(target/"audio/test-1-audio.mp3").read_bytes())

    def test_policy_normalization_and_private_json(self):
        for broad in (False, True):
            with self.subTest(broad=broad):
                resources = ["arn:aws:s3:::toeic/*/*/audio/*", "arn:aws:s3:::toeic/*/*/images/*"]
                if broad:
                    resources.append("arn:aws:s3:::toeic/*")
                policy = {"Version":"2012-10-17", "Statement":[{"Sid":"ToeicMediaRead", "Effect":"Allow", "Principal":{"AWS":["*"]}, "Action":["s3:GetObject"], "Resource":resources}]}
                client = Mock()
                client.get_bucket_policy.return_value = json.dumps(policy)
                if broad:
                    with self.assertRaises(ValueError):
                        storage.ensure_bucket(client)
                else:
                    storage.ensure_bucket(client)
                client.set_bucket_policy.assert_not_called()

    def test_layout_rejects_bad_slugs_and_separates_tests(self):
        digest = "a" * 64
        for category, test, kind, valid in [
            ("ets-2024", "test-1", "image/png", True),
            ("ets-2024", "test-2", "audio/mpeg", True),
            ("../ets", "test-1", "image/png", False),
            ("ets", "test/1", "image/png", False),
            ("ets", "test-1", "application/json", False),
            ("", "test-1", "image/png", False),
        ]:
            with self.subTest(category=category, test=test, kind=kind):
                if valid:
                    key = storage.media_key(category, test, {"sha256": digest}, kind, label="audio" if kind.startswith("audio") else "part-1-q01")
                    self.assertTrue(key.startswith(f"{category}/{test}/"))
                    self.assertIn("/audio/" if kind.startswith("audio") else "/images/", key)
                else:
                    with self.assertRaises(ValueError):
                        storage.media_key(category, test, {"sha256": digest}, kind, label="audio" if kind.startswith("audio") else "part-1-q01")

    def test_relocation_updates_all_references_without_rewriting_text(self):
        for source in ("Audio/Test_02.mp3", "https://example.test/audio.mp3", "exams/toeic/" + "a" * 64):
            with self.subTest(source=source):
                original = {"audioAssetId": source, "assets": [{"sourceId": source, "sha256": "a" * 64}], "parts": [{"items": [{"itemAudio": {"assetId": source}, "passages": [{"blocks": [{"assetId": source, "text": source}]}]}]}]}
                before = copy.deepcopy(original)
                target = "toeic/ets-2024/test-2/audio/" + "a" * 64 + ".mp3"
                stored = storage.relocate_content(original, {source: target})
                self.assertEqual(before, original)
                self.assertEqual(target, stored["audioAssetId"])
                self.assertEqual(target, stored["assets"][0]["sourceId"])
                self.assertEqual(target, stored["parts"][0]["items"][0]["itemAudio"]["assetId"])
                block = stored["parts"][0]["items"][0]["passages"][0]["blocks"][0]
                self.assertEqual(target, block["assetId"])
                self.assertEqual(source, block["text"])

    def test_publisher_dry_run_and_verified_upload_use_same_layout(self):
        for dry_run in (True, False):
            with self.subTest(dry_run=dry_run), tempfile.TemporaryDirectory() as directory:
                path = Path(directory) / "audio.mp3"
                path.write_bytes(b"audio")
                asset = {"sourceId": "audio.mp3", "sha256": "a" * 64}
                content = {"audioAssetId": "audio.mp3", "assets": [asset], "parts": []}
                key = storage.media_key("ets-2024", "test-2", asset, "audio/mpeg", label="audio")
                client = Mock()
                with patch.object(publisher, "prepare", return_value=[(asset, path, key, "audio/mpeg")]), patch.object(publisher, "minio_client", return_value=client) as create, patch.object(publisher, "ensure_bucket"), patch.object(publisher, "assert_json_available"), patch.object(publisher, "verify_object") as verify, patch.object(publisher, "publish_json", return_value="ets-2024/test-2/content.json") as publish_json:
                    stored, summary = publisher.publish_content(content, dry_run=dry_run, category_slug="ets-2024", test_slug="test-2")
                    self.assertEqual("toeic/" + key, stored["audioAssetId"])
                    self.assertEqual(1, summary["validatedAssets"])
                    if dry_run:
                        create.assert_not_called()
                        publish_json.assert_not_called()
                    else:
                        client.fput_object.assert_called_once_with("toeic", key, str(path), content_type="audio/mpeg")
                        verify.assert_called_once_with(client, "toeic", key, asset["sha256"])
                        publish_json.assert_called_once_with(client, "ets-2024", "test-2", stored, expected_content=content)

    def test_upload_failure_never_publishes_exam_json(self):
        for failure in ("upload", "verification"):
            with self.subTest(failure=failure), tempfile.TemporaryDirectory() as directory:
                path = Path(directory) / "audio.mp3"; path.write_bytes(b"audio")
                asset = {"sourceId": "audio.mp3", "sha256": "a" * 64}
                client = Mock()
                if failure == "upload":
                    client.fput_object.side_effect = OSError("upload failed")
                with patch.object(publisher, "prepare", return_value=[(asset, path, "ets/test/audio/a.mp3", "audio/mpeg")]), patch.object(publisher, "minio_client", return_value=client), patch.object(publisher, "ensure_bucket"), patch.object(publisher, "assert_json_available"), patch.object(publisher, "verify_object", side_effect=ValueError("hash mismatch") if failure == "verification" else None), patch.object(publisher, "publish_json") as publish_json:
                    with self.assertRaises((OSError, ValueError)):
                        publisher.publish_content({"audioAssetId": "audio.mp3", "assets": [asset], "parts": []}, category_slug="ets", test_slug="test")
                    publish_json.assert_not_called()
