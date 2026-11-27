# Modular Monolith: implementation and operating notes

## Boundaries

One deployable Go API and one PostgreSQL business database remain. `apps/api/internal/composition/root.go` constructs repositories, transaction manager, provider clients, controllers and policies once. Module registration only mounts routes. `apps/api/ownership.json` and `apps/api/scripts/check-module-boundaries.py` declare and check owner-specific consumer interfaces. No DI framework, broker, Kubernetes or service extraction was added.

| Owner | Data / behavior | Published read ports |
|---|---|---|
| Learning | lessons, activities, exams, videos, transcripts, attempts | scoped attempts and progress reads |
| Vocabulary | dictionary, decks, cards, user vocabulary state | dictionary/deck access policies |
| Learner | Clerk-backed directory adapter, profiles, bookmarks | profile reader, directory enrichment |
| Conversation | scenarios, personas, sessions, snapshots, feedback | scoped progress reader, scenario access |
| Billing | plans, feature registry, subscriptions, payments | common PremiumChecker, subscription administration |

Repositories return nested model graphs. Belongs-to relations use read-only GORM joins; has-many graphs use constant query counts inside repositories. Response mapping stays in controllers; services return models/read graphs. Clerk remains identity authority; no local users table exists. Directory enrichment attaches active subscriptions in one batch query.

## Contract changes (same API routes/envelope)

- Deck cards: `itemId`, `sequence`, `exampleSentence`, `note` and nested `item`. Flat phrase/meaning/video references were removed. Dictionary edits change dictionary content; card edits never change a shared definition.
- Dictionary: nullable enum columns `domain` and `partOfSpeech`; IPA in `details.ipa`. Explicit null clears either nullable enum; omission preserves it. Learner list and saved-list responses expose metadata; protected detail remains gated.
- Video progress: clients must send the acknowledged `revision`. Successful save increments it. Stale or closed writes return 409; clients refetch and explicitly accept the saved progress before resuming. Concurrent identical completion is idempotent.
- Admin directory: nested subscription summary with plan/status/expiry replaces flat premium enrichment. Admin attempt labels remain nested lesson/exam/video models.
- Internal Voice callbacks: `engineSessionId`, monotonic `revision`, full `turns`; finalize also contains `durationSec`. Not a public TypeScript contract.
- `/readyz`: DB ping with two-second timeout; 503 while draining. Liveness excludes external AI providers.

Regenerate, do not hand-edit: mockery, `make contracts`, `make swagger`, `pnpm compile:messages`. Swagger explicitly names duplicated package-local subscription schemas.

## Access and concurrency

Draft curated content is 404 to non-admins; published premium content checks active unexpired subscription for each action. Admin preview applies only to curated content. Personal content remains owner-only, including for admins. Deck usage checks both deck and referenced dictionary entry. Lists carry browse metadata instead of protected prompts/answers. Unsave/delete-owned/end-conversation remain available as cleanup.

Provider inference runs outside row-lock transactions. Lesson mutations then reacquire the attempt row, validate ownership/policy/state and merge the latest result. Completion requires every expected question/item/activity. Partial unique indexes protect open attempts. Video updates use optimistic revisions. Conversation creation serializes per user through a transaction advisory lock plus partial unique active-session constraint.

Voice snapshots lock the conversation row. Old revisions are ignored, equal revisions must have identical content, and corrections replace stale tails. Final snapshot, duration and terminal state commit atomically. Feedback records carry an analysis-input fingerprint; transcript revision must still match when analysis commits, and changed context invalidates affected feedback. Empty rows are retained for replay safety and excluded from history; absence of earlier callbacks is never evidence for deletion.

## Runtime / release

API timeouts: read-header 5s, read 30s, write 150s, idle 60s. Graceful API drain waits up to 150s before closing shared resources. DB defaults: 20 open / 10 idle, max lifetime 30m, max idle time 5m; configurable. Request body caps apply before parsing: JSON 1MiB, audio multipart 2MiB plus 256KiB overhead, internal snapshot 4MiB. Provider audio limits are also enforced in services.

Voice uses official Pipecat runner routes behind internal-secret middleware. Start idempotency is process-local and keyed by verified caller/conversation/input. Go always confirms start against the engine, including after engine restart. Voice runs one process/replica. Snapshot updates coalesce; one shared HTTP session retries network/429/5xx at most five times with jitter and a 30-second budget. Validation/auth errors are not retried. Oversize snapshots stop content intake and end the session explicitly. SIGTERM refuses new starts, drains sessions, flushes callbacks within a bounded window and logs unacknowledged work. Compose termination grace is 120s.

