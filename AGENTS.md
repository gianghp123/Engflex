# engflex-ultimate — Contributor Guide

## NOTE
- always ask or read real code if you have any confusion, DO NOT GUESSING.

> Stack: Go + Gin + GORM (control plane), PostgreSQL, Clerk auth.
> AI Voice (Python + Pipecat) is a separate stateless service, **uv-managed**
> (`pyproject.toml` + `uv.lock` + `.python-version`; always `uv run`) — Go never
> proxies audio.
> Frontend: Vite + TS + TanStack Start/Router, TanStack Store/Query, Clerk, shadcn/ui, Paraglide i18n.
> Run Go commands from `<api-dir>` (e.g. `apps/api`).

Monorepo (optional): `<api-dir>` (Go), `<web-dir>` (frontend),
`<contracts-dir>` (generated TypeScript shared package), `<voice-dir>` (Python +
Pipecat, uv-managed, independent).

## Standing preferences (human decisions — do not override without asking)

- **References over flat regrouping — HIGH NOTE.** See dedicated subsection
  below: nest relations instead of flattening, assemble in the repo, one Map.
- **No commits.** Leave all work uncommitted in the working tree unless the human
  explicitly asks for a commit. `docs/` is gitignored by intent — specs, plans
  and ledgers are never committed.
- **Use the library, don't reinvent it.** Prefer library-provided integrations
  over custom implementations; verify claims against the installed code
  (`node_modules`, generated output), not training data or pasted snippets.
- **Web i18n (Paraglide + inlang message-format):** locales `vi` (base/default)
  + `en` only.
  - `messages/{locale}/<domain>.json`, each file nested under its domain key
    (the plugin merges all files into one flat id namespace — unprefixed
    duplicates silently override each other).
  - Ids nest **by page or part**, with no depth limit: `lessons.hub.title`,
    `lessons.dictation.placeholder`, `voice.room.feedback.span.incorrect`. One
    file per domain, but never one flat namespace inside it — every screen's keys
    share a prefix. A file whose level-2 keys are all leaves is a smell.
  - `common.*` is reserved for copy consumed by shared components
    (`components/common`, `components/ui`, `components/layout`) and
    cross-cutting chrome, itself grouped by concern (`brand`, `actions`, `audio`,
    `a11y`, `theme`, `toast`). Copy used by exactly one page or feature moves to
    that feature's domain.
  - Read copy only as `m["<static string literal>"]()`. Never build the key with
    a template literal — Paraglide's generated types are what catch a rename, and
    an interpolated key bypasses them.
  - Strategy `["cookie", "baseLocale"]` — never `url` (path prefixes 404 in
    TanStack Router and are client-only, causing SSR mismatch). SSR locale via
    the official `paraglideMiddleware` in `src/start.ts` `requestMiddleware`;
    never import `node:*` in hand-written `src` (it leaks into the client
    bundle — the lib lazy-imports it internally).
  - `pnpm compile:messages` before typecheck; `src/paraglide/` is
    generated-only, never hand-edit.
  - Scope: UI chrome localized; fixture/learning content (lesson titles, terms,
    transcripts, passages) stays English. EN copy verbatim from mocks; VI authored.
  - Locale-sensitive code must resolve at render: breadcrumb `staticData`
    labels and `PART_META` labels are thunks; fixture-held chrome keyed by item id.
- **Web routes:** directory convention (`_app/route.tsx`,
  `_app/lessons/index.tsx`), not dotted flat names.
- **Web URL maps nest by resource, never flat.** `APP_ROUTES` and `API_ROUTES`
  mirror each other by design: `APP_ROUTES.LESSONS.{LIST,DETAIL,PART}`,
  `API_ROUTES.CONVERSATIONS.{LIST,BY_ID,START}`. Leaf keys are `LIST` for the
  index route, `DETAIL` for the parametrised one, then named actions. The two
  differ deliberately: `APP_ROUTES` keeps `$param` templates as **strings**
  (TanStack infers the `params` object from the literal, so a function would
  destroy that checking), `API_ROUTES` uses `(id) => …` functions. Zero URL
  literals outside those two files — including `href`, and including absolute
  URLs.
- **Web gates (no test framework):** `pnpm check`, `pnpm exec tsc --noEmit`,
  `pnpm check:conventions`, `pnpm build`, plus HTTP/SSR probes. Changes to
  `src/start.ts` / `vite.config.ts` require a dev-server restart (HMR cannot
  apply them). `pnpm check:conventions` is the only gate that catches an
  orphaned message id, an en/vi key-set divergence, or a hardcoded URL — run it.

## Folder structure (`<api-dir>`)

