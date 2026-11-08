# /// script
# requires-python = ">=3.11"
# dependencies = ["mcp>=1.20,<2", "python-dotenv>=1,<2", "httpx>=0.28,<1", "pillow>=11,<13", "mutagen>=1.47,<2", "minio>=7.2,<8"]
# ///
"""Export a Study4 page through Playwright MCP. See docs/crawl/mcp-export.md.

Uses an independently configured MCP server, not Codex's private UI session.
Only export content you are permitted to copy. No login or submission automation.
"""

import argparse
import asyncio
import hashlib
import json
import os
import re
import sys
import tempfile
import traceback
from contextlib import AsyncExitStack
from datetime import datetime, timezone
from pathlib import Path
from urllib.parse import urlsplit

# Preserve nesting and ordering without guessing Study4's question selectors.
# Do not export scripts, hidden credentials, comments, navigation, or input values.
EXPORT_JS = r"""() => {
  const selector = __SELECTOR__ || (document.querySelector('.test-content') ? '.test-content' : '.practice-test-wrapper');
  if (location.hostname !== 'study4.com' && location.hostname !== 'www.study4.com')
    throw new Error('Select a study4.com tab before exporting.');
  const roots = [...document.querySelectorAll(selector)];
  if (!roots.length) throw new Error('Content selector matched no elements: ' + selector);
  const skip = 'script,style,noscript,iframe,nav,header,footer,textarea,' +
    'input[type="hidden"],input[type="password"],input[type="email"],' +
    '#test-comments,.test-comments-wrapper,[contenteditable="true"]';
  const assets = new Map();
  function asset(value, kind) {
    if (!value) return null;
    const url = new URL(value, document.baseURI);
    if (!['https:', 'http:', 'blob:'].includes(url.protocol)) return null;
    const key = url.href;
    if (!assets.has(key)) assets.set(key, {id: 'asset-' + (assets.size + 1), kind, url: key});
    return assets.get(key).id;
  }
  function walk(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent.replace(/\s+/g, ' ').trim();
      return text ? {text} : null;
    }
    if (node.nodeType !== Node.ELEMENT_NODE || node.matches(skip)) return null;
    const tag = node.tagName.toLowerCase();
    const out = {tag};
    if (node.id) out.id = node.id;
    if (typeof node.className === 'string' && node.className) out.class = node.className;
    if (node.getAttribute('role')) out.role = node.getAttribute('role');
    for (const attr of ['colspan', 'rowspan'])
      if (node.hasAttribute(attr)) out[attr] = node.getAttribute(attr);
    if (tag === 'br') return out;
    if (tag === 'img') {
      out.alt = node.alt;
      out.asset = asset(node.currentSrc || node.getAttribute('src') ||
        node.getAttribute('data-src'), 'image');
    }
    if (['audio', 'video', 'source'].includes(tag))
      out.asset = asset(node.currentSrc || node.getAttribute('src'),
        tag === 'source' ? node.parentElement.tagName.toLowerCase() : tag);
    if (tag === 'input') {
      if (!['radio', 'checkbox'].includes(node.type)) return null;
      out.type = node.type;
      out.value = node.value;
      out.labels = [...(node.labels || [])].map(label => label.textContent.trim());
    }
    if (tag === 'a') {
      const raw = node.getAttribute('href');
      if (raw) {
        const url = new URL(raw, document.baseURI);
        if (url.origin === location.origin && url.pathname.startsWith('/tests/'))
          out.href = url.origin + url.pathname + url.search + url.hash;
      }
    }
    const children = [...node.childNodes].map(walk).filter(Boolean);
    if (children.length) out.children = children;
    return children.length || out.asset || tag === 'input' ? out : null;
  }
  const landingNode = document.querySelector('#saved-attempt-landing-data');
  let parts = null;
  if (landingNode) parts = JSON.parse(landingNode.textContent).parts;
  const counts = {
    parts: roots.reduce((n, root) => n + root.querySelectorAll('.test-questions-wrapper').length, 0),
    questions: roots.reduce((n, root) => n + root.querySelectorAll('.question-wrapper').length, 0),
    explanation_blocks: roots.reduce((n, root) => n + root.querySelectorAll('.question-explanation-wrapper').length, 0),
    transcript_blocks: roots.reduce((n, root) => n + root.querySelectorAll('.context-transcript').length, 0)
  };
  return {
    schema_version: 1,
    source: {url: location.origin + location.pathname, title: document.title},
    selector,
    page_kind: landingNode ? 'landing' : (counts.questions ? 'exam' : 'unclassified'),
    complete_exam_verified: false,
    counts,
    parts,
    content: roots.map(walk).filter(Boolean),
    assets: [...assets.values()],
    notes: ['DOM export; question/answer schema and completeness are not verified.',
      'Includes collapsed DOM content; excludes scripts, comments and credential inputs.',
      'Media URLs only; blob URLs require the original browser session.',
      'Question bodies, choices, explanations and transcripts remain nested in the DOM tree.']
  };
}"""