Deployment: validate probe configuration -> provision PostgreSQL/Redis/MinIO -> wait for DB -> Goose from one release runner -> roll out applications -> readiness -> authenticated smoke. `smoke.py` never initiates inference, checkout, or changes an existing conversation/payment. It uses a random missing conversation for callback checks and a signed random missing transaction for IPN acknowledgement. A failed probe fails the release; inspect the named check without exposing credentials.

## Schema rebuild and rollback

This workspace is pre-deploy. Original create migrations were rewritten and seeds converted together; this is not an upgrade chain for an already-deployed old schema. Use an explicitly selected disposable database and rebuild the full Goose chain. After rebuilding a learning dev database, run `make seed-audio` with the configured MinIO and TTS tools. Never run a reset against a retained environment just because a migration version already exists.

For a release with retained data: take and verify a backup before replacing the old schema, and rehearse restore with the matched API/web/Voice artifacts. Do not roll back API/web alone across the deck/profile/callback wire changes. Stop new sessions, allow drain, then restore the matching database snapshot and all three application versions. PostgreSQL down migrations delete tables and are not a data-preserving rollback mechanism. Pending subscription duration is snapshotted at checkout; plan edits do not change it. IPN commits in one transaction; repeated callbacks do not extend entitlement again. Admin revoke cancels all active subscriptions.

## Verification ledger — 2026-11-03

All implementation groups A–G have been delivered together with frontend consumers, generated contracts, fixtures and CI gates. No commit, infrastructure apply or real deployment was performed. Existing working-tree changes, including admin attempt browsing and content import tooling, were retained.

Verified locally:

| Gate | Result / evidence under `.tmp/` |
|---|---|
| PostgreSQL 16 fresh Goose up / full reset / up | Passed; `pg16-up.log`, `pg16-down.log`, `pg16-up2.log` |
| Full Go suite with actual PostgreSQL 16 and vet | Passed; `go-pg16-final.log`, `go-vet-nov3.log` |
| Linux race detector: full repository, TOEIC service and payment packages | Passed; `linux-race-full.log` (no test-name filter) |
| Voice pytest / Ruff | 125 passed / passed; `voice-nov3.log` |
| Web Biome, TypeScript, conventions, production build | Passed; `web-check-nov3.log`, `web-build-nov3.log`; 1059 referenced EN/VI message IDs |
| Contracts regeneration / standalone TypeScript, Swagger, ownership checks | Passed; 12 modules / 29 repository owners; runtime and callback DTOs excluded from public contracts |
| Main Playwright suite on PostgreSQL 16 | 22 passed; `playwright-pg16-final.log` |
| Admin Playwright suite on its dedicated database | 4 passed (two setup and two acceptance tests); `playwright-admin-final.log` |
| Docker API and Voice builds / boot / readiness / stop | Passed; API exited 0, Voice logged application shutdown complete (uv wrapper exit 143), neither OOM-killed |
| Actual Go ↔ official Pipecat runner acceptance | Passed; `runtime-acceptance4.log` |
| Release smoke regression tests | Three passed with local HTTP doubles |
| Terraform fmt / validate, Compose config, Bash syntax | Passed; no apply or deploy |

Database suites exercise actual constraints, enum/CHECK/schema parity, nullable joins, full completion invariants, concurrent starts and updates, concurrent IPN and revocation. Video detail coverage includes both present and absent category relations; the computed `video_count` is omitted from join columns. External black-box runtime tests prove in-flight HTTP requests finish during drain and DB failure changes readiness without breaking liveness.

The main browser suite covers HTTP publication/current-entitlement enforcement, draft/expired/revoked direct access, actual video revision conflicts, premium loss with preserved drafts, persisted analysis, engine-down recovery and signed-out EN/VI SSR. A completion-response-loss test commits completion on the backend, aborts only the browser RPC response, and proves retry succeeds without saving over the completed attempt or incrementing its revision.