```text
cmd/server/main.go          # entrypoint; wires config + router only, no business logic
config/                     # auth, config, cors, database, log, server
docs/                       # generated swagger (swag); served at /swagger
internal/
  common/                   # AppError, ApiResponse envelope, shared constants/limits
  database/
    migrations/             # goose SQL — source of truth for schema (never AutoMigrate)
    models/                 # gorm structs + TableName() + scopes (no gorm.Model)
    repositories/           # hand-written repos over *gorm.DB + mocks/
  logger/                   # slog setup (New/NewWithWriter/Report/WithRequestID/WithUserID)
  modules/<domain>/         # contents, exercises, attempts, vocabulary, personas, conversations, feedback, profiles
    module.go               # RegisterRoutes(rg *gin.RouterGroup, deps...) — wiring only
    controllers/            # gin handlers (HTTP), swagger annotations, no SQL
    services/               # business logic (returns *Model / []*Model + error, no gin)
    dtos/requests|responses # wire DTOs (tygo inputs, gin binding tags)
  server/middleware/        # RequireAuth, RequestID, RequestLogger, Recovery
  utils/                    # Map/MapSlice, ParsePagination/Offset, IDs, Slug
```

No `live/` or `transports/` here: realtime audio goes FE ↔ Python directly over
WebRTC. Go only provisions AI sessions (persona/context/token) and ingests
results (transcript/feedback) afterwards.

## Models, Router, Service, Repo — GORM only

### Model (GORM)

```go
type UserProfile struct {
  ClerkUserID string `gorm:"primaryKey;type:text"` // Clerk subject, no users table
  Level       string
  Goals       []string `gorm:"type:text[]"`
}
func (UserProfile) TableName() string { return "user_profiles" }
```

- Explicit IDs/timestamps, no `gorm.Model`.
- `clerk_user_id text PK` on profile; all `user_id` FKs are Clerk strings.
- JSON columns (`config`, `score`) as `datatypes.JSON` / custom types.
- Arrays (`goals`, `skill_tags`, `topic`) as `pq.StringArray` / `text[]`.
- **A closed set is a Go `enums.*` type + a column with a `CHECK`, never free
  text inside a jsonb blob.** Display-only *free text* (IPA, transcription) may
  live in `details`; anything the product enumerates (part of speech, source,
  status) gets an enum type, a nullable column, and a `CHECK` listing the
  members — so the DB rejects an unknown member instead of storing it.
- `time.Time` → RFC3339 string in contracts.

#### Relations: `->` for anything read-only

A relation tag makes the field **writable**, and GORM will auto-save it. This is
not a warning, it is what the tag asks for: `UpsertTurns` does
`Create(&turns)`, so a turn carrying a populated `Feedback` silently INSERTs a
duplicate `feedbacks` row — `err == nil`, nothing logged. Any read-then-write
round-trip of a model does this. So pick the permission deliberately:

| Tag | Joins | Auto-saves on write | Use for |
| --- | --- | --- | --- |
| `gorm:"foreignKey:X;references:Y"` | yes | **yes** | relations you genuinely own and write |
| `gorm:"->;foreignKey:X;references:Y"` | yes | no | **read-mostly relations loaded by a join** |
| `gorm:"-"` | **no** | no | never use on a field a query fills |
| `gorm:"-:migration"` | yes | **yes** | never use to disable writes |

`->` keeps `Creatable`/`Updatable` false and `Readable` true, so the field stays
in `relationshipFields` (`schema.go:301` gates on
`DataType == "" && GORMDataType == "" && (Creatable || Updatable || Readable)`),
`Joins("Feedback")` still resolves, and `SelectAndOmitColumns` marks the relation
excluded so `SaveBeforeAssociations` skips it. `-` clears all three permissions
**and** `DataType`: the relation unregisters, the LEFT JOIN drops out of the SQL,
and the query returns zero rows with no error. `-:migration` sets only
`IgnoreMigration` — CRUD is untouched.

Relations with no foreign key (the polymorphic `feedbacks.subject_id` pair) are
declared in Go only; no migration. `models.ConversationTurn.Feedback` is the
worked example, and `models/tests/turn_feedback_relation_test.go` pins all three
properties from the parsed schema so the tag cannot be "simplified" back.

### Repository (DB access only, no HTTP, no business rules)

Returns pointers: single `*Model`, list `[]*Model`. Never return values.

```go
type AttemptRepository interface {
  Create(ctx context.Context, m *models.Attempt) error
  GetByID(ctx context.Context, id string) (*models.Attempt, error)
  ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Attempt, error)
  CountByUser(ctx context.Context, userID string) (int64, error)
  WithTx(tx *gorm.DB) AttemptRepository
}
func NewAttemptRepository(db *gorm.DB) AttemptRepository
```

- One method = one query. Aggregates (e.g. Progress over `attempts`) live in repos
  via query builder / `db.Raw()`, never string-built in services.
