# IELTS source packages

Publisher: `apps/api/scripts/publish_ielts.py`. DOM adapter:
`scripts/study4_ielts_content.py`. These commands use the existing uv-managed
`scripts/toeic_pipeline` environment and its `study4` dependency group.

From repository root:

```powershell
uv run --project scripts/toeic_pipeline --group study4 --inexact python apps/api/scripts/publish_ielts.py --export <path-to-export.json> --category-slug ielts-academic --test-slug <test-slug> --sub-type academic
```

Without `--publish`, verifies source coverage, media bytes/types/SHA-256 and the
real Go IELTS parser, then exports the package under this directory's `output/`.
Add `--publish` to upload media first and private `content.json` last to MinIO.
Uses `apps/api/.env` for MinIO, never connects to PostgreSQL. Existing differing
content/media is rejected. Uploaded objects are downloaded and hash-verified.

```text
output/ielts/<category-slug>/<test-slug>/
  content.json
  audio/<test-slug>-section-<number>.mp3
  images/<test-slug>-section-<number>-image-<source-asset-index>.png
```

Image extensions follow verified bytes. Empty media directories are not created.
Local receipts sit next to test directories and are not uploaded. Generated
packages and receipts are gitignored. Source-review status is retained separately;
technical validation and publishing do not attest human review.

Published on 11/11/2026: `ielts-academic/ielts-collection-1-listening-test-1`
(40 questions, four section audio files and one diagram) and
`ielts-academic/ielts-collection-1-reading-test-1` (40 questions, no media assets).
Both tests are now imported into PostgreSQL and integrated into the IELTS frontend.
Runtime content comes from `exam_tests.content`; MinIO serves audio/images only.
Taking responses omit answers/explanations/transcripts; submitted review includes them.

Import from verified MinIO objects (dry run by default; append `--apply` to write):

```powershell
uv run --project scripts/toeic_pipeline --group study4 --inexact python apps/api/scripts/import_ielts.py --test 6842 ielts-academic/ielts-collection-1-listening-test-1/content.json --test 6832 ielts-academic/ielts-collection-1-reading-test-1/content.json
```

The importer validates with the real Go parser, checks category/subtype, question
coverage and media hashes, then writes both tests in one transaction. Identical
existing records are skipped; differing records are rejected without overwriting.

Opt-in real-database tests: from `apps/api`, set `TEST_DATABASE_URL` and
`TEST_PUBLISHED_IELTS=1`, then run `go test ./internal/modules/exams/tests -run TestIeltsPublishedPackagesHTTP -v`.
All test attempts are rolled back. Browser verification reports live under
`docs/crawl/study4/` (ignored workspace artifacts).