The admin suite uses the explicitly authorized existing Clerk development admin. It verifies dictionary nullable fields/IPA, dictionary picker and deck-card CRUD, shared definition preservation, nested lesson/exam/video attempt labels, grant/revoke directory refresh, and checkout → signed test IPN → return UI → subscription refresh. The gateway is a local test double; no real payment occurs. No Clerk role or metadata was changed. Fixtures restore pre-existing subscription status and delete only their own rows. Normal Clerk test authentication is an external development dependency.

Container acceptance verifies concurrent idempotent official runner `/start`, authentication and mismatched-input rejection, Go re-confirmation after engine restart, stale-engine callback rejection, revision replay/conflict/reordering, transcript correction truncation, and atomic zero-duration finalization. Public `sessionId` remains the Go conversation ID; callbacks use the actual engine session ID. No WebRTC offer was submitted and no live Voice inference was initiated. Shared callback retry, active-session drain and failure behavior also have test-double coverage.

The API and Voice validation containers were stopped gracefully. Task-labelled validation containers/network and temporary credential/identity copies were removed after evidence collection. Existing dev PostgreSQL/Redis/MinIO containers remain available. Container builds exclude `.tmp`, graph indexes and browser artifacts; Voice boot uses `uv run --no-sync` and does not fetch development dependencies.

## Local dev rebuild and lesson audio

The original local dev database at loopback port 5777, database `engflex`, used the old schema. Before rebuilding it, a PostgreSQL 16 custom-format backup was restored into an isolated verification database and row counts matched for all 29 original tables. The SHA256 and row-count manifest are retained alongside the backup:

- `.tmp/dev-before-modular-20261103.dump`
- `.tmp/dev-backup-manifest.json`

The disposable dev schema was then rebuilt using the edited Goose chain and fresh seeds. Original personal dev rows are retained in the backup, rather than in the freshly seeded database. `dev-migration-up.log` records the rebuild. Checksum, exact local target and absence of other DB sessions were checked before rebuilding.

With explicit user approval, lesson seed text was sent to Microsoft Edge TTS and 96 lesson audio objects were uploaded to local MinIO using the existing seed-audio command. All 96 referenced objects were subsequently checked for nonzero size and `audio/mpeg` content type. See `seed-audio-dev.log`. No bucket policy was changed.

The configured CI workflow includes PostgreSQL 16 integration, Linux race and generated-contract drift gates; remote CI execution is pending a future push (this task leaves all work uncommitted). Live LLM/STT/Voice inference and a real VNPay transaction were not run; default tests use doubles. Edge TTS seed generation above is the explicitly authorized external provider operation. No production environment was touched.

Accepted operational limits remain: media is public, and a killed/crashed Voice process can lose data not yet acknowledged by Go. Voice stays single-process/single-replica; this implementation does not promise multi-instance session routing or crash-durable delivery.

## TOEIC catalog recovery — 2026-11-03

The dev rebuild initially restored only migration seeds, leaving imported TOEIC categories/tests empty. After the user reported the empty exams hub, the original backup checksum was rechecked and only `exam_categories` (7 rows) and `exam_tests` (3 rows) were restored atomically, guarded by empty-table checks and table locks. Counts of attempts, lesson activities and vocabulary items were unchanged. The actual production catalog repositories/services then listed all 7 categories and successfully parsed every restored test through both ListTests and GetTest. Existing import assets in MinIO were not modified. Imported content must be explicitly preserved/restored when rebuilding disposable schemas; migration seeds alone do not include these TOEIC tests.

## TOEIC media recovery — 2026-11-03

The follow-up screenshot exposed an incomplete recovery: restoring database rows and parsing content had not verified actual media. A full audit found all 17 ETS TOEIC 2024 Test 02 objects absent from local MinIO; both Study4 tests already had all 12 objects each. The existing publish_toeic_media.py tool validated original local MP3/crop SHA256, image decoding and audio duration/ranges, then published and read-back-verified all 17 objects (53,442,659 bytes). No question, answer, transcript or attempt rows were changed.

All 41 manifest assets were subsequently fetched through the running frontend `http://localhost:3000/api/media/exams/toeic/<sha256>` and matched SHA256 (76,508,203 bytes). An explicit audio byte-range probe returned 206 and the correct 1024-byte range. Chrome headless decoded the first Part 1 image (657×948), loaded audio metadata (2752.520756 seconds), sought to 99.2 seconds and played forward to 99.76004 seconds. This is actual browser media decoding/playback, not merely a DB/parser check. The previously open React error state requires a page reload to recreate failed audio/image elements.
