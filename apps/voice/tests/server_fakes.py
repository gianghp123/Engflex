"""Test doubles for the dispatcher's WebRTC seam.

SmallWebRTCRequestHandler owns live aiortc state; these fakes pin its call
contract — handle_web_request(request, callback) -> answer dict,
handle_patch_request(request), close() — so the dispatcher's HTTP and state
behavior is testable without negotiating a peer connection.

The real handler invokes the callback only for NEW connections (a
renegotiation reuses the peer connection and never calls it); the fake
always invokes it, so tests must not assert on callback non-invocation for
renegotiation. `before_callback` lets a test expire a lease between the
route's check and the callback, the TTL-wins-the-race path.
"""


class FakeConnection:
    def __init__(self, pc_id: str = "pc-1"):
        self.pc_id = pc_id

        class _PC:
            connectionState = "new"

        self.pc = _PC()
        self.disconnected = False

    async def disconnect(self) -> None:
        self.disconnected = True


class FakeHandler:
    def __init__(self, answer: dict | None = None):
        self.answer = answer or {"sdp": "v=0 answer", "type": "answer", "pc_id": "pc-1"}
        self.before_callback = None
        self.web_requests: list = []
        self.patch_requests: list = []
        self.last_connection: FakeConnection | None = None
        self.closed = False

    async def handle_web_request(self, request, webrtc_connection_callback):
        self.web_requests.append(request)
        if self.before_callback is not None:
            self.before_callback()
        connection = FakeConnection()
        self.last_connection = connection
        await webrtc_connection_callback(connection)
        return self.answer

    async def handle_patch_request(self, request):
        self.patch_requests.append(request)

    async def close(self):
        self.closed = True