def redact_secrets(text):
    return re.sub(r"(?i)([?&](?:token|access_token|key)=)[^&\s)]+", r"\1[REDACTED]", text)


def check_tool_result(result, tool_name):
    """Keep server error details instead of replacing them with a generic hint."""
    if result.isError:
        details = "\n".join(
            block.text for block in result.content if getattr(block, "type", None) == "text"
        ).strip()
        details = redact_secrets(details)
        raise RuntimeError(f"{tool_name}: {details[:6000] or 'MCP returned an error without details.'}")


def describe_error(error):
    """Unwrap AnyIO task groups and chained exceptions at the CLI boundary."""
    lines = []
    seen = set()

    def visit(current):
        if id(current) in seen:
            return
        seen.add(id(current))
        if isinstance(current, BaseExceptionGroup):
            for child in current.exceptions:
                visit(child)
            return
        message = str(current) or "No error message."
        if isinstance(current, TimeoutError):
            message = "Timed out. Check the Playwright Extension connection or increase --timeout."
        lines.append(f"{type(current).__name__}: {message}")
        cause = current.__cause__
        if cause is None and not current.__suppress_context__:
            cause = current.__context__
        if cause is not None:
            visit(cause)

    visit(error)
    return redact_secrets("\n".join(lines))


async def select_study4_tab(session, args, target_url=None):
    """Select an existing exam tab; never export the extension's connect page."""
    url = target_url if target_url is not None else getattr(args, "url", None)
    if isinstance(url, list):
        url = url[0] if url else None
    if url:
        parsed = urlsplit(url)
        if parsed.scheme != "https" or parsed.hostname not in {"study4.com", "www.study4.com"}:
            raise ValueError("--url must be an HTTPS Study4 URL.")
        if parsed.username or parsed.password:
            raise ValueError("Do not embed credentials in --url.")
    result = await session.call_tool("browser_tabs", {"action": "list"})
    check_tool_result(result, "browser_tabs")
    tabs = []
    for block in result.content:
        if getattr(block, "type", None) != "text":
            continue
        for line in block.text.splitlines():
            match = re.match(r"^- (\d+):.*\]\((https?://.*)\)(?: \[.*\])?$", line)
            if not match:
                continue
            tab_url = match[2]
            if urlsplit(tab_url).hostname in {"study4.com", "www.study4.com"}:
                tabs.append((int(match[1]), tab_url))
    if getattr(args, "tab_index", None) is not None:
        candidates = [tab for tab in tabs if tab[0] == args.tab_index]
        if not candidates:
            raise RuntimeError("--tab-index does not identify a Study4 tab in this MCP connection.")
        if url and candidates[0][1] != url:
            raise RuntimeError("Selected tab does not match --url; use one selection option.")
    elif url:
        candidates = [tab for tab in tabs if tab[1] == url]
        if not candidates:
            result = await session.call_tool("browser_tabs", {"action": "new", "url": url})
            check_tool_result(result, "browser_tabs")
            return None
        # Duplicate tabs of the same exact URL are safe to reuse deterministically.
        candidates = candidates[:1]
    else:
        candidates = tabs
    if len(candidates) != 1:
        choices = ", ".join(f"{index}: {urlsplit(u).path}" for index, u in candidates)
        raise RuntimeError(
            "No unique Study4 tab. Pass --url with the exact exam URL or --tab-index. "
            f"Study4 tabs: {choices or 'none visible to MCP'}."
        )
    selected_index = candidates[0][0]
    result = await session.call_tool("browser_tabs", {"action": "select", "index": selected_index})
    check_tool_result(result, "browser_tabs")
    return selected_index


async def close_study4_tab(session, tab_index=None):
    """Close tab to conserve browser resources across batch runs."""
    try:
        params = {"action": "close"}
        if tab_index is not None:
            params["index"] = tab_index
        await session.call_tool("browser_tabs", params)
    except Exception:
        pass


