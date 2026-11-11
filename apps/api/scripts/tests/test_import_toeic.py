import copy
import hashlib
import importlib.util
import io
import json
import sys
import tempfile
from types import SimpleNamespace
from unittest.mock import Mock
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("import_toeic", Path(__file__).resolve().parents[1] / "import_toeic.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
FIXTURE = Path(__file__).resolve().parents[2] / "internal/modules/exams/toeic/tests/testdata/synthetic.json"


class Response(io.BytesIO):
    def __init__(self, data, url):
        super().__init__(data)
        self.url = url

    def geturl(self):
        return self.url


class ImportTests(unittest.TestCase):
    def test_local_package_preserves_review_and_exact_exam_content(self):
        sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
        from toeic_storage import relocate_content
        for fault in (None, "answer", "missing", "preview", "stale_review"):
            with self.subTest(fault=fault):
                content = json.loads(FIXTURE.read_text(encoding="utf-8"))
                packaged = relocate_content(content, {asset["sourceId"]: "toeic/ets/test/images/test-"+str(i)+".png" for i,asset in enumerate(content["assets"])})
                provenance = {"status":"release","exportSchemaVersion":"0.1.0","contentHash":module.digest(content),"draftHash":"draft"}
                report = {"payloadHash":"draft","releaseReady":True,"issues":[],"issueCounts":{}}
                if fault == "answer": packaged["parts"][0]["items"][0]["questions"][0]["correctKey"] = "D"
                if fault == "missing": packaged["assets"] = []
                if fault == "preview": provenance["status"] = "preview-needs-review"
                if fault == "stale_review": report["payloadHash"] = "old"
                with patch.object(module,"validate_content"):
                    if fault:
                        with self.assertRaises(ValueError): module.prepare_package(content,provenance,report,packaged)
                    else:
                        self.assertEqual(packaged,module.prepare_package(content,provenance,report,packaged))

    def test_release_and_media_checks(self):
        for fault in (None, "preview", "changed", "unreviewed", "stale_review", "missing", "hash", "short_audio", "redirect"):
            with self.subTest(fault=fault):
                content = json.loads(FIXTURE.read_text(encoding="utf-8"))
                blobs = {"audio.wav": b"audio", "photo.png": b"image"}
                for asset in content["assets"]:
                    asset["sha256"] = hashlib.sha256(blobs[asset["sourceId"]]).hexdigest()
                provenance = {"status": "release", "exportSchemaVersion": "0.1.0", "contentHash": module.digest(content), "draftHash": "reviewed-draft"}
                report = {"payloadHash": "reviewed-draft", "releaseReady": True, "issues": [{"status": "verified"}], "issueCounts": {}}
                urls = {name: f"https://example.test/{name}" for name in blobs}
                if fault == "preview": provenance["status"] = "preview-needs-review"
                if fault == "changed": content["title"] = "changed"
                if fault == "unreviewed": report["issues"][0]["status"] = "open"
                if fault == "stale_review": report["payloadHash"] = "old-draft"
                if fault == "missing": del urls["photo.png"]
                if fault == "hash": blobs["photo.png"] = b"changed"
                def open_url(url, timeout):
                    name = url.rsplit("/", 1)[-1]
                    return Response(blobs[name], "file:///bad" if fault == "redirect" else url)
                probe = type("Probe", (), {"stdout": json.dumps({"streams": [{"codec_type": "audio"}, {"codec_type": "video"}], "format": {"duration": "1" if fault == "short_audio" else "1000"}})})()
                with patch.object(module, "validate_content"), patch.object(module, "urlopen", side_effect=open_url), patch.object(module.subprocess, "run", return_value=probe):
                    if fault:
                        with self.assertRaises(ValueError): module.prepare(content, provenance, report, urls)
                    else:
                        before = copy.deepcopy(content)
                        result = module.prepare(content, provenance, report, urls)
                        self.assertEqual(before, content)
                        self.assertEqual(urls["audio.wav"], result["parts"][0]["items"][0]["itemAudio"]["assetId"])
                        self.assertEqual(urls["photo.png"], result["parts"][0]["items"][0]["media"][0]["assetId"])

    def test_import_aborts_before_database_write_when_media_not_ready(self):
        for fault in ("release", "publish"):
            with self.subTest(fault=fault), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                release = root / "release"
                release.mkdir()
                for path in (release / "content.json", release / "provenance.json", root / "validation.json", root / "media.json"):
                    path.write_text("{}", encoding="utf-8")
                publisher = Mock(side_effect=ValueError("media not ready"))
                connect = Mock(side_effect=AssertionError("DB must not be opened"))
                argv = ["import_toeic.py", "--release", str(release), "--media", str(root / "media.json"), "--category-slug", "test", "--category-title", "Test", "--test-slug", "test"]
                modules = {"publish_toeic_media": SimpleNamespace(publish_content=publisher), "psycopg": SimpleNamespace(connect=connect)}
                with patch.object(sys, "argv", argv), patch.dict(sys.modules, modules), patch.object(module, "prepare", side_effect=ValueError("release not ready") if fault == "release" else None, return_value={"prepared": True}):
                    with self.assertRaises(ValueError):
                        module.main()
                connect.assert_not_called()
                if fault == "release":
                    publisher.assert_not_called()
                else:
                    publisher.assert_called_once()


if __name__ == "__main__":
    unittest.main()