- No joins in services; add a repo method instead.
- **One repository per table, named for the model it returns.** The repo that
  answers with turns is `ConversationTurnRepository`, even if its SQL joins
  feedbacks; the scenario preview list is `ScenarioTopicRepository`, even
  though its second query reads scenarios. The file is named for what comes
  back, not for every table the SQL happens to touch.
- **No custom struct in a repository.** A repo returns models — `*models.X`,
  `[]*models.X` — never a query-only row shape. See "Never hand-build a query
  result as a custom struct" below.
- No in-memory join in a controller or service. If the page needs two tables,
  that is one repo method with a `Joins`, not two calls plus a map-by-id loop.
- All repo interfaces get mockery mocks in `repositories/mocks/`.

#### Reading relations: `Joins`, and the four ways it bites

`Joins("Relation")` is the read path — one round trip, and GORM fills the
struct from the `Relation__*` column aliases it generates. `GetDetail`
(`Joins("Topic").Joins("Persona")`) is the pattern. `Preload` is the other
option and buys nothing here: it is a second query.

1. **`Joins("Relation", conds...)` silently drops the conds.** They never reach
   the generated SQL. Put relation filters in `Where` instead.
2. **A `Where` on the joined table must use the join ALIAS, quoted.**
   `Joins("Feedback")` aliases the table as `"Feedback"`, and that alias shadows
   the bare name — `Where("feedbacks.subject_type = ?")` fails with SQLSTATE
   42P01 `invalid reference to FROM-clause entry`. Write
   `"Feedback".subject_type`.
3. **Filters must be NULL-tolerant on a LEFT JOIN.** A bare `= ?` deletes every
   row that has no match, because `NULL = 'x'` is never true. Use
   `"Feedback".subject_type IS NULL OR "Feedback".subject_type = ?`, or the
   missing rows vanish from the result entirely.
4. **Column types must match across the join.** GORM emits
   `ON parent.pk = child.fk` as a bare `clause.Eq{Column, Column}` — no cast, no
   hook to add one. `uuid = text` is a runtime SQLSTATE 42883, so a text-typed
   polymorphic key cannot be joined to a uuid parent. Fix the column type; do not
   reach for a cast.

A NULL join row yields a **nil** pointer, never a zero-valued struct — that is
what makes "no coaching yet" expressible as `turn.Feedback == nil`.

Two GORM behaviours worth knowing before you design a query around them: a
has-many `Joins` is forbidden (use two queries, or group in Go from two
constant queries), and `Preload` cannot limit rows per parent. The join
builder does not discriminate by relation type, so a has-many join executes
as a cartesian product — one row per parent×child — silently breaking
`Limit`/`Offset` pagination and row counts. (The old "panics in
`scanIntoStruct`" wording was version-specific; verified absent on gorm
v1.31.0. The duplication is the version-independent reason.)

#### Never hand-build a query result as a custom struct

The anti-pattern this codebase has already paid for twice. Both times the repo
declared a bespoke row struct, `db.Raw`/`db.Joins`d into it, and the service or
controller reassembled the model by hand:

```go
// NEVER: a repo-only shape, then Go code to rebuild the model from it
type topicPreviewRow struct { ... }              // repo struct, not a model
var rows []topicPreviewRow
... then in the controller: build []*ScenarioTopic by hand, one loop
```

```go
// NEVER: two queries plus an in-memory join in a controller
turns, _ := repo.ListTurns(ctx, id)
feedbacks, _ := repo.ListForSubjects(ctx, userID, subjectType, ids)
dto.Turns = attachFeedback(turns, feedbacks)     // maps by subject id in a loop
```

Both discard the model's type for a shape that only that query needs, and both
push relationship logic into the layer that must not have it. Do this instead:

- Return **`[]*models.X`** from the repository, with relations attached, and let
  `utils.MapSlice` copy the whole graph in one call.
- When a relation needs no foreign key, declare it in Go on the model
  (`gorm:"->;foreignKey:SubjectID;references:ID"`) and let one `Joins` fill it.
- When a query genuinely cannot be expressed as a model (a window function, a
  CTE), it is the **only** raw SQL in the repo, it returns a model pointer, and
  the result is grouped in Go from constant queries — never a custom struct, and
  never a per-row hand-built model. `rankedIDs` in `scenario_repository.go` is
  the worked example.

One round trip beats two, but not at the price of the model's identity.

#### Verify the SQL against a real database, not the string

Asserting the emitted SQL string with sqlmock is necessary and **not
sufficient** — the string was byte-for-byte what the plan expected while the
query failed at runtime three separate ways:

| Failure | sqlmock said | Postgres said |
| --- | --- | --- |
| `uuid = text` (42883) | correct | error |
| `"Feedback"` alias unresolvable (42P01) | guard present | error |
| subject guard deleted the feedback-less row | guard present | 0 turns |

The first two are invisible in the string; the third is invisible until you look
at row counts. So: check the SQL string *and* run the query against the real
database and read the rows. Use a uuid for any id column in a fixture — a
placeholder like `"c1"` fails validation before reaching SQL, which hides
everything.