async def follow_result_details(session):
    """Discover the site's actual review link; never start or submit an attempt."""
    result = await session.call_tool(
        "browser_evaluate",
        {
            "function": r"""() => {
      if (!['study4.com', 'www.study4.com'].includes(location.hostname))
        throw new Error('Select a Study4 page first.');
      let expectedPath;
      if (/\/results\/\d+\/$/.test(location.pathname)) expectedPath = location.pathname + 'details/';
      else if (/^\/tests\/\d+\/[^/]+\/$/.test(location.pathname)) expectedPath = location.pathname + 'solutions/';
      else return null;
      const link = [...document.querySelectorAll('a[href]')].find(a => {
        const url = new URL(a.href, location.href);
        return url.origin === location.origin && url.pathname === expectedPath;
      });
      return link ? new URL(link.href).origin + expectedPath : null;
    }"""
        },
    )
    check_tool_result(result, "browser_evaluate")
    for block in result.content:
        if getattr(block, "type", None) != "text":
            continue
        raw = block.text.strip()
        if "### Result\n" in raw:
            raw = raw.split("### Result\n", 1)[1].split("\n### ", 1)[0].strip()
        try:
            url = json.loads(raw)
        except json.JSONDecodeError:
            continue
        if url is None:
            return
        if isinstance(url, str):
            parsed = urlsplit(url)
            if (
                parsed.scheme != "https"
                or parsed.hostname not in {"study4.com", "www.study4.com"}
                or not re.fullmatch(r"/tests/\d+/[^/]+/(?:results/\d+/details/|solutions/)", parsed.path)
            ):
                raise RuntimeError("Unexpected result review URL; export stopped.")
            print("Opening the full solutions/review page linked by Study4...")
            await select_study4_tab(session, argparse.Namespace(url=url, tab_index=None))
            return
    raise RuntimeError("Could not read the result review link from MCP response.")


def decode_export(result):
    """Decode Playwright's Result section, rejecting tool errors/truncated JSON."""
    check_tool_result(result, "browser_evaluate")
    for block in result.content:
        if getattr(block, "type", None) != "text":
            continue
        raw = block.text.strip()
        candidates = [raw]
        if "### Result\n" in raw:
            candidates.insert(0, raw.split("### Result\n", 1)[1].split("\n### ", 1)[0].strip())
        for candidate in candidates:
            if candidate.startswith("```"):
                candidate = "\n".join(candidate.splitlines()[1:-1])
            try:
                value = json.loads(candidate)
            except json.JSONDecodeError:
                continue
            if isinstance(value, dict) and value.get("schema_version") == 1:
                if not isinstance(value.get("content"), list) or not value["content"]:
                    raise RuntimeError("Export contains no content; choose a narrower content selector.")
                return value
    raise RuntimeError("No complete export JSON in MCP response. Output may be truncated.")


def media_format(path, kind):
    """Validate actual bytes, not the URL extension or a server MIME claim."""
    if kind == "image":
        from PIL import Image

        with Image.open(path) as image:
            image.verify()
            content_type = Image.MIME[image.format]
        extensions = {
            "image/jpeg": ".jpg",
            "image/png": ".png",
            "image/webp": ".webp",
            "image/gif": ".gif",
            "image/avif": ".avif",
        }
    elif kind == "audio":
        from mutagen import File

        audio = File(path)
        if audio is None or not audio.info.length > 0:
            raise ValueError("Downloaded audio has no valid duration")
        content_type = audio.mime[0]
        content_type = {
            "audio/mp3": "audio/mpeg",
            "audio/mpg": "audio/mpeg",
            "audio/x-mpeg": "audio/mpeg",
        }.get(content_type, content_type)
        extensions = {
            "audio/mpeg": ".mp3",
            "audio/mp4": ".m4a",
            "audio/x-wav": ".wav",
            "audio/wav": ".wav",
            "audio/ogg": ".ogg",
            "audio/flac": ".flac",
            "audio/x-flac": ".flac",
        }
    else:
        raise ValueError(f"Unsupported media kind: {kind}")
    if content_type not in extensions:
        raise ValueError(f"Unsupported validated media type: {content_type}")
    return content_type, extensions[content_type]


def get_minio_client_and_bucket(args=None):
    from minio import Minio

    env = {}
    try:
        from dotenv import dotenv_values

        env.update(dotenv_values(Path(__file__).resolve().parents[1] / ".env"))
        env.update(dotenv_values(Path(__file__).resolve().parents[1] / "apps" / "api" / ".env"))
    except Exception:
        pass
    env.update(os.environ)

    raw_endpoint = (
        getattr(args, "minio_endpoint", None) or env.get("MINIO_ENDPOINT") or "http://localhost:9000"
    )
    parsed = urlsplit(raw_endpoint)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        raise ValueError(f"MINIO_ENDPOINT must be an HTTP(S) URL: {raw_endpoint}")
    user = getattr(args, "minio_user", None) or env.get("MINIO_ROOT_USER") or "engflex"
    password = getattr(args, "minio_password", None) or env.get("MINIO_ROOT_PASSWORD") or "engflex-dev-secret"
    bucket = getattr(args, "minio_bucket", None) or env.get("MINIO_BUCKET") or "lesson-audio"

    client = Minio(
        parsed.netloc,
        access_key=user,
        secret_key=password,
        secure=parsed.scheme == "https",
    )
    if not client.bucket_exists(bucket):
        client.make_bucket(bucket)
    return client, bucket


