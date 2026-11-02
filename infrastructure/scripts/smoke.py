"""Release probes. No inference, checkout, or existing-record mutations.

Use short-lived test learner tokens. Never pass credentials on the command line.
SMOKE_API_URL is the API origin; content paths include /api/v1.
"""

import argparse
import hashlib
import hmac
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid


class ProbeFailure(Exception):
    pass


class NoRedirects(urllib.request.HTTPRedirectHandler):
    # A redirect must not forward Authorization/internal secrets to another host.
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


REQUIRED = (
    "SMOKE_LEARNER_TOKEN",
    "SMOKE_PREMIUM_TOKEN",
    "SMOKE_PUBLIC_CONTENT_PATH",
    "SMOKE_PREMIUM_CONTENT_PATH",
    "VOICE_INTERNAL_SECRET",
    "VNPAY_HASH_SECRET",
)


def settings(env):
    for name in REQUIRED:
        if not env.get(name):
            raise ProbeFailure(f"missing release probe setting: {name}")
    for name in ("SMOKE_PUBLIC_CONTENT_PATH", "SMOKE_PREMIUM_CONTENT_PATH"):
        path = urllib.parse.urlsplit(env[name])
        if path.scheme or path.netloc or not path.path.startswith("/api/v1/"):
            raise ProbeFailure(f"{name} must be an API-relative path")
    return env


class Probes:
    def __init__(self, origin):
        parsed = urllib.parse.urlsplit(origin)
        if parsed.scheme not in ("http", "https") or not parsed.netloc or parsed.path not in ("", "/"):
            raise ProbeFailure("SMOKE_API_URL must be an HTTP(S) origin")
        self.origin = origin.rstrip("/")
        self.client = urllib.request.build_opener(NoRedirects())

    def check(self, label, path, status, headers=None, body=None):
        payload = None if body is None else json.dumps(body).encode()
        request = urllib.request.Request(
            self.origin + path,
            data=payload,
            headers={"Content-Type": "application/json", **(headers or {})},
        )
        try:
            response = self.client.open(request, timeout=10)
        except urllib.error.HTTPError as error:
            response = error
        except (OSError, urllib.error.URLError):
            raise ProbeFailure(f"{label}: connection failed") from None
        with response:
            if response.code != status:
                raise ProbeFailure(f"{label}: expected HTTP {status}, received {response.code}")
            try:
                result = json.loads(response.read(1 << 20))
            except (ValueError, OSError):
                raise ProbeFailure(f"{label}: invalid JSON response") from None
        print(f"PASS {label}")
        return result


def run(probes, env):
    learner = {"Authorization": "Bearer " + env["SMOKE_LEARNER_TOKEN"]}
    premium = {"Authorization": "Bearer " + env["SMOKE_PREMIUM_TOKEN"]}
    probes.check("readiness", "/readyz", 200)
    probes.check("authentication required", "/api/v1/profiles/me", 401)
    probes.check("authenticated learner", "/api/v1/profiles/me", 200, learner)
    probes.check("public content", env["SMOKE_PUBLIC_CONTENT_PATH"], 200, learner)
    probes.check("premium denied", env["SMOKE_PREMIUM_CONTENT_PATH"], 403, learner)
    probes.check("premium granted", env["SMOKE_PREMIUM_CONTENT_PATH"], 200, premium)
    # Random missing identity: checks auth, validation and repository path without
    # modifying a real conversation, even when pointed at a release environment.
    path = f"/api/v1/internal/conversations/{uuid.uuid4()}/finalize"
    snapshot = {"engineSessionId": str(uuid.uuid4()), "revision": 1, "turns": [], "durationSec": 0}
    probes.check("callback rejects unauthenticated caller", path, 401, body=snapshot)
    probes.check("callback verifies missing conversation", path, 404,
                 {"X-Internal-Secret": env["VOICE_INTERNAL_SECRET"]}, snapshot)
    params = {"vnp_TxnRef": "smoke-" + uuid.uuid4().hex, "vnp_Amount": "10000", "vnp_ResponseCode": "00"}
    encoded = urllib.parse.urlencode(sorted(params.items()))
    signature = hmac.new(env["VNPAY_HASH_SECRET"].encode(), encoded.encode(), hashlib.sha512).hexdigest()
    result = probes.check("VNPay signed unknown order", "/api/v1/payments/vnpay/ipn?" + encoded + "&vnp_SecureHash=" + signature, 200)
    if result.get("RspCode") != "01":
        raise ProbeFailure("VNPay: expected unknown-order acknowledgement 01")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--validate-config", action="store_true")
    args = parser.parse_args()
    try:
        env = settings(os.environ)
        if not args.validate_config:
            run(Probes(env.get("SMOKE_API_URL", "")), env)
    except ProbeFailure as error:
        print(f"FAIL {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