### Service (business logic, returns *Model / []*Model + error)

```go
func (s *AttemptService) CreateAttempt(ctx context.Context, userID string, req requests.CreateAttempt) (*models.Attempt, error)
func (s *AttemptService) ListAttempts(ctx context.Context, userID string, limit, offset int) ([]*models.Attempt, int64, error)
```

- Constructors take interfaces, never concretes:
  `NewAttemptService(attempts AttemptRepository, contents ContentRepository)`.
- Map DB errors at the boundary: `common.FromDBError(err, "<resource>")`,
  log with `logger.Report(ctx, "op failed", appErr, "k", v)`.
  Success path: `slog.InfoContext`.
- No `gin.Context`, no `c.JSON`, no SQL strings here.
- Domain rules live here: difficulty heuristic on content create, completion-status
  derivation, AI session provisioning (persona/context/token), feedback async rules.
- Mapping via `utils.Map`/`MapSlice` (same rule as controllers — one call per
  mapping, no loops, no manual field lines). Services map `request→model` and
  `model→voice/internal`, never `model→response` (that stays in controllers):
  `Create` maps `StartConversation→Conversation` then sets `UserID`/`Status`;
  `IngestTurns` maps `[]IngestTurn→[]*ConversationTurn` then injects
  `ConversationID`; `promptInputs` maps `Scenario→VoiceScenarioBody` and
  `Persona→VoicePersonaBody`; `priorTurns` maps `[]*ConversationTurn→[]ContextTurn`.
  copier matches Go field names (never json tags), handles `*string→string`
  (nil→`""`, so no `deref` helper) and named-string→`string` (verified).
  When the function returns `error`, check the Map error and return
  `common.Internal()`; helpers without an error return (`promptInputs`,
  `priorTurns`) use `_ =` like controllers.

### Controller + Router (HTTP in/out, no SQL)

`modules/<domain>/module.go`:
```go
func RegisterRoutes(rg *gin.RouterGroup, repo repositories.AttemptRepository, ...) {
  svc := services.NewAttemptService(repo)
  ctl := controllers.NewAttemptController(svc)
  g := rg.Group("/attempts")
  g.POST("", middleware.RequireAuth(), ctl.Create)
  g.GET("", middleware.RequireAuth(), ctl.List)
  g.GET("/:id", middleware.RequireAuth(), ctl.GetByID)
}
```
`internal/server/router.go: func NewRouter(db *gorm.DB, ...) *gin.Engine` mounts
all `RegisterRoutes` under `/api/v1` + global `RequestID, RequestLogger, Recovery, CORS`.

Controller pattern:
```go
func (h *Ctl) Create(c *gin.Context) {
  var req requests.CreateAttempt
  if err := c.ShouldBindJSON(&req); err != nil { common.Fail(c, common.BadRequest(...)); return }
  userID := middleware.UserID(c) // Clerk subject, used directly as FK
  m, err := h.service.CreateAttempt(c.Request.Context(), userID, req)
  if err != nil { common.Fail(c, err); return }
  var dto responses.Attempt
  _ = utils.Map(&dto, m)
  common.Created(c, "created", dto)
}
```
- Input: `ShouldBindJSON`/`ShouldBindQuery`/`Param` → `common.BadRequest` on fail.
- Pagination: `utils.ParsePagination(c)` → `Offset()` → `List+Count` → `MapSlice`.
- Keep swagger annotations on handlers.

### Input (DTOs + validation)

- `dtos/requests/*.go`: gin `binding:"required,..."` tags.
  Limits MUST equal `internal/common` constants, guarded by a `limits_test.go`.
- `dtos/responses/*.go`: wire DTOs only (tygo inputs). Map via `utils.Map/MapSlice`.
- `time.Time` → RFC3339 string in contracts.

### DTO shape and mapping (house preference)

- DTOs mirror the model structure, nested the same way; fields a view
  doesn't need are removed, never flattened. Same names on both sides
  (`ScenarioDetail`, not `ScenarioDetails`) so the mirror is grep-visible.
- **The DTO adapts to the model and repository function, never the reverse.** Do
  not reshape a model or bend a repository signature to fit a DTO.
- One `utils.Map`/`MapSlice` call per mapping, in controllers **and** services —
  no loops, no manual field lines, no per-call options. DeepCopy is the util
  default and descends into nested structs (proven), so a same-shaped DTO maps
  whole. Service examples: `IngestTurns` (`MapSlice` + `ConversationID` inject),
  `priorTurns` (`MapSlice`), `promptInputs` (`Map` twice, no `deref`).
- copier matches Go field names, never json tags: names must match exactly
  across model and DTO; tags own the wire names independently
  (`MaxDuration int \`json:"maxDuration"\``).