def find_q_numbers(node):
    if not isinstance(node, dict):
        return []
    res = []
    cls = node.get("class", "")
    if "question-number" in cls:
        for c in node.get("children", []):
            if "text" in c and c["text"].strip().isdigit():
                res.append(int(c["text"].strip()))
            for sc in c.get("children", []):
                if "text" in sc and sc["text"].strip().isdigit():
                    res.append(int(sc["text"].strip()))
    for c in node.get("children", []):
        res.extend(find_q_numbers(c))
    return res


def compute_asset_names(content, assets):
    asset_to_q = {}

    def walk_tree(node, current_wrapper_q=None):
        if not isinstance(node, dict):
            return
        cls = node.get("class", "")
        wrapper_q = current_wrapper_q
        if any(k in cls for k in ("question-item-wrapper", "question-group-wrapper")):
            q_nums = find_q_numbers(node)
            if q_nums:
                wrapper_q = sorted(list(set(q_nums)))
        if "asset" in node:
            aid = node["asset"]
            if aid not in asset_to_q and wrapper_q:
                asset_to_q[aid] = wrapper_q
        for child in node.get("children", []):
            walk_tree(child, wrapper_q)

    for root in content:
        walk_tree(root)

    name_counts = {}
    names = {}
    audio_idx = 0
    for a in assets:
        aid = a["id"]
        kind = a.get("kind")
        if kind == "audio":
            audio_idx += 1
            base = "audio_full" if audio_idx == 1 else f"audio_{audio_idx}"
        elif aid in asset_to_q:
            qs = asset_to_q[aid]
            if len(qs) == 1:
                base = f"q{qs[0]:02d}"
            else:
                base = f"q{qs[0]}-{qs[-1]}"
        else:
            base = None

        if base:
            count = name_counts.get(base, 0)
            name_counts[base] = count + 1
            name = base if count == 0 else f"{base}_{count + 1}"
            names[aid] = name
        else:
            names[aid] = None

    return names


def download_asset(
    asset,
    test_root,
    client,
    minio_client=None,
    minio_bucket=None,
    keep_local_media=True,
    suggested_name=None,
):
    url = asset["url"]
    parsed = urlsplit(url)
    if parsed.scheme not in {"http", "https"} or not parsed.netloc or parsed.username or parsed.password:
        raise ValueError(
            "Asset requires an HTTP(S) URL; blob media must be exported from its browser session"
        )
    staging = test_root / "media" / ".staging"
    staging.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=staging, suffix=".part", delete=False) as handle:
        temporary = Path(handle.name)
    try:
        digest = hashlib.sha256()
        size = 0
        with client.stream("GET", url) as response, temporary.open("wb") as output:
            response.raise_for_status()
            if urlsplit(str(response.url)).scheme not in {"http", "https"}:
                raise ValueError("Invalid asset redirect")
            for chunk in response.iter_bytes(65536):
                output.write(chunk)
                digest.update(chunk)
                size += len(chunk)
        if not size:
            raise ValueError("Downloaded asset is empty")
        content_type, extension = media_format(temporary, asset["kind"])
        checksum = digest.hexdigest()
        if suggested_name:
            filename = (
                suggested_name if suggested_name.endswith(extension) else f"{suggested_name}{extension}"
            )
        else:
            filename = f"{checksum}{extension}"
        # The web media route resolves the content hash, not a mutable test filename.
        minio_key = f"exams/toeic/{checksum}"

        if minio_client is not None and minio_bucket:
            minio_client.fput_object(
                bucket_name=minio_bucket,
                object_name=minio_key,
                file_path=str(temporary),
                content_type=content_type,
            )

        local_path = None
        if keep_local_media:
            directory = "images" if asset["kind"] == "image" else "audio"
            relative = Path("media") / directory / f"{checksum}{extension}"
            destination = test_root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            if destination.exists():
                with destination.open("rb") as existing:
                    if hashlib.file_digest(existing, "sha256").hexdigest() != checksum:
                        raise ValueError("Existing local asset failed SHA-256 verification")
            else:
                import shutil

                shutil.copyfile(temporary, destination)
            local_path = relative.as_posix()

        return {
            **asset,
            "status": "uploaded" if minio_client else "downloaded",
            "sourceId": local_path or minio_key,
            "localPath": local_path,
            "filename": filename,
            "sha256": checksum,
            "bytes": size,
            "contentType": content_type,
            "minioKey": minio_key,
            "minioBucket": minio_bucket if minio_client else None,
        }
    finally:
        temporary.unlink(missing_ok=True)


