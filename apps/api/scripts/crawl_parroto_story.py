"""Export Vietnamese Parroto stories to Engflex story packages.

Run from apps/api:
    uv run --with playwright --with python-dotenv python scripts/crawl_parroto_story.py OUTPUT_DIR
    uv run --with playwright --with minio --with python-dotenv python scripts/crawl_parroto_story.py OUTPUT_DIR --all --publish

Uses installed Chrome. Add --publish (and --with minio) to upload the package.
"""

import argparse
import json
import mimetypes
import re
import shutil
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from urllib.parse import urlsplit
from urllib.request import Request, urlopen

from playwright.sync_api import sync_playwright

from publish_story import build_package, publish_package


CATALOGUE_URL = "https://parroto.app/vi/stories"
SLUG_PATTERN = r"[a-z0-9]+(?:-[a-z0-9]+)*"


def catalogue_slugs(page):
    page.goto(CATALOGUE_URL, wait_until="domcontentloaded")
    initial = page.locator("#__NEXT_DATA__").evaluate(
        "(node) => JSON.parse(node.textContent).props.pageProps.initial"
    )
    cards = page.locator('a[href^="/vi/stories/"]')
    slugs = set()
    for number in range(1, initial["totalPages"] + 1):
        page.get_by_text(f"Trang {number} / {initial['totalPages']}", exact=True).wait_for()
        current = cards.evaluate_all("(links) => links.map(link => link.getAttribute('href'))")
        slugs.update(
            slug for href in current
            if re.fullmatch(SLUG_PATTERN, slug := href.rsplit("/", 1)[-1])
        )
        if number < initial["totalPages"]:
            page.get_by_role("button", name="Sau", exact=True).click()
    if len(slugs) != initial["total"]:
        raise ValueError(f"catalogue expected {initial['total']} stories, found {len(slugs)}")
    return sorted(slugs)


class UnavailableStory(Exception):
    pass


def read_story(page, url):
    page.goto(url, wait_until="domcontentloaded")
    sentences = page.locator("[data-sentence-index]")
    pro_gate = page.get_by_role("button", name="Mở khóa Pro", exact=True)
    sentences.first.or_(pro_gate.first).first.wait_for(timeout=20_000)
    if not sentences.count():
        raise UnavailableStory("the full reader requires Pro")
    if not page.locator("audio").count():
        raise UnavailableStory("the reader has no narration")
    page.wait_for_function("document.querySelector('audio')?.readyState >= 1")
    return page.evaluate(
            r"""() => {
                const meta = JSON.parse(document.querySelector('#__NEXT_DATA__').textContent)
                    .props.pageProps.meta;
                const audio = document.querySelector('audio');
                const scenes = [...document.querySelectorAll('section[data-scene-index]')];
                const sceneImages = Object.fromEntries(
                    [...document.images].flatMap(img => {
                        const match = new URL(img.src, location.href).pathname.match(/\/scene-(\d+)-/);
                        return match ? [[match[1], img.src]] : [];
                    })
                );
                const sentences = scenes.flatMap(scene =>
                    [...scene.querySelectorAll('[data-sentence-index]')].map(sentence => {
                        const translation = sentence.querySelector(':scope > span.block');
                        const english = [...sentence.querySelectorAll(':scope > span:not(.block)')]
                            .map(word => word.textContent).join('').trim();
                        sentence.click();
                        const start = Math.max(0, Math.round(audio.currentTime * 1000) - 1) / 1000;
                        audio.pause();
                        return {
                            index: Number(sentence.dataset.sentenceIndex),
                            englishText: english,
                            vietnameseText: translation?.textContent.trim() || '',
                            imageUrl: sceneImages[scene.dataset.sceneIndex],
                            startTimestamp: start
                        };
                    })
                );
                const heading = document.querySelector('h1');
                return {
                    slug: meta.slug, title: meta.title, cefrLevel: meta.level, topic: 'fairy-tale',
                    titleVietnamese: heading?.parentElement?.nextElementSibling?.textContent.trim(),
                    thumbnailUrl: meta.cover, audioUrl: audio.currentSrc, duration: audio.duration,
                    sentences
                };
            }"""
    )


