import hashlib
import sys
from unittest.mock import patch

import pytest

from content_export import package_content
from pipeline_io import ROOT

sys.path.insert(0, str(ROOT / "apps/api/scripts"))
import publish_toeic_media as publisher
from toeic_storage import media_key


@pytest.mark.parametrize(
    "category,test",
    [("ets-2024", "ets-2024-test-02"), ("toeic-lr-collection-1", "toeic-lr-collection-1-test-1")],
)
def test_pipeline_and_publisher_export_identical_named_tree(tmp_path, category, test):
    source = tmp_path / "audio.mp3"
    source.write_bytes(b"verified bytes")
    asset = {"sourceId": "audio.mp3", "sha256": hashlib.sha256(source.read_bytes()).hexdigest()}
    content = {"audioAssetId": "audio.mp3", "assets": [asset], "parts": []}
    key = media_key(category, test, asset, "audio/mpeg", label="audio")
    with patch.object(publisher, "prepare", return_value=[(asset, source, key, "audio/mpeg")]):
        target = package_content(content, tmp_path / "output", category, test, tmp_path, tmp_path)
    assert target == tmp_path / "output/toeic" / category / test
    assert (target / "audio" / f"{test}-audio.mp3").read_bytes() == source.read_bytes()
    assert (target / "content.json").is_file()
    assert content["audioAssetId"] == "audio.mp3"