- Cross-package types in DTOs break tygo (per-package resolution): wire
  copies stay in the DTO package with identical field names.

#### Never flatten a relation into its parent DTO

A flattened field cannot be filled by copier, and copier **fails silently** —
verified: mapping `*models.Feedback` into a flat `*responses.TurnFeedback`
returns `err == nil` and yields `Corrected == ""` with zero spans. Green tests,
empty panel, nothing in the logs.

So nest, and drop only the columns the view does not render:

```go
type Turn struct { ... Feedback *Feedback `json:"feedback,omitempty"` }
type Feedback struct { Payload TurnFeedback `json:"payload"` } // no UserID/SubjectType/SubjectID
```

The identity columns are relationship facts the join already proved; they do not
cross the wire. `turn.feedback.tip` (flattened) is the shape to avoid.

#### References over flat regrouping (HIGH NOTE)

When a view needs a parent with children, carry the **reference** through
every layer — never a flat list regrouped by hand:

- Model carries the relation (`Units []*Lesson` with `->`, read-only).
- Repository fills it (belongs-to via `Joins`; has-many via a second query
  grouped in Go from constant queries — never a has-many `Joins`).
- DTO nests the same way (`UnitSection{ Section, Units }`).
- Controller does one `utils.Map`/`MapSlice`, no loops.

The two shapes this forbids are the same mistake from opposite ends:
flattening child fields onto the parent DTO (copier fails silently — see
above), and returning a flat child list plus a grouping loop in a
controller or service (an in-memory join by another name — same cost as the
`attachFeedback` loop this codebase already paid for). Worked example: the
sections hub reads section-first (`LessonSection` with `Units` attached),
not flat lessons regrouped by a `groupSections` helper.

### Response format (fixed envelope)

```go
// internal/common/response.go
type ApiResponse struct { Message string `json:"message"`; Data any `json:"data,omitempty"`; Pagination *Pagination `json:"pagination,omitempty"` }
func OK(c, message, data)      // 200
func Created(c, message, data) // 201
func Paginated(c, message, data, page, pageSize, total)
func Fail(c, err) // *AppError -> {message}+Status, else hidden 500
```

Errors `internal/common/errors.go` (GORM mapping):
`BadRequest/Unauthorized/Forbidden/NotFound/Conflict/UnprocessableEntity/TooManyRequests/Internal/ServiceUnavailable`
as `*AppError`.

| DB outcome | GORM | HTTP |
| --- | --- | --- |
| bad ID | `InvalidIDError` / parse fail | 400 |
| no rows | `gorm.ErrRecordNotFound` | 404 `<resource> not found` |
| unique | duplicate key / `23505` | 409 |
| FK | `23503` / constraint | 404/422 |
| else | — | 500 hidden |

Assert with `require.ErrorAs(err, &appErr)` + check `.Status`.

## Code style (Go)

- `gofmt` clean, `go vet` clean. No `nil` into constructors.
- Status codes via `net/http` consts, never hardcoded ints.
- `slog` via `logger` package only (see Logging convention).

## Logging convention

- Setup once in `main.go`: `slog.SetDefault(logger.New(cfg.Log.Level, cfg.Log.Format))`.
  `text` in development, `json` elsewhere. `debug` also enables source locations.
- Services log via stdlib `slog` package-level funcs with request ctx:
  success `slog.InfoContext(ctx, "op done", "k", v)`,
  failure `logger.Report(ctx, "op failed", appErr, "k", v)`.
  `Report` maps `*AppError`: 4xx → Warn, 5xx/unknown → Error.
> **Every service function logs both paths — no silent returns.** Each
> success return gets its `InfoContext`; each 5xx/DB-error return gets its
> `Report` first (pure-validation 4xx stays silent, matching file
> precedent). A `return nil, common.Internal()` without a `Report` above it
> is a defect, and a success return without an `InfoContext` is one too —
> check both on every function you write or touch, in every module.
- Never `fmt.Print`/`log.Print`/`slog.New` at call sites.
- Context carries `request_id` + `user_id` (canonical `userID`, Clerk subject):
  `logger.WithRequestID/WithUserID`, auto-attached by `contextHandler`.
- Middleware: `RequestID()` honors/generates + echoes `X-Request-ID`;
  `RequireAuth()` stamps `userID` into gin ctx + log ctx;
  `RequestLogger()` writes one record per request
  (`method,path,status,duration_ms,client_ip,size`), skips `/health` + `/swagger/`;
  `Recovery()` logs panic + stack via `slog.ErrorContext`, returns `common.Internal()`.
- Tests: `logger.NewWithWriter(buf, level, format)` for buffer assertions
  (no ANSI in buffers, JSON fields, `Report` level mapping).

## Shared contracts package (Go DTOs → TypeScript)