def save_json(path, value):
    with path.open("x", encoding="utf-8") as handle:
        json.dump(value, handle, ensure_ascii=False, indent=2)
        handle.write("\n")


def save_bundle(data, args):
    import httpx

    source = urlsplit(data.get("source", {}).get("url", ""))
    match = re.match(r"^/tests/(\d+)(?:/|$)", source.path)
    if source.hostname not in {"study4.com", "www.study4.com"} or not match:
        raise ValueError("Cannot organize export without a Study4 exam source URL")
    test_id = match[1]
    test_root = Path(args.crawl_root).resolve() / "study4" / "tests" / test_id
    run_id = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ")
    run_root = test_root / "runs" / run_id
    run_root.mkdir(parents=True, exist_ok=False)
    # Keep the source before starting network I/O; failed downloads remain reviewable.
    save_json(run_root / "raw.json", data)

    use_minio = getattr(args, "minio", False)
    keep_local = not use_minio or getattr(args, "keep_local_media", False)
    minio_client = None
    minio_bucket = None
    if use_minio:
        minio_client, minio_bucket = get_minio_client_and_bucket(args)

    asset_name_map = compute_asset_names(data.get("content", []), data.get("assets", []))

    manifest = {
        "schemaVersion": "1.0.0",
        "source": data["source"],
        "sourceTestId": test_id,
        "runId": run_id,
        "assetRoot": "../..",
        "status": "downloading",
        "objectKeyPattern": (
            None if getattr(args, "exam_type", "toeic") == "ielts-reading" else "exams/toeic/{sha256}"
        ),
        "minioBucket": minio_bucket,
        "assets": [],
    }
    with httpx.Client(
        follow_redirects=True,
        timeout=60,
        headers={"User-Agent": "Engflex-Study4-Exporter/1.0", "Referer": data["source"]["url"]},
    ) as client:
        for asset in data["assets"]:
            try:
                entry = download_asset(
                    asset,
                    test_root,
                    client,
                    minio_client=minio_client,
                    minio_bucket=minio_bucket,
                    keep_local_media=keep_local,
                    suggested_name=asset_name_map.get(asset["id"]),
                )
                if minio_client:
                    print(
                        f"Uploaded {asset['id']} -> MinIO [{minio_bucket}] {entry['minioKey']} ({entry['bytes']} bytes)"
                    )
                else:
                    print(f"Downloaded {asset['id']}: {entry['localPath']} ({entry['bytes']} bytes)")
            except Exception as error:
                entry = {**asset, "status": "failed", "error": redact_secrets(str(error))}
                print(f"Download/Upload failed for {asset['id']}: {entry['error']}", file=sys.stderr)
            manifest["assets"].append(entry)
    failures = sum(entry["status"] == "failed" for entry in manifest["assets"])
    manifest["status"] = "incomplete" if failures else "complete"
    manifest["downloadedCount"] = len(manifest["assets"]) - failures
    manifest["failedCount"] = failures
    save_json(run_root / "media-manifest.json", manifest)
    # Same asset IDs keep every DOM reference resolvable. Paths are relative to test_root.
    packaged = {
        **data,
        "assets": manifest["assets"],
        "media_manifest": "media-manifest.json",
        "asset_root": "../..",
        "media_download_complete": not failures,
    }
    save_json(run_root / "export.json", packaged)
    if getattr(args, "output", None):
        alias = Path(args.output).resolve()
        alias.parent.mkdir(parents=True, exist_ok=True)
        alias_data = {
            **packaged,
            "asset_root": os.path.relpath(test_root, alias.parent).replace("\\", "/"),
            "media_manifest": os.path.relpath(run_root / "media-manifest.json", alias.parent).replace(
                "\\", "/"
            ),
        }
        save_json(alias, alias_data)
    print(f"Bundle saved: {run_root}")
    if minio_client:
        print(
            f"Media: {manifest['downloadedCount']} uploaded to MinIO ({minio_bucket}); {failures} failed. No local media retained."
        )
    else:
        print(f"Media: {manifest['downloadedCount']} downloaded; {failures} failed. Not uploaded to MinIO.")
    if failures:
        raise RuntimeError(
            f"Media bundle incomplete: {failures} failed; see {run_root / 'media-manifest.json'}"
        )
    return run_root