def download_asset(url, name, directory, kind, referer):
    if urlsplit(url).scheme != "https":
        raise ValueError(f"expected an HTTPS source for {name}")
    request = Request(url, headers={"Referer": referer, "User-Agent": "Mozilla/5.0"})
    with urlopen(request, timeout=45) as response:
        content_type = response.headers.get_content_type()
        if not content_type.startswith(f"{kind}/"):
            raise ValueError(f"unexpected content type for {name}: {content_type}")
        extension = Path(urlsplit(url).path).suffix.lower()
        if mimetypes.guess_type(f"asset{extension}")[0] != content_type:
            extension = mimetypes.guess_extension(content_type)
        if not extension or not re.fullmatch(r"\.[a-z0-9]+", extension):
            raise ValueError(f"unsupported media format: {content_type}")
        filename = f"{name}{extension}"
        temporary = directory / f"{filename}.part"
        with temporary.open("wb") as output:
            shutil.copyfileobj(response, output)
        temporary.replace(directory / filename)
    return filename


def save_story(story, slug, directory):
    url = f"{CATALOGUE_URL}/{slug}"
    if story["slug"] != slug:
        raise ValueError("reader returned a different story")
    sentences = story["sentences"]
    for index, sentence in enumerate(sentences):
        if sentence.pop("index") != index:
            raise ValueError(f"incomplete reader data for sentence {index + 1}")
        sentence["sequence"] = index + 1
        sentence["endTimestamp"] = (
            sentences[index + 1]["startTimestamp"] if index + 1 < len(sentences) else story["duration"]
        )

    directory.mkdir(parents=True, exist_ok=True)
    story["audioUrl"] = download_asset(story["audioUrl"], "narration", directory, "audio", url)
    if story.get("thumbnailUrl"):
        story["thumbnailUrl"] = download_asset(story["thumbnailUrl"], "cover", directory, "image", url)
    scenes = {}
    for sentence in sentences:
        image = sentence.get("imageUrl")
        if image:
            if image not in scenes:
                scenes[image] = download_asset(image, f"scene-{len(scenes) + 1:02d}", directory, "image", url)
            sentence["imageUrl"] = scenes[image]

    manifest = directory / "manifest.json"
    manifest.write_text(json.dumps(story, ensure_ascii=False, indent=2), encoding="utf-8")
    return manifest


def process_story(story, slug, directory, publish):
    manifest = directory / "manifest.json"
    if story is not None:
        manifest = save_story(story, slug, directory)
    package = build_package(manifest)
    if publish:
        publish_package(*package)
    return {"status": "published" if publish else "exported"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output_directory", type=Path)
    selection = parser.add_mutually_exclusive_group()
    selection.add_argument("--slug", default="cinderella")
    selection.add_argument("--all", action="store_true")
    parser.add_argument("--publish", action="store_true")
    args = parser.parse_args()
    if not re.fullmatch(SLUG_PATTERN, args.slug):
        parser.error("invalid story slug")
    report = {}

    def record(slug, result):
        report[slug] = result
        print(f"{slug}: {result}", flush=True)
        if args.all:
            (args.output_directory / "report.json").write_text(
                json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8"
            )

    def record_completed(future, slug):
        try:
            record(slug, future.result())
        except Exception as exc:
            if not args.all:
                raise
            record(slug, {"status": "failed", "reason": str(exc)})

    with sync_playwright() as playwright, ThreadPoolExecutor(max_workers=4) as workers:
        browser = playwright.chromium.launch(channel="chrome")
        page = browser.new_page()
        slugs = catalogue_slugs(page) if args.all else [args.slug]
        args.output_directory.mkdir(parents=True, exist_ok=True)
        pending = {}
        for slug in slugs:
            directory = args.output_directory / slug if args.all else args.output_directory
            try:
                story = None
                if not args.all or not (directory / "manifest.json").exists():
                    story = read_story(page, f"{CATALOGUE_URL}/{slug}")
                future = workers.submit(process_story, story, slug, directory, args.publish)
                pending[future] = slug
            except UnavailableStory as exc:
                record(slug, {"status": "unavailable", "reason": str(exc)})
            except Exception as exc:
                if not args.all:
                    raise
                record(slug, {"status": "failed", "reason": str(exc)})
            for future in list(pending):
                if future.done():
                    record_completed(future, pending.pop(future))
        for future in as_completed(pending):
            record_completed(future, pending[future])
    return int(any(item["status"] == "failed" for item in report.values()))


if __name__ == "__main__":
    raise SystemExit(main())