- Config `<api-dir>/tygo.yaml`; output `<contracts-dir>/src/*.ts`.
- Commands (from `<api-dir>`): `make contracts` or `tygo generate`.
- Rules: generated files overwritten — hand code only in
  `<contracts-dir>/src/index.ts` (barrel: re-exports, the star-export ambiguity
  resolutions, and the `declare global` enum aliases).
  `time.Time` → RFC3339 string. Export only wire DTOs; runtime structs must not
  leak. `common/errors.go` excluded (`AppError` never crosses).
- **Enums stay `union` style — never `enum_style: "enum"`.** A TS `enum` is
  nominal, so a string literal is not assignable to it: switching broke ~15
  existing sites that assign `"B1"`/`"recent"`/`"lesson"`/`"in_progress"`/
  `"free_talk"` to enum-typed fields, plus `contracts` `tsc` (spike
  2026-10-15). `union` keeps the structural string-literal type the app relies on.
- **Member lists are generated, not hand-written.** `make contracts` runs
  `packages/contracts/scripts/gen-enum-values.mjs` after `tygo generate`,
  emitting `enum-values.ts` (`<ENUM>_VALUES` arrays, `as const satisfies readonly
  Enums.<Enum>[]`) from `enums.ts`. Never hand-write an enum's members in TS.
- tygo has no cross-package resolution, so each enum needs one `type_mappings`
  line **and** one `declare global` alias in the barrel (generated files are
  import-free); a missing alias is a `Cannot find name` error, a missing mapping
  degrades to `unknown /* pkg.Type */`.
- Frontend imports types + limit constants from `<contracts-pkg>` only.

## Web structure (`<web-dir>` — TanStack Start + Router, TanStack Store/Query, Clerk, shadcn/ui, Paraglide i18n)

```text
<web-dir>/
  messages/{en,vi}/<domain>.json  # inlang source (common, nav, onboarding, dashboard, lessons, vocabulary, voice, progress, errors, about)
  project.inlang/settings.json    # baseLocale vi, locales [en,vi], pathPattern array
  scripts/check-conventions.mjs   # pnpm check:conventions — message ids, dynamic keys, hardcoded URLs
  vite.config.ts                  # paraglideVitePlugin strategy ["cookie","baseLocale"]
  src/
    app/
      app-route.ts                # APP_ROUTES, nested by resource — the only frontend URLs allowed
      api-routes.ts               # API_ROUTES, nested by resource, (id) => … leaves — backend paths only
      breadcrumbs.ts              # breadcrumb protocol: staticData specs, targets from APP_ROUTES (labels are thunks)
    assets/topics/*.png
    components/
      common/                     # Logo, PageHeader, StatCard, CardShell, CefrBadge, SourcePill, AudioButton, Kbd, ProgressBar, SubmitButton
      layout/                     # AppShell, AppSidebar, Topbar, LocaleSwitcher (EN/VI pill)
      ui/*                        # shadcn primitives (a11y copy via common.a11y.*)
    features/<domain>/            # attempts, dashboard, dictation, lessons, onboarding, profiles, reading, vocabulary, voice, writing
      components/                 # presentational only; copy via m["<domain>.*"](), never hardcoded UI strings
      fixtures.ts                 # mock-only data (learning content stays English)
      store.ts                    # @tanstack/store where state is cross-screen (profiles, lessons, vocabulary, voice)
      types.ts | parts.ts | dates.ts | diff.ts | scripts.ts  # domain-local helpers, only where needed
    hooks/use-mobile.ts
    integrations/
      clerk/                      # provider, header-user
      tanstack-query/             # root-provider, devtools
    lib/
      api.ts                      # fetch wrapper (envelope unwrap, ApiError)
      auth-guard.ts               # Clerk guard: beforeLoad + server auth() probe + <Show> defence-in-depth
      auth-token.ts               # token getter bridge for non-component api.*
      paraglide-middleware.ts     # official paraglideMiddleware in requestMiddleware (no node:* imports)
      utils.ts
    paraglide/                    # generated-only (messages/runtime/server), never hand-edit
    routes/                       # directory convention; thin shells composing features + layout
      __root.tsx, _app/route.tsx (Clerk guard) + _app/<domain>/..., _app/about.tsx
      not-found.tsx, error.tsx    # status pages at root, OUTSIDE _app — must render signed-out
      onboarding.tsx, sign-in/out
    router.tsx, start.ts (requestMiddleware), styles.css, routeTree.gen.ts, env.ts
```

- `lib/api.*` pattern: base URL from env, Bearer from auth provider,
  envelope `{message,data,pagination}` unwrap, `!ok -> ApiError(status,message)`.
- `features` pattern (mock phase, no fetching layer yet): `fixtures.ts` (typed
  consts, `@tanstack/store` where cross-screen, `useState` for ephemeral)
  → `components/` → `routes/` (thin shell). Types via
  `import type {...} from '<contracts-pkg>'`; fixture-held UI chrome lives in
  `messages/` keyed by item id, never duplicated in fixtures.