def resolve_target_urls(args):
    raw_urls = []
    url_attr = getattr(args, "url", None)
    if url_attr:
        if isinstance(url_attr, list):
            raw_urls.extend(url_attr)
        else:
            raw_urls.append(url_attr)
    urls_attr = getattr(args, "urls", None)
    if urls_attr:
        if isinstance(urls_attr, list):
            raw_urls.extend(urls_attr)
        else:
            raw_urls.append(urls_attr)
    urls_file_attr = getattr(args, "urls_file", None)
    if urls_file_attr:
        file_path = Path(urls_file_attr).resolve()
        if not file_path.is_file():
            raise FileNotFoundError(f"URLs file not found: {file_path}")
        for line in file_path.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if line and not line.startswith("#"):
                raw_urls.append(line)

    valid_urls = []
    for u in raw_urls:
        u = u.strip()
        if not u:
            continue
        parsed = urlsplit(u)
        if parsed.scheme != "https" or parsed.hostname not in {"study4.com", "www.study4.com"}:
            raise ValueError(f"Invalid Study4 URL: {u}. URLs must be HTTPS and on study4.com.")
        if parsed.username or parsed.password:
            raise ValueError(f"Do not embed credentials in URL: {u}")
        valid_urls.append(u)

    return list(dict.fromkeys(valid_urls))


async def export_single_test(session, target_url, args, tab_index=None):
    await select_study4_tab(session, args, target_url=target_url)
    try:
        await follow_result_details(session)
        script = EXPORT_JS.replace("__SELECTOR__", json.dumps(args.selector))
        result = await session.call_tool("browser_evaluate", {"function": script})
        data = decode_export(result)
        data["exported_at"] = datetime.now(timezone.utc).isoformat()
        run_root = await asyncio.to_thread(save_bundle, data, args)
        validate_exam_export(data, getattr(args, "exam_type", "toeic"))
    finally:
        if tab_index is None:
            # Following a results link can select a different tab. Close the
            # current export tab rather than an index that may now be stale.
            await close_study4_tab(session)
    if data.get("counts"):
        print("DOM counts: " + json.dumps(data["counts"]))
    return {
        "url": target_url,
        "status": "success",
        "run_root": str(run_root),
        "page_kind": data.get("page_kind"),
        "counts": data.get("counts"),
        "exported_at": data["exported_at"],
    }


def validate_exam_export(data, exam_type="toeic"):
    """An archived DOM is not a successful full-test crawl until coverage is checked."""
    from study4_content import exam_items, nodes, question_number

    if exam_type == "ielts-reading":
        if data.get("page_kind") != "exam":
            raise ValueError("Open the IELTS Reading review page with all questions visible to your account")
        sections = [
            part
            for root in data.get("content", [])
            for part in nodes(root, "test-questions-wrapper")
        ]
        if len(sections) != 3:
            raise ValueError(f"Expected three IELTS Reading sections, found {len(sections)}")
        questions_by_section = [nodes(section, "question-wrapper") for section in sections]
        if not all(questions_by_section):
            raise ValueError("Every IELTS Reading section must contain questions")
        numbers = [
            question_number(question)
            for questions in questions_by_section
            for question in questions
        ]
        if numbers != list(range(1, 41)):
            raise ValueError("Expected IELTS Reading questions 1–40 in order and exactly once")
        return

    exam_items(data)
    if len([asset for asset in data.get("assets", []) if asset.get("kind") == "audio"]) != 1:
        raise ValueError("A full-test export requires one complete listening audio asset; the solution index or a partial attempt is insufficient")


