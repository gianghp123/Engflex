"""Offline checks; no browser, credentials, network, or MCP installation needed."""

import asyncio
import hashlib
import importlib.util
import io
import json
import sys
import tempfile
import unittest
import wave
from argparse import Namespace
from contextlib import asynccontextmanager
from pathlib import Path
from types import ModuleType, SimpleNamespace
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[3] / "scripts"))

spec = importlib.util.spec_from_file_location(
    "study4_export", Path(__file__).resolve().parents[3] / "scripts/study4_mcp_export.py"
)
export = importlib.util.module_from_spec(spec)
spec.loader.exec_module(export)


def response(text, error=False):
    return SimpleNamespace(isError=error, content=[SimpleNamespace(type="text", text=text)])


class ExportTests(unittest.TestCase):
    def test_ielts_reading_accepts_complete_questions_without_audio(self):
        sections = [
            {
                "class": "test-questions-wrapper",
                "children": [
                    {
                        "class": "question-wrapper",
                        "children": [
                            {"class": "question-number", "children": [{"text": str(number)}]}
                        ],
                    }
                    for number in range(start, end)
                ],
            }
            for start, end in ((1, 15), (15, 28), (28, 41))
        ]
        data = {"page_kind": "exam", "content": sections, "assets": []}
        export.validate_exam_export(data, "ielts-reading")

        sections[-1]["children"].pop()
        with self.assertRaisesRegex(ValueError, "questions 1–40"):
            export.validate_exam_export(data, "ielts-reading")

    def test_no_minio_retains_media_even_when_keep_local_flag_is_false(self):
        import httpx
        from PIL import Image

        body = io.BytesIO()
        Image.new("RGB", (2, 2), "red").save(body, format="PNG")
        original_client = httpx.Client
        with tempfile.TemporaryDirectory() as folder:
            data = {
                "source": {"url": "https://study4.com/tests/9999/review/"},
                "content": [{"asset": "asset-1"}],
                "assets": [{"id": "asset-1", "kind": "image", "url": "https://media.example/image"}],
            }
            test_root = Path(folder) / "study4/tests/9999"
            old_media = test_root / "media/images/previous.png"
            old_media.parent.mkdir(parents=True)
            old_media.write_bytes(b"previous run")
            transport = httpx.MockTransport(lambda request: httpx.Response(200, content=body.getvalue()))
            with patch(
                "httpx.Client", side_effect=lambda **kwargs: original_client(transport=transport, **kwargs)
            ):
                run = export.save_bundle(
                    data, Namespace(crawl_root=folder, output=None, minio=False, keep_local_media=False)
                )
            manifest = json.loads((run / "media-manifest.json").read_text(encoding="utf-8"))
            asset = manifest["assets"][0]
            self.assertEqual((test_root / asset["localPath"]).read_bytes(), body.getvalue())
            self.assertEqual(old_media.read_bytes(), b"previous run")
            self.assertEqual(manifest["status"], "complete")

    def test_named_media_has_immutable_hash_keys(self):
        import httpx
        from PIL import Image

        with tempfile.TemporaryDirectory() as folder:
            results = []
            for color in ("red", "blue"):
                body = io.BytesIO()
                Image.new("RGB", (2, 2), color).save(body, format="PNG")
                transport = httpx.MockTransport(lambda request: httpx.Response(200, content=body.getvalue()))
                with httpx.Client(transport=transport) as client:
                    results.append(
                        export.download_asset(
                            {"id": "photo", "kind": "image", "url": "https://media.example/image"},
                            Path(folder),
                            client,
                            suggested_name="q01",
                        )
                    )
            self.assertNotEqual(results[0]["minioKey"], results[1]["minioKey"])
            self.assertNotEqual(results[0]["localPath"], results[1]["localPath"])
            self.assertTrue(all((Path(folder) / result["localPath"]).is_file() for result in results))

    def test_mp3_mime_aliases(self):
        for mime in ("audio/mp3", "audio/mpg", "audio/x-mpeg", "audio/mpeg"):
            with self.subTest(mime=mime):
                audio = SimpleNamespace(info=SimpleNamespace(length=1.0), mime=[mime])
                with patch("mutagen.File", return_value=audio):
                    self.assertEqual(
                        export.media_format(Path("fixture.part"), "audio"), ("audio/mpeg", ".mp3")
                    )

    def test_complete_bundle_deduplicates_bytes(self):
        import httpx
        from PIL import Image

        buffer = io.BytesIO()
        Image.new("RGB", (2, 2), "blue").save(buffer, format="PNG")
        body = buffer.getvalue()
        with tempfile.TemporaryDirectory() as folder:
            data = {
                "source": {"url": "https://study4.com/tests/6852/review/"},
                "content": [{"asset": "asset-1"}, {"asset": "asset-2"}],
                "assets": [
                    {"id": f"asset-{index}", "kind": "image", "url": f"https://media.example/{index}"}
                    for index in (1, 2)
                ],
            }
            args = Namespace(crawl_root=folder, output=str(Path(folder) / "copy.json"))
            original_client = httpx.Client
            transport = httpx.MockTransport(lambda request: httpx.Response(200, content=body))
            with patch(
                "httpx.Client", side_effect=lambda **kwargs: original_client(transport=transport, **kwargs)
            ):
                run = export.save_bundle(data, args)
                args.output = None
                second = export.save_bundle(data, args)
            self.assertNotEqual(run, second)
            manifest = json.loads((run / "media-manifest.json").read_text())
            self.assertEqual(manifest["status"], "complete")
            self.assertEqual(manifest["downloadedCount"], 2)
            root = (run / manifest["assetRoot"]).resolve()
            self.assertEqual(len(list((root / "media/images").glob("*.png"))), 1)
            packaged = json.loads((run / "export.json").read_text())
            self.assertEqual(packaged["content"], data["content"])
            self.assertEqual([a["id"] for a in packaged["assets"]], ["asset-1", "asset-2"])
            alias = json.loads(Path(folder, "copy.json").read_text())
            self.assertTrue((Path(folder) / alias["asset_root"] / alias["assets"][0]["localPath"]).is_file())
            self.assertTrue((Path(folder) / alias["media_manifest"]).is_file())

    def test_media_download(self):
        import httpx
        from PIL import Image

        image_buffer = io.BytesIO()
        Image.new("RGB", (2, 2), "red").save(image_buffer, format="PNG")
        audio_buffer = io.BytesIO()
        with wave.open(audio_buffer, "wb") as stream:
            stream.setnchannels(1)
            stream.setsampwidth(2)
            stream.setframerate(8000)
            stream.writeframes(b"\x00\x00" * 8000)
        cases = [
            ("image", image_buffer.getvalue(), 200, False),
            ("audio", audio_buffer.getvalue(), 200, False),
            ("image", b"<html>Login required</html>", 200, True),
            ("image", b"", 200, True),
            ("audio", b"Forbidden", 403, True),
        ]
        for kind, body, status, fails in cases:
            with self.subTest(kind=kind, status=status, fails=fails), tempfile.TemporaryDirectory() as folder:
                asset = {"id": "asset-1", "kind": kind, "url": "https://media.example/opaque"}
                transport = httpx.MockTransport(lambda request: httpx.Response(status, content=body))
                with httpx.Client(transport=transport) as client:
                    if fails:
                        with self.assertRaises(Exception):
                            export.download_asset(asset, Path(folder), client)
                    else:
                        first = export.download_asset(asset, Path(folder), client)
                        second = export.download_asset(asset, Path(folder), client)
                        digest = hashlib.sha256(body).hexdigest()
                        self.assertEqual(first["sha256"], digest)
                        self.assertEqual(first["minioKey"], f"exams/toeic/{digest}")
                        self.assertEqual((Path(folder) / first["localPath"]).read_bytes(), body)
                        self.assertEqual(second["localPath"], first["localPath"])
                        (Path(folder) / first["localPath"]).write_bytes(b"corrupt")
                        with self.assertRaisesRegex(ValueError, "SHA-256"):
                            export.download_asset(asset, Path(folder), client)
                self.assertEqual(list((Path(folder) / "media/.staging").glob("*.part")), [])

    def test_failed_bundle_retains_source(self):
        import httpx

        with tempfile.TemporaryDirectory() as folder:
            data = {
                "source": {"url": "https://study4.com/tests/6852/review/"},
                "content": [{"text": "Synthetic fixture"}],
                "assets": [{"id": "asset-1", "kind": "image", "url": "https://media.example/private"}],
            }
            args = Namespace(crawl_root=folder, output=None)
            original_client = httpx.Client
            transport = httpx.MockTransport(lambda request: httpx.Response(403))
            with patch(
                "httpx.Client", side_effect=lambda **kwargs: original_client(transport=transport, **kwargs)
            ):
                with self.assertRaisesRegex(RuntimeError, "incomplete"):
                    export.save_bundle(data, args)
            runs = list((Path(folder) / "study4/tests/6852/runs").iterdir())
            self.assertEqual(len(runs), 1)
            self.assertEqual(json.loads((runs[0] / "raw.json").read_text())["content"], data["content"])
            manifest = json.loads((runs[0] / "media-manifest.json").read_text())
            self.assertEqual(manifest["status"], "incomplete")
            self.assertEqual(manifest["failedCount"], 1)
            self.assertEqual(manifest["assets"][0]["status"], "failed")

    def test_follow_review_link(self):
        details_url = "https://study4.com/tests/6852/toeic-lr-collection-1-test-1/results/39251660/details/"
        cases = [
            ("not_results", "null", False, False),
            ("review", json.dumps(details_url), True, False),
            ("external", '"https://example.com/results/1/details/"', False, True),
            ("bad_response", "truncated", False, True),
        ]
        for name, payload, follows, fails in cases:
            with self.subTest(name=name):
                calls = []

                class Session:
                    async def call_tool(self, tool, arguments):
                        calls.append((tool, arguments))
                        return (
                            response("### Result\n" + payload)
                            if tool == "browser_evaluate"
                            else response("OK")
                        )

                if fails:
                    with self.assertRaises(RuntimeError):
                        asyncio.run(export.follow_result_details(Session()))
                else:
                    asyncio.run(export.follow_result_details(Session()))
                self.assertEqual(any(tool == "browser_tabs" for tool, _ in calls), follows)

    def test_url_prompt(self):
        url = "https://study4.com/tests/6852/practice/"
        cases = [
            ("valid", [url], url),
            ("trim", [f"  {url}  "], url),
            ("retry", ["", "https://example.com/", "https://study4.com.evil.example/", url], url),
            ("reject_credentials", ["https://user:pass@study4.com/", url], url),
            ("no_stdin", EOFError(), None),
        ]
        for name, entered, expected in cases:
            with self.subTest(name=name), patch("builtins.input", side_effect=entered):
                if expected is None:
                    with self.assertRaisesRegex(RuntimeError, "Pass --url"):
                        export.prompt_study4_url()
                else:
                    self.assertEqual(export.prompt_study4_url(), expected)

    def test_tab_selection(self):
        exam_url = "https://study4.com/tests/6852/toeic-lr-collection-1-test-1/"
        cases = [
            (
                "extension_only",
                "- 0: (current) [Welcome](chrome-extension://id/connect.html)",
                None,
                None,
                None,
            ),
            (
                "new_explicit_url",
                "- 0: [Welcome](chrome-extension://id/connect.html)",
                exam_url,
                None,
                {"action": "new", "url": exam_url},
            ),
            (
                "ambiguous",
                f"- 1: [Exam]({exam_url})\n- 2: [Other](https://study4.com/tests/999/)",
                None,
                None,
                None,
            ),
            (
                "explicit_index",
                f"- 1: [Exam]({exam_url})\n- 2: [Other](https://study4.com/tests/999/)",
                None,
                2,
                {"action": "select", "index": 2},
            ),
        ]
        for name, listing, url, index, expected in cases:
            with self.subTest(name=name):
                calls = []

                class Session:
                    async def call_tool(self, tool, arguments):
                        calls.append(arguments)
                        return response(listing)

                args = Namespace(url=url, tab_index=index)
                if expected is None:
                    with self.assertRaises(RuntimeError):
                        asyncio.run(export.select_study4_tab(Session(), args))
                else:
                    asyncio.run(export.select_study4_tab(Session(), args))
                    self.assertEqual(calls[-1], expected)

    def test_redaction(self):
        for value in (
            "?token=synthetic-secret)",
            "?x=1&token=synthetic-secret&y=2",
            "?access_token=synthetic-secret\n",
        ):
            with self.subTest(value=value):
                self.assertNotIn("synthetic-secret", export.redact_secrets(value))
                self.assertIn("[REDACTED]", export.redact_secrets(value))

    def test_error_details(self):
        cases = [
            (
                "nested",
                ExceptionGroup("outer", [ExceptionGroup("inner", [OSError("missing executable")])]),
                "OSError: missing executable",
            ),
            ("timeout", ExceptionGroup("outer", [TimeoutError()]), "Check the Playwright Extension"),
            ("ordinary", RuntimeError("Connection closed"), "RuntimeError: Connection closed"),
        ]
        for name, error, expected in cases:
            with self.subTest(name=name):
                self.assertIn(expected, export.describe_error(error))
        with self.assertRaisesRegex(RuntimeError, "Extension connection timed out"):
            export.check_tool_result(response("Extension connection timed out", True), "browser_evaluate")

    def test_decode_responses(self):
        data = {"schema_version": 1, "content": [{"text": "Synthetic fixture"}]}
        encoded = json.dumps(data)
        cases = [
            ("plain", response(encoded), data),
            ("playwright", response(f"### Result\n{encoded}\n### Ran Playwright code\nignored"), data),
            ("fenced", response(f"```json\n{encoded}\n```"), data),
            ("truncated", response(encoded[:-4]), None),
            ("tool_error", response(encoded, True), None),
            ("empty", response('{"schema_version":1,"content":[]}'), None),
        ]
        for name, result, expected in cases:
            with self.subTest(name=name):
                if expected is None:
                    with self.assertRaises(RuntimeError):
                        export.decode_export(result)
                else:
                    self.assertEqual(export.decode_export(result), expected)

    def test_mocked_mcp_flow(self):
        for url in (None, "https://study4.com/tests/6852/practice/"):
            with self.subTest(url=url), tempfile.TemporaryDirectory() as folder:
                calls = []

                class Session:
                    def __init__(self, *streams):
                        pass

                    async def __aenter__(self):
                        return self

                    async def __aexit__(self, *args):
                        pass

                    async def initialize(self):
                        calls.append(("initialize", None))

                    async def list_tools(self, cursor=None):
                        calls.append(("list_tools", cursor))
                        name = "browser_tabs" if cursor else "browser_evaluate"
                        return SimpleNamespace(
                            tools=[SimpleNamespace(name=name)], nextCursor=None if cursor else "page2"
                        )

                    async def call_tool(self, name, arguments):
                        calls.append((name, arguments))
                        if name == "browser_tabs":
                            return response(
                                "### Open tabs\n- 0: (current) [Welcome](chrome-extension://id/connect.html)\n"
                                "- 1: [Exam](https://study4.com/tests/6852/practice/)"
                            )
                        if "expectedPath" in arguments.get("function", ""):
                            return response("### Result\nnull")
                        return response(
                            "### Result\n"
                            + json.dumps(
                                {
                                    "schema_version": 1,
                                    "source": {"url": "https://study4.com/tests/6852/review/"},
                                    "page_kind": "landing",
                                    "content": [{"text": "Synthetic fixture"}],
                                    "assets": [],
                                }
                            )
                        )

                @asynccontextmanager
                async def transport(*args):
                    calls.append(("transport", args[0]))
                    yield (object(), object(), lambda: None)

                modules = {
                    name: ModuleType(name)
                    for name in ("mcp", "mcp.client", "mcp.client.stdio", "mcp.client.streamable_http")
                }
                modules["mcp"].ClientSession = Session
                modules["mcp"].StdioServerParameters = lambda **kw: kw
                modules["mcp.client.stdio"].stdio_client = transport
                modules["mcp.client.streamable_http"].streamablehttp_client = transport
                args = Namespace(
                    mcp_url=None,
                    command="fake-mcp",
                    browser_executable=None,
                    server_arg=["fixture"],
                    list_tools=False,
                    diagnose=False,
                    url=url,
                    tab_index=None,
                    selector="#synthetic",
                    output=str(Path(folder) / "export.json"),
                    crawl_root=folder,
                )
                with (
                    patch.object(export, "validate_exam_export"),
                    patch.dict("sys.modules", modules),
                    patch.dict("os.environ", {"PLAYWRIGHT_MCP_EXTENSION_TOKEN": "synthetic-test-token"}),
                ):
                    asyncio.run(export.export_study4_page(args))
                    parameters = next(value for name, value in calls if name == "transport")
                    self.assertEqual(
                        parameters["env"], {"PLAYWRIGHT_MCP_EXTENSION_TOKEN": "synthetic-test-token"}
                    )
                    saved = json.loads(Path(args.output).read_text(encoding="utf-8"))
                    self.assertEqual(saved["content"], [{"text": "Synthetic fixture"}])
                    self.assertIn("exported_at", saved)
                    self.assertIn(("list_tools", "page2"), calls)
                    evaluate = next(
                        arguments
                        for name, arguments in calls
                        if name == "browser_evaluate" and "const selector" in arguments["function"]
                    )
                    self.assertIn('const selector = "#synthetic" ||', evaluate["function"])
                    self.assertIn(("browser_tabs", {"action": "select", "index": 1}), calls)
                    with self.assertRaises(FileExistsError):
                        asyncio.run(export.export_study4_page(args))
                    args.output = None
                    args.url = "https://example.com/tests/"
                    with self.assertRaises(ValueError):
                        asyncio.run(export.export_study4_page(args))

    def test_resolve_target_urls(self):
        url1 = "https://study4.com/tests/1001/review/"
        url2 = "https://www.study4.com/tests/1002/details/"
        url3 = "https://study4.com/tests/1003/"

        # 1. Single string url
        args = Namespace(url=url1, urls=None, urls_file=None)
        self.assertEqual(export.resolve_target_urls(args), [url1])

        # 2. Repeated url list
        args = Namespace(url=[url1, url2], urls=None, urls_file=None)
        self.assertEqual(export.resolve_target_urls(args), [url1, url2])

        # 3. --urls flag
        args = Namespace(url=None, urls=[url1, url2], urls_file=None)
        self.assertEqual(export.resolve_target_urls(args), [url1, url2])

        # 4. --urls-file with comments and empty lines
        with tempfile.NamedTemporaryFile("w+", encoding="utf-8", delete=False) as tf:
            tf.write(f"# Comment line\n\n{url1}\n   {url2}   \n# Another comment\n{url3}\n")
            tf_path = Path(tf.name)
        try:
            args = Namespace(url=None, urls=None, urls_file=tf_path)
            self.assertEqual(export.resolve_target_urls(args), [url1, url2, url3])

            # 5. Combined and deduplicated preserving order
            args = Namespace(url=[url2], urls=[url1], urls_file=tf_path)
            self.assertEqual(export.resolve_target_urls(args), [url2, url1, url3])
        finally:
            tf_path.unlink(missing_ok=True)

        # 6. Reject invalid host or credentials
        bad_args = Namespace(url="https://evil.com/tests/1/", urls=None, urls_file=None)
        with self.assertRaises(ValueError):
            export.resolve_target_urls(bad_args)

        cred_args = Namespace(url="https://user:pass@study4.com/tests/1/", urls=None, urls_file=None)
        with self.assertRaises(ValueError):
            export.resolve_target_urls(cred_args)

    def test_batch_export_flow(self):
        with tempfile.TemporaryDirectory() as folder:
            calls = []
            url1 = "https://study4.com/tests/1001/practice/"
            url2 = "https://study4.com/tests/1002/practice/"

            class Session:
                def __init__(self, *streams):
                    pass

                async def __aenter__(self):
                    return self

                async def __aexit__(self, *args):
                    pass

                async def initialize(self):
                    pass

                async def list_tools(self, cursor=None):
                    name = "browser_tabs" if cursor else "browser_evaluate"
                    return SimpleNamespace(
                        tools=[SimpleNamespace(name=name)], nextCursor=None if cursor else "page2"
                    )

                async def call_tool(self, name, arguments):
                    calls.append((name, arguments))
                    if name == "browser_tabs":
                        action = arguments.get("action")
                        if action == "list":
                            return response(
                                "### Open tabs\n- 0: (current) [Welcome](chrome-extension://id/connect.html)\n"
                                f"- 1: [Test 1]({url1})\n- 2: [Test 2]({url2})"
                            )
                        return response("OK")
                    if "expectedPath" in arguments.get("function", ""):
                        return response("### Result\nnull")
                    return response(
                        "### Result\n"
                        + json.dumps(
                            {
                                "schema_version": 1,
                                "source": {
                                    "url": url1
                                    if len([c for c in calls if c[0] == "browser_evaluate"]) <= 2
                                    else url2
                                },
                                "page_kind": "exam",
                                "content": [{"text": "Batch question text"}],
                                "assets": [],
                            }
                        )
                    )

            @asynccontextmanager
            async def transport(*args):
                yield (object(), object(), lambda: None)

            modules = {
                name: ModuleType(name)
                for name in ("mcp", "mcp.client", "mcp.client.stdio", "mcp.client.streamable_http")
            }
            modules["mcp"].ClientSession = Session
            modules["mcp"].StdioServerParameters = lambda **kw: kw
            modules["mcp.client.stdio"].stdio_client = transport
            modules["mcp.client.streamable_http"].streamablehttp_client = transport

            batch_summary_output = str(Path(folder) / "summary.json")
            args = Namespace(
                mcp_url=None,
                command="fake-mcp",
                browser_executable=None,
                server_arg=["fixture"],
                list_tools=False,
                diagnose=False,
                url=[url1, url2],
                urls=None,
                urls_file=None,
                tab_index=None,
                selector=None,
                output=batch_summary_output,
                crawl_root=folder,
                timeout=60,
            )
            with (
                patch.object(export, "validate_exam_export"),
                patch.dict("sys.modules", modules),
                patch.dict("os.environ", {"PLAYWRIGHT_MCP_EXTENSION_TOKEN": "synthetic-test-token"}),
            ):
                asyncio.run(export.export_study4_page(args))

            summary = json.loads(Path(batch_summary_output).read_text(encoding="utf-8"))
            self.assertEqual(summary["total"], 2)
            self.assertEqual(summary["successful"], 2)
            self.assertEqual(summary["failed"], 0)
            self.assertEqual(len(summary["results"]), 2)
            close_calls = [c for c in calls if c[0] == "browser_tabs" and c[1].get("action") == "close"]
            self.assertGreaterEqual(len(close_calls), 2)

    def test_batch_retry_on_failure(self):
        with tempfile.TemporaryDirectory() as folder:
            url1 = "https://study4.com/tests/2001/practice/"
            url2 = "https://study4.com/tests/2002/practice/"
            eval_count = 0

            class Session:
                def __init__(self, *streams):
                    pass

                async def __aenter__(self):
                    return self

                async def __aexit__(self, *args):
                    pass

                async def initialize(self):
                    pass

                async def list_tools(self, cursor=None):
                    name = "browser_tabs" if cursor else "browser_evaluate"
                    return SimpleNamespace(
                        tools=[SimpleNamespace(name=name)], nextCursor=None if cursor else "page2"
                    )

                async def call_tool(self, name, arguments):
                    nonlocal eval_count
                    if name == "browser_tabs":
                        return response(
                            "### Open tabs\n- 0: (current) [Welcome](chrome-extension://id/connect.html)\n"
                            f"- 1: [Test 1]({url1})\n- 2: [Test 2]({url2})"
                        )
                    if "expectedPath" in arguments.get("function", ""):
                        return response("### Result\nnull")
                    eval_count += 1
                    if eval_count == 1:
                        raise RuntimeError("Transient network glitch")
                    return response(
                        "### Result\n"
                        + json.dumps(
                            {
                                "schema_version": 1,
                                "source": {"url": url1 if eval_count <= 2 else url2},
                                "page_kind": "exam",
                                "content": [{"text": "Retry success"}],
                                "assets": [],
                            }
                        )
                    )

            @asynccontextmanager
            async def transport(*args):
                yield (object(), object(), lambda: None)

            modules = {
                name: ModuleType(name)
                for name in ("mcp", "mcp.client", "mcp.client.stdio", "mcp.client.streamable_http")
            }
            modules["mcp"].ClientSession = Session
            modules["mcp"].StdioServerParameters = lambda **kw: kw
            modules["mcp.client.stdio"].stdio_client = transport
            modules["mcp.client.streamable_http"].streamablehttp_client = transport

            summary_file = str(Path(folder) / "summary_retry.json")
            args = Namespace(
                mcp_url=None,
                command="fake-mcp",
                browser_executable=None,
                server_arg=["fixture"],
                list_tools=False,
                diagnose=False,
                url=[url1, url2],
                urls=None,
                urls_file=None,
                tab_index=None,
                selector=None,
                output=summary_file,
                crawl_root=folder,
                timeout=60,
            )
            with (
                patch.object(export, "validate_exam_export"),
                patch.dict("sys.modules", modules),
                patch.dict("os.environ", {"PLAYWRIGHT_MCP_EXTENSION_TOKEN": "synthetic-test-token"}),
            ):
                asyncio.run(export.export_study4_page(args))

            summary = json.loads(Path(summary_file).read_text(encoding="utf-8"))
            self.assertEqual(summary["total"], 2)
            self.assertEqual(summary["successful"], 2)
            self.assertEqual(summary["failed"], 0)

    def test_minio_upload_flow(self):
        import httpx
        from PIL import Image

        buffer = io.BytesIO()
        Image.new("RGB", (2, 2), "green").save(buffer, format="PNG")
        body = buffer.getvalue()
        checksum = hashlib.sha256(body).hexdigest()

        with tempfile.TemporaryDirectory() as folder:
            data = {
                "source": {"url": "https://study4.com/tests/6852/review/"},
                "content": [
                    {
                        "tag": "div",
                        "class": "question-item-wrapper",
                        "children": [
                            {"tag": "div", "class": "question-number", "children": [{"text": "1"}]},
                            {"tag": "img", "asset": "asset-1"},
                        ],
                    }
                ],
                "assets": [{"id": "asset-1", "kind": "image", "url": "https://media.example/photo.png"}],
            }
            args = Namespace(
                crawl_root=folder,
                output=None,
                minio=True,
                keep_local_media=False,
                minio_endpoint="http://localhost:9000",
                minio_bucket="lesson-audio",
            )
            original_client = httpx.Client
            transport = httpx.MockTransport(lambda request: httpx.Response(200, content=body))

            fput_calls = []
            mock_minio = SimpleNamespace(
                bucket_exists=lambda b: True,
                fput_object=lambda bucket_name, object_name, file_path, content_type: fput_calls.append(
                    {"bucket": bucket_name, "object": object_name, "content_type": content_type}
                ),
            )
            with (
                patch(
                    "httpx.Client",
                    side_effect=lambda **kwargs: original_client(transport=transport, **kwargs),
                ),
                patch.object(
                    export, "get_minio_client_and_bucket", return_value=(mock_minio, "lesson-audio")
                ),
            ):
                run = export.save_bundle(data, args)

            expected_key = f"exams/toeic/{checksum}"
            self.assertEqual(len(fput_calls), 1)
            self.assertEqual(fput_calls[0]["bucket"], "lesson-audio")
            self.assertEqual(fput_calls[0]["object"], expected_key)
            self.assertEqual(fput_calls[0]["content_type"], "image/png")

            manifest = json.loads((run / "media-manifest.json").read_text())
            self.assertEqual(manifest["status"], "complete")
            self.assertEqual(manifest["assets"][0]["status"], "uploaded")
            self.assertEqual(manifest["assets"][0]["filename"], "q01.png")
            self.assertEqual(manifest["assets"][0]["minioKey"], expected_key)
            self.assertIsNone(manifest["assets"][0]["localPath"])

            # Verify no heavy image was retained locally in media/
            local_images = (
                list((Path(folder) / "study4/tests/6852/media/images").glob("*.png"))
                if (Path(folder) / "study4/tests/6852/media/images").exists()
                else []
            )
            self.assertEqual(len(local_images), 0)


if __name__ == "__main__":
    unittest.main()