- Breadcrumbs come from the router's match chain: each route declares
  `staticData: breadcrumb(...)`, `Topbar` renders links from declared targets.
  Neither route ids nor paths are hardcoded outside `APP_ROUTES`.
- Errors: one `ErrorPage({ title, body, action? })` in `components/common/`, with
  `not-found.tsx` / `error.tsx` at route root passing their own copy from
  `errors.*`. It renders bare (no `AppShell`, no Clerk) precisely because it must
  work signed-out; `error-pages.tsx` keeps `RouteNotFound` /
  `RouteErrorFallback` as thin adapters for `__root.tsx` and `_app/route.tsx`.
- `PageLayout` hero action: drill-down pages pass back navigation via the
  `action` slot with `components/common/BackButton` (`APP_ROUTES` target +
  label from messages), never a hand-rolled button row in `children`. The slot
  renders as its own left-aligned row above the hero (or above the content
  when there is no hero).
- Mutation buttons: anything driven by a `useMutation` (or equivalent async
  `pending` flag) uses `components/common/SubmitButton` with `pending={...}`,
  never a raw `Button` with manual `disabled={...isPending}`. Plain `disabled`
  stays for non-mutation guards only (empty input, locked question, null id).
  Omit `loadingLabel` unless the pending label differs — the default holds
  the label invisible and centers the spinner over it (width frozen),
  needing no new message key.
- Pending frames: a card whose content waits on a query renders through
  `components/common/CardShell` — `pending` + a shape-matched `skeleton` set
  mirroring the loaded layout (no shift when the data lands), or `pending`
  alone to hold fixed-shape children invisible behind a centered spinner. A
  block pending inside an already-rendered card holds its own shape with
  skeletons instead (never a nested shell); never a floating skeleton where a
  card will land, and a page never `return null` while loading. One exemption:
  a section that renders nothing when empty anyway (a "recent" preview strip)
  renders nothing while pending too — skeleton-then-vanish is the worse flash.
- Stat metrics: a metric tile renders through `components/common/StatCard` —
  never a hand-rolled `surface-card` carrying its own label/number markup. The
  frame fixes exactly three things so they cannot drift: the title's typography,
  the metric's `size` (`md` = `text-2xl`, `lg` = `stat-display`), and the
  `CardShell` pending contract. Everything else belongs to the caller: `aside`
  is the right column (ring, icon tile, badge), `children` is the body under the
  metric (bar, caption, a CTA pinned with `mt-auto`), and `skeleton` overrides
  the placeholder when those are an unusual shape. Never add a prop that switches
  on a widget type (`ringValue`, `icon`, `bar`, `tone`) — a new visual goes in
  `aside` or `children`, and a prop per caller is the smell this replaced. The
  grid (`md:grid-cols-3`, …) stays at the call site: the card is the unit, not
  the row, and a row's stats mapping (which three metrics a page shows) stays
  with that page or feature.
- Pagination: numbered lists page through `components/common/Pager`, never a
  hand-rolled pager. It wraps the shadcn `Pagination` primitives
  (`ui/pagination.tsx` is house-adapted: `PaginationLink` has `asChild`, the
  active number uses the primary variant) and takes
  `{ page, totalPages, onPageChange, summary? }`. It always renders for a real
  list — a single page shows one active number with both arrows disabled, so
  the control never pops in and out as filters change the page count (zero
  items renders the empty state, not the pager). The caller owns navigation via
  `onPageChange` and supplies the `summary` copy, so the row reads as one unit
  with no locale string in the component: a range while the list spans pages, a
  plain count on one page (`vocabulary.hub.showingRange` / `showingAll`
  precedent). The page lives in the URL (`validateSearch` + `loaderDeps`),
  coerced by `lib/search-params.ts` (`parsePage` handles both the number the
  router parses out of `?page=2` and a garbage string).
- Enum pickers derive their options from the generated value array:
  `enumOptions(CEFR_VALUES, LABELS?)` — never a hand-written options array.
  Codes (CEFR, ScenarioDifficulty) render as the code; prose enums supply a
  `Record<Enum, () => string>` of labels, so a new member is a compile error
  until its copy exists, and the picker itself never changes.
- React effects (`docs/superpowers/rules/react-effects.md` is the authority):
  `useEffect` syncs React with outside-React only — never React state with
  React state. Derive during render, use event handlers for user work, `key`
  for entity resets, TanStack Query for server state, TanStack Store for
  client UI state (never server snapshots). Mount queries pair
  `enabled: isLoaded` with retry-except-control-flow, never `retry: false`.
- `scripts/check-conventions.mjs` is the web app's only real test harness, so use
  it as one: run it before a change to watch it fail, then after. It derives the
  watched URL segments from `app-route.ts` and the message accessor from each
  file's import — if you extend either, extend the guard rather than hand-listing.