async def export_study4_page(args):
    if args.output and Path(args.output).exists():
        raise FileExistsError(f"Output already exists: {args.output}; omit --output for a new organized run")
    # Lazy imports let --help work before installing the optional MCP dependency.
    from mcp import ClientSession, StdioServerParameters
    from mcp.client.stdio import stdio_client
    from mcp.client.streamable_http import streamablehttp_client

    async with AsyncExitStack() as stack:
        if args.mcp_url:
            transport = streamablehttp_client(args.mcp_url)
        else:
            command = args.command or ("npx.cmd" if os.name == "nt" else "npx")
            server_args = args.server_arg or ["-y", "@playwright/mcp@latest", "--extension"]
            if args.browser_executable:
                executable = Path(args.browser_executable).resolve()
                if not executable.is_file():
                    raise FileNotFoundError(f"Browser executable not found: {executable}")
                server_args = [*server_args, "--executable-path", str(executable)]
            # MCP inherits only a safe allowlist, so forward this extension token explicitly.
            token = os.environ.get("PLAYWRIGHT_MCP_EXTENSION_TOKEN")
            if not token:
                try:
                    from dotenv import dotenv_values

                    # Resolve relative to the script, independent of the terminal's cwd.
                    token = dotenv_values(Path(__file__).resolve().parents[1] / ".env").get(
                        "PLAYWRIGHT_MCP_EXTENSION_TOKEN"
                    )
                except ImportError:
                    token = None
            server_env = {"PLAYWRIGHT_MCP_EXTENSION_TOKEN": token} if token else None
            transport = stdio_client(StdioServerParameters(command=command, args=server_args, env=server_env))
        streams = await stack.enter_async_context(transport)
        session = await stack.enter_async_context(ClientSession(streams[0], streams[1]))
        await session.initialize()
        names = set()
        cursor = None
        while True:
            page = await session.list_tools(cursor=cursor)
            names.update(tool.name for tool in page.tools)
            cursor = page.nextCursor
            if not cursor:
                break
        if args.list_tools:
            print("\n".join(sorted(names)))
            return
        if "browser_evaluate" not in names:
            raise RuntimeError("This server does not expose Playwright's browser_evaluate tool.")
        if "browser_tabs" not in names:
            raise RuntimeError("This server does not expose Playwright's browser_tabs tool.")
        if args.diagnose:
            await select_study4_tab(session, args)
            result = await session.call_tool(
                "browser_evaluate",
                {
                    "function": """() => ({
                hostname: location.hostname,
                pathname: location.pathname,
                readyState: document.readyState
            })"""
                },
            )
            check_tool_result(result, "browser_evaluate")
            for block in result.content:
                if getattr(block, "type", None) == "text":
                    print(redact_secrets(block.text))
            print("Browser connection OK. No exam content exported.")
            return

        urls = resolve_target_urls(args)
        if not urls:
            urls = [getattr(args, "url", None)]
        total = len(urls)
        batch_timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ")
        results = []

        for idx, target_url in enumerate(urls, 1):
            if total > 1:
                print("\n==================================================")
                print(f"[{idx}/{total}] Processing: {target_url}")
                print("==================================================")
            item_args = argparse.Namespace(**vars(args))
            if total > 1:
                item_args.output = None  # Do not overwrite single output file in batch mode

            max_attempts = 2 if total > 1 else 1
            last_error = None
            item_result = None

            timeout = getattr(args, "timeout", 600)
            for attempt in range(1, max_attempts + 1):
                try:
                    item_result = await asyncio.wait_for(
                        export_single_test(
                            session, target_url, item_args, tab_index=getattr(args, "tab_index", None)
                        ),
                        timeout=timeout,
                    )
                    break
                except Exception as err:
                    last_error = err
                    if total > 1:
                        print(
                            f"⚠️ [{idx}/{total}] Attempt {attempt}/{max_attempts} failed for {target_url}:\n{describe_error(err)}",
                            file=sys.stderr,
                        )
                        if attempt < max_attempts:
                            print("Retrying in 2 seconds...", file=sys.stderr)
                            await asyncio.sleep(2)

            if item_result:
                results.append(item_result)
            else:
                if total == 1:
                    raise last_error
                results.append(
                    {
                        "url": target_url,
                        "status": "failed",
                        "error": describe_error(last_error),
                    }
                )

        if total > 1:
            successful = sum(1 for r in results if r["status"] == "success")
            failed = sum(1 for r in results if r["status"] == "failed")
            summary = {
                "schema_version": 1,
                "batch_id": batch_timestamp,
                "created_at": datetime.now(timezone.utc).isoformat(),
                "total": total,
                "successful": successful,
                "failed": failed,
                "results": results,
            }
            summary_path = (
                Path(args.crawl_root).resolve() / "study4" / f"batch-summary-{batch_timestamp}.json"
            )
            summary_path.parent.mkdir(parents=True, exist_ok=True)
            save_json(summary_path, summary)
            if args.output:
                save_json(Path(args.output).resolve(), summary)
            print("\n==================================================")
            print(f"Batch completed: {successful}/{total} successful, {failed} failed.")
            print(f"Batch summary saved: {summary_path}")
            print("==================================================")
            if failed > 0:
                raise RuntimeError(f"Batch completed with {failed} failure(s); see {summary_path}")


def prompt_study4_urls():
    while True:
        try:
            raw = input("Nhap 1 hoac nhieu link Study4 (ngan cach boi dau phay hoac xuong dong): ").strip()
        except EOFError as error:
            raise RuntimeError("No interactive input available. Pass --url with the Study4 link.") from error
        tokens = re.split(r"[\s,]+", raw)
        valid_urls = []
        for token in tokens:
            token = token.strip()
            if not token:
                continue
            parsed = urlsplit(token)
            if (
                parsed.scheme == "https"
                and parsed.hostname in {"study4.com", "www.study4.com"}
                and not parsed.username
                and not parsed.password
            ):
                valid_urls.append(token)
        if valid_urls:
            return list(dict.fromkeys(valid_urls))
        print("Link chua hop le. Hay dan link https://study4.com/...", file=sys.stderr)


