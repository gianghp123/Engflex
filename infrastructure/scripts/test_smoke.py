"""Provider-free checks for release gate failure behavior."""

import contextlib
import io
import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from smoke import ProbeFailure, Probes, run, settings


class ReleaseSmokeTests(unittest.TestCase):
    def test_configuration_rejects_missing_or_external_credentials_destination(self):
        base = dict.fromkeys(("SMOKE_LEARNER_TOKEN", "SMOKE_PREMIUM_TOKEN", "VOICE_INTERNAL_SECRET", "VNPAY_HASH_SECRET"), "test-secret")
        base.update(SMOKE_PUBLIC_CONTENT_PATH="/api/v1/public", SMOKE_PREMIUM_CONTENT_PATH="/api/v1/premium")
        for name, patch in (
            ("missing token", {"SMOKE_LEARNER_TOKEN": ""}),
            ("external origin", {"SMOKE_PUBLIC_CONTENT_PATH": "https://elsewhere.invalid/api/v1/public"}),
            ("protocol-relative", {"SMOKE_PREMIUM_CONTENT_PATH": "//elsewhere.invalid/api/v1/premium"}),
        ):
            with self.subTest(name=name), self.assertRaises(ProbeFailure):
                settings({**base, **patch})

    def test_release_gates_fail_closed(self):
        for name, code, payload, expected in (
            ("valid response", 200, {"data": {}}, None),
            ("provider outage", 503, {"message": "unavailable"}, "received 503"),
            ("HTML error page", 200, None, "invalid JSON"),
            ("redirect", 302, {}, "received 302"),
        ):
            with self.subTest(name=name):
                class Handler(BaseHTTPRequestHandler):
                    def do_GET(self):
                        self.send_response(code)
                        self.send_header("Location", "http://127.0.0.1:1/secret")
                        self.end_headers()
                        self.wfile.write(b"<html>error</html>" if payload is None else json.dumps(payload).encode())

                    def log_message(self, *args):
                        pass

                server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
                worker = threading.Thread(target=server.serve_forever, daemon=True)
                worker.start()
                try:
                    with contextlib.redirect_stdout(io.StringIO()):
                        probes = Probes(f"http://127.0.0.1:{server.server_port}")
                        if expected:
                            with self.assertRaisesRegex(ProbeFailure, expected):
                                probes.check("probe", "/", 200)
                        else:
                            self.assertEqual(probes.check("probe", "/", 200), payload)
                finally:
                    server.shutdown()
                    server.server_close()
                    worker.join()

    def test_payment_ack_cannot_silently_pass(self):
        class Stub:
            def __init__(self, code):
                self.code = code

            def check(self, *args, **kwargs):
                return {"RspCode": self.code}

        env = dict.fromkeys(("SMOKE_LEARNER_TOKEN", "SMOKE_PREMIUM_TOKEN", "VOICE_INTERNAL_SECRET", "VNPAY_HASH_SECRET", "SMOKE_PUBLIC_CONTENT_PATH", "SMOKE_PREMIUM_CONTENT_PATH"), "test")
        for code in ("00", "97", "99", None):
            with self.subTest(code=code), self.assertRaisesRegex(ProbeFailure, "acknowledgement"):
                run(Stub(code), env)


if __name__ == "__main__":
    unittest.main()