- Fixed stack: Vite + TS + TanStack Start/Router, TanStack Store/Query, Clerk,
  shadcn/ui + Tailwind, Paraglide (cookie strategy, vi default).
- Voice screens talk WebRTC directly to `<voice-dir>`; Go is never in the audio path.

## mockery (generated mocks only)

- `mockery v3` (testify template). Config `<api-dir>/.mockery.yaml`.
- Command: `make mocks` (or `mockery`) from `<api-dir>`.
- Generated, never hand-edit: `internal/database/repositories/mocks/mock_*.go`
  plus any other interface mocks under owning package `mocks/`.
- Exception (concrete structs only): one hand-maintained factory allowed,
  e.g. `internal/<pkg>/mocks/mock_client.go`. No other hand-written `Mock` structs.

## testify

- `github.com/stretchr/testify`: `require` for fatal premises
  (`NoError`, `NotNil`, `Len`, `ErrorAs`), `assert` for value checks
  (`Equal`, `Contains`, `True/False`). `mock.Anything` for expectations;
  mocks auto-`AssertExpectations` via `t.Cleanup`.
- `requireAppError(t, err, http.StatusX)` helper per service tests package.

## Test style (mandatory)

- Location: per-package `tests/` dirs, external black-box `package tests`.
  Shared fixtures in `services/tests/services_test.go`-style helpers.
- Table-driven, always: `tests := []struct{ name ... }{...}` +
  `for _, tt := range tests { t.Run(tt.name, ...) }`. No single-case tests.
- Mocks mandatory via `repositories/mocks`. Never pass `nil` into constructors.
- GORM: mock via repo interface (preferred) or `sqlmock`; assert
  `ErrRecordNotFound→404`, duplicate→409. Never hit a real DB in unit tests.
- No mirror tests ("hiển nhiên"): do not test env→struct passthrough in
  config loaders, literal echoes (`TableName`, enum strings), or mock-graph
  echoes (service returns the fixture the mock returned). Test branches,
  error mapping, key namespacing, and cross-package drift guards instead.
  The only config tests worth keeping assert real behavior:
  parsing/conversions, fail-closed branches, validation.

## Commands (`<api-dir>`)

| Task | Command |
| --- | --- |
| All tests | `go test ./... -count=1` |
| One suite | `go test ./internal/modules/<domain>/... -count=1 -v` |
| Vet/fmt | `go vet ./...`, `gofmt -l internal/` (must be empty) |
| Mocks | `make mocks` |
| Contracts | `make contracts` |
| Migrations | goose SQL is the **only** schema mechanism — never auto-apply from models. **One migration per table**, named `create_<table>` and ordered so parents precede children (goose orders by version; the creation order IS the apply order). New: `make migration-create name=...`; apply/rollback: `make migration-up` / `migration-down` (`DATABASE_URL`). New tables do not get repositories until a module reads them: the interface arrives with the caller, not before. **This project is pre-deploy: edit the existing migration in place and re-run it (`make migration-down` then `make migration-up`) rather than adding a corrective one.** A `goose_db_version` row only means "applied to some local database", not "released" — there is no environment whose schema you may not rewrite. Do not gate a schema fix on a "frozen"/"deployed" distinction, and do not offer a new migration as the safer option: the databases here are disposable dev copies, the seed data is re-derivable, and splitting one logical change across two files leaves the history lying about the shape of the schema. After re-running an edited chain (`make migration-down && make migration-up`), re-run `make seed-audio` to regenerate lesson audio |
| Seed audio | `make seed-audio` — TTS-seed lesson audio into MinIO; requires migrated DB + `docker compose up -d minio minio-init` |
| Swagger | `make swagger` (`swag init -g cmd/server/main.go`) |

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **Engflex-Ultimate** (9372 symbols, 25060 relationships, 810 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact before editing.** Use `impact({target: "symbolName", direction: "upstream"})` or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .`; report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "dev"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "dev" --repo .`.
- MUST warn on HIGH/CRITICAL `risk` pre-edit; never use `riskSharedAxes` to waive a HIGH/CRITICAL `risk` warning. Compare File/symbol: MCP File omits axes; Graph-RAG expands File.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- **MUST use `query({search_query: "concept"})` for concepts/flows, `context({name: "symbolName"})` for a named symbol, or `impact` for blast radius, on read-only callers, dependencies, imports, or execution flow.** Graph first; text search only for empty/`UNKNOWN`/literals.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/Engflex-Ultimate/context` | Codebase overview, check index freshness |
| `gitnexus://repo/Engflex-Ultimate/clusters` | All functional areas |
| `gitnexus://repo/Engflex-Ultimate/processes` | All execution flows |
| `gitnexus://repo/Engflex-Ultimate/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