def prompt_study4_url():
    urls = prompt_study4_urls()
    return urls[0] if urls else ""


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    transport = parser.add_mutually_exclusive_group()
    transport.add_argument("--mcp-url", help="Streamable HTTP endpoint, e.g. http://localhost:8931/mcp")
    transport.add_argument("--command", help="stdio MCP executable (default: npx.cmd on Windows)")
    parser.add_argument(
        "--server-arg", action="append", help="Repeat for each stdio argument; use = for --flags"
    )
    parser.add_argument(
        "--url",
        action="append",
        help="Select this exact Study4 URL, or open a new tab if not already open (can be repeated)",
    )
    parser.add_argument(
        "--urls",
        nargs="+",
        help="One or more Study4 URLs separated by spaces",
    )
    parser.add_argument(
        "--urls-file",
        type=Path,
        help="Text file containing Study4 URLs, one per line",
    )
    parser.add_argument("--tab-index", type=int, help="Select an existing Study4 tab by its MCP tab index")
    parser.add_argument(
        "--exam-type", choices=("toeic", "ielts-reading"), default="toeic",
        help="Question coverage to validate (default: toeic)",
    )
    parser.add_argument(
        "--selector",
        default=None,
        help="CSS content scope; auto-detect .test-content or .practice-test-wrapper",
    )
    parser.add_argument(
        "--output",
        help="Optional additional JSON copy or batch summary; organized bundle is always created",
    )
    parser.add_argument(
        "--crawl-root", default="docs/crawl", help="Root for per-test media and timestamped runs"
    )
    parser.add_argument(
        "--from-json", type=Path, help="Package/download media from a previous raw export without opening MCP"
    )
    parser.add_argument(
        "--browser-executable", help="Browser executable for Playwright extension mode (stdio only)"
    )
    parser.add_argument("--list-tools", action="store_true")
    parser.add_argument(
        "--diagnose", action="store_true", help="Test extension connection without exporting content"
    )
    parser.add_argument(
        "--debug", action="store_true", help="Print the complete exception traceback on failure"
    )
    parser.add_argument(
        "--timeout", type=float, default=600, help="Per-test export timeout in seconds (default: 600)"
    )
    parser.add_argument(
        "--minio",
        action=argparse.BooleanOptionalAction,
        default=None,
        help="Upload TOEIC media to MinIO (on by default for TOEIC, off for IELTS Reading)",
    )
    parser.add_argument(
        "--keep-local-media",
        action="store_true",
        default=False,
        help="Keep a local copy of downloaded media files in docs/crawl (default: False)",
    )
    parser.add_argument(
        "--minio-endpoint",
        help="MinIO endpoint URL (default: from MINIO_ENDPOINT or http://localhost:9000)",
    )
    parser.add_argument(
        "--minio-bucket",
        help="MinIO bucket name (default: from MINIO_BUCKET or lesson-audio)",
    )
    args = parser.parse_args()
    if args.exam_type == "ielts-reading" and args.minio:
        parser.error("IELTS Reading media is local-only")
    if args.minio is None:
        args.minio = args.exam_type == "toeic"
    if args.mcp_url and (args.server_arg or args.browser_executable):
        parser.error("--server-arg and --browser-executable apply only to stdio, not --mcp-url")
    if args.timeout <= 0:
        parser.error("--timeout must be positive")
    try:
        if args.output and Path(args.output).exists():
            raise FileExistsError(f"Output already exists: {args.output}")
        if args.from_json:
            data = json.loads(args.from_json.read_text(encoding="utf-8"))
            save_bundle(data, args)
            validate_exam_export(data, args.exam_type)
            return 0
        target_urls = resolve_target_urls(args)
        if not target_urls and args.tab_index is None and not (args.list_tools or args.diagnose):
            target_urls = prompt_study4_urls()
            args.url = target_urls
        if not args.mcp_url and not args.browser_executable:
            try:
                from dotenv import dotenv_values

                args.browser_executable = os.environ.get("PLAYWRIGHT_MCP_EXECUTABLE_PATH") or dotenv_values(
                    Path(__file__).resolve().parents[1] / ".env"
                ).get("PLAYWRIGHT_MCP_EXECUTABLE_PATH")
            except ImportError:
                args.browser_executable = os.environ.get("PLAYWRIGHT_MCP_EXECUTABLE_PATH")
        asyncio.run(export_study4_page(args))
    except KeyboardInterrupt:
        return 130
    except Exception as error:
        print(f"Export failed:\n{describe_error(error)}", file=sys.stderr)
        if args.debug:
            print(redact_secrets("".join(traceback.format_exception(error))), file=sys.stderr)
        else:
            print("Run again with --debug for the complete traceback.", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
