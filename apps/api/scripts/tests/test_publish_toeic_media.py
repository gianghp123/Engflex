import copy
import hashlib
import importlib.util
import io
from unittest.mock import patch
from pathlib import Path
import tempfile
import unittest
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))


spec = importlib.util.spec_from_file_location("publish_toeic_media", Path(__file__).resolve().parents[1] / "publish_toeic_media.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class PublishTests(unittest.TestCase):
    def test_all_media_selection_and_missing_references(self):
        for fault in (None, "missing", "duplicate"):
            with self.subTest(fault=fault):
                content = {
                    "audioAssetId": "audio.mp3",
                    "assets": [{"sourceId": "audio.mp3"}, {"sourceId": "photo.png"}, {"sourceId": "reading.png"}, {"sourceId": "unused.png"}],
                    "parts": [
                        {"partNumber": 1, "items": [{"itemAudio": {"assetId": "audio.mp3"}, "media": [{"assetId": "photo.png"}]}]},
                        {"partNumber": 4, "items": [{"itemAudio": {"assetId": "audio.mp3"}}]},
                        {"partNumber": 7, "items": [{"passages": [{"blocks": [{"assetId": "reading.png"}]}]}]},
                    ],
                }
                if fault == "missing":
                    content["assets"] = content["assets"][1:]
                elif fault == "duplicate":
                    content["assets"].append(content["assets"][0])
                before = copy.deepcopy(content)
                if fault:
                    with self.assertRaises(ValueError):
                        module.media_assets(content)
                else:
                    self.assertEqual(["audio.mp3", "photo.png", "reading.png", "unused.png"], [asset["sourceId"] for asset in module.media_assets(content)])
                self.assertEqual(before, content)

    def test_remote_media_is_verified_before_upload(self):
        for fault in (None, "hash", "redirect", "scheme", "digest"):
            with self.subTest(fault=fault), tempfile.TemporaryDirectory() as directory:
                blob = b"verified remote media"
                digest = hashlib.sha256(blob).hexdigest()
                asset = {"sourceId": "https://media.example.test/file.png", "sha256": digest}
                if fault == "scheme":
                    asset["sourceId"] = "file:///private.png"
                elif fault == "digest":
                    asset["sha256"] = "invalid"
                class Response(io.BytesIO):
                    def geturl(self):
                        return "file:///redirect" if fault == "redirect" else asset["sourceId"]
                with patch.object(module, "urlopen", return_value=Response(b"changed" if fault == "hash" else blob)):
                    if fault:
                        with self.assertRaises(ValueError):
                            module.remote_asset(asset, Path(directory))
                    else:
                        path, key = module.remote_asset(asset, Path(directory))
                        self.assertEqual(blob, path.read_bytes())
                        self.assertEqual(f"exams/toeic/{digest}", key)

    def test_local_asset_hash_and_root_boundary(self):
        for fault in (None, "hash", "traversal", "absolute", "missing", "invalid_digest"):
            with self.subTest(fault=fault), tempfile.TemporaryDirectory() as directory:
                root = Path(directory) / "source"
                root.mkdir()
                (root / "audio.mp3").write_bytes(b"original media")
                (Path(directory) / "outside.mp3").write_bytes(b"original media")
                digest = hashlib.sha256(b"original media").hexdigest()
                asset = {"sourceId": "audio.mp3", "sha256": digest}
                if fault == "hash":
                    (root / "audio.mp3").write_bytes(b"changed media")
                elif fault == "traversal":
                    asset["sourceId"] = "../outside.mp3"
                elif fault == "absolute":
                    asset["sourceId"] = str((root / "audio.mp3").resolve())
                elif fault == "missing":
                    asset["sourceId"] = "missing.mp3"
                elif fault == "invalid_digest":
                    asset["sha256"] = "invalid"
                if fault:
                    with self.assertRaises(ValueError):
                        module.local_asset(asset, root, root)
                else:
                    path, key = module.local_asset(asset, root, root)
                    self.assertEqual(root.resolve() / "audio.mp3", path)
                    self.assertEqual(f"exams/toeic/{digest}", key)


if __name__ == "__main__":
    unittest.main()
