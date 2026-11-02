"""Shared HTTP callback client. Only transient failures are retried."""
import asyncio
import random
import time

import aiohttp
from loguru import logger

from config import settings

INTERNAL_HEADER = "X-Internal-Secret"
_session: aiohttp.ClientSession | None = None

async def close_callbacks() -> None:
    global _session
    if _session is not None:
        await _session.close()
        _session = None

async def post_callback(conversation_id: str, operation: str, payload: dict) -> bool:
    global _session
    if _session is None or _session.closed:
        _session = aiohttp.ClientSession()
    url = f"{settings.api_url}/api/v1/internal/conversations/{conversation_id}/{operation}"
    deadline = time.monotonic() + 30
    for attempt in range(5):
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            break
        try:
            async with _session.post(
                url, json=payload, headers={INTERNAL_HEADER: settings.internal_secret},
                timeout=aiohttp.ClientTimeout(total=min(remaining, 10)),
            ) as response:
                await response.read()
                if 200 <= response.status < 300:
                    return True
                if response.status != 429 and response.status < 500:
                    logger.error("callback rejected", conversation_id=conversation_id,
                                 operation=operation, status=response.status)
                    return False
        except (aiohttp.ClientError, TimeoutError):
            pass
        if attempt < 4:
            delay = min(0.5 * 2**attempt + random.uniform(0, 0.25),
                        max(0, deadline-time.monotonic()))
            await asyncio.sleep(delay)
    logger.error("callback unacknowledged", conversation_id=conversation_id,
                 operation=operation, revision=payload.get("revision"))
    return False

async def finalize_session(conversation_id: str, duration_sec: int,
                           snapshot: dict) -> bool:
    return await post_callback(conversation_id, "finalize",
                               {**snapshot, "durationSec": duration_sec})
