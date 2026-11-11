"""Allowlisted server import document. Never use directly as an exam-taking API DTO."""

import argparse
import json
import sys
import tempfile
from pathlib import Path
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

from pipeline_io import ROOT, digest, read, test_dir, write
from schemas import AudioRange, Inline, Test


class ContentModel(BaseModel):
    # Intentionally project only declared fields; new extraction metadata cannot leak.
    model_config = ConfigDict(extra="ignore", strict=True, allow_inf_nan=False)


class ContentOption(ContentModel):
    key: Literal["A", "B", "C", "D"]
    text: str | None


class ContentVocabulary(ContentModel):
    word: str | None = None
    pos: str | None = None
    meaning: str | None = None
    raw: str | None = None


class ContentBlock(ContentModel):
    id: str
    type: Literal["sourceRegion", "paragraph", "heading", "table", "image", "emailHeader", "chatMessage"]
    text: str | None = None
    rows: list[list[str]] | None = None
    inline: list[Inline] = Field(default_factory=list)
    assetId: str | None = None
    vietnameseText: str | None = None


class ContentPassage(ContentModel):
    id: str
    format: str | None = None
    blocks: list[ContentBlock]


class ContentTranscript(ContentModel):
    id: str
    role: Literal["prompt", "option", "utterance"]
    speaker: str | None = None
    optionKey: str | None = None
    text: str
    vietnameseText: str | None = None


class ContentQuestion(ContentModel):
    id: str
    questionNumber: int
    stem: str | None
    options: list[ContentOption]
    correctKey: str | None
    explanation: str | None
    vietnameseText: str | None
    vocabulary: list[ContentVocabulary]
    targetIds: list[str]


class ContentItem(ContentModel):
    id: str
    type: Literal["single", "group"]
    questions: list[ContentQuestion]
    passages: list[ContentPassage]
    transcript: list[ContentTranscript]
    itemAudio: AudioRange | None
    stimulusAudio: AudioRange | None
    media: list[ContentBlock]


class ContentPart(ContentModel):
    partNumber: int
    section: Literal["listening", "reading"]
    items: list[ContentItem]


class ContentAsset(ContentModel):
    sourceId: str
    sha256: str


class ContentTest(ContentModel):
    schemaVersion: Literal["0.1.0"]
    collectionId: str
    testId: str
    testNumber: int
    title: str
    audioAssetId: str
    assets: list[ContentAsset]
    parts: list[ContentPart]


def project(payload):
    """Keep graph identity/meaning intact and retain only referenced media assets."""
    result = ContentTest.model_validate(payload).model_dump(exclude_unset=True)
    references = {result["audioAssetId"]}

    def collect(value):
        if isinstance(value, dict):
            if value.get("assetId"):
                references.add(value["assetId"])
            for child in value.values():
                collect(child)
        elif isinstance(value, list):
            for child in value:
                collect(child)

    collect(result["parts"])
    assets = {}
    for asset in result["assets"]:
        if asset["sourceId"] not in references:
            continue
        if asset["sourceId"] in assets:
            raise ValueError("Duplicate referenced asset identity")
        assets[asset["sourceId"]] = asset
    if references != set(assets):
        raise ValueError(f"Unresolved asset references: {sorted(references - set(assets))}")
    result["assets"] = [assets[key] for key in sorted(assets)]
    return result


def export_content(test: Test, output, *, release=False):
    payload = test.model_dump()
    content = project(payload)
    write(output / "content.json", content)
    # Compact copy is for sizing/transport; semantic content is identical.
    (output / "content.compact.json").write_text(
        json.dumps(content, ensure_ascii=False, separators=(",", ":"), allow_nan=False) + "\n",
        encoding="utf-8",
    )
    metadata = {
        "status": "release" if release else "preview-needs-review",
        "sourceManifestHash": test.sourceManifestHash,
        "draftHash": digest(payload),
        "contentHash": digest(content),
        "exportSchemaVersion": "0.1.0",
        "evidenceLocation": "../raw/",
        "note": "Server import document contains answers; not a learner-facing payload.",
    }
    write(output / "provenance.json", metadata)
    return content


def package_content(content, output_root, category_slug, test_slug, source_root, crop_root):
    # Share the publisher's naming, hash checks and bucket/key layout.
    scripts = str(ROOT / "apps/api/scripts")
    if scripts not in sys.path:
        sys.path.insert(0, scripts)
    from publish_toeic_media import prepare
    from toeic_storage import BUCKET, export_package, relocate_content

    with tempfile.TemporaryDirectory() as directory:
        prepared = prepare(content, source_root, crop_root, Path(directory), category_slug, test_slug)
        stored = relocate_content(
            content, {asset["sourceId"]: f"{BUCKET}/{key}" for asset, _, key, _ in prepared}
        )
        return export_package(stored, prepared, output_root, category_slug, test_slug)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--test", type=int, choices=range(1, 11), default=2)
    args = parser.parse_args()
    root = test_dir(args.test)
    test = Test.model_validate(read(root / "draft.json"))
    export_content(test, root / "preview")
    print(f"Preview only (not released): {root / 'preview/content.json'}")


if __name__ == "__main__":
    main()
