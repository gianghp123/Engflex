# Cấu trúc MinIO của EngFlex

Đối chiếu trực tiếp MinIO local và code ngày **11/11/2026** (Asia/Saigon).
Đây là snapshot của môi trường dev hiện tại; số lượng object có thể thay đổi.
Phiên làm việc sau cần đọc file này trước khi thêm, đổi tên hoặc di chuyển media.

## Kết nối và lưu trữ

- Compose: `apps/api/docker-compose.yaml`, project `engflex`.
- Container: `engflex_minio`; S3 API local mặc định `http://localhost:9000`, console `http://localhost:9010`.
- Container lưu dữ liệu tại `/data`, được mount từ volume Compose `engflex_minio_data`.
- Script đọc `MINIO_ENDPOINT`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `MINIO_BUCKET` từ `apps/api/.env` hoặc environment. Không chép credentials vào tài liệu.
- Web đọc `MEDIA_INTERNAL_ENDPOINT` và `MEDIA_BUCKET` ở server; bucket mặc định là `lesson-audio`.
- “Folder” trong tài liệu là prefix của object key, không phải thư mục filesystem. Bucket và object key là hai thành phần riêng biệt.

## Các bucket đang có

| Bucket | Số object lúc kiểm tra | Nội dung |
| --- | ---: | --- |
| `lesson-audio` | 96 | 18 audio listening và 78 audio mẫu speaking |
| `stories` | 554 | 63 truyện, mỗi truyện có `story.json` và media |
| `toeic` | 138 | 10 đề đã đóng gói, gồm JSON, ảnh và audio |
| `ielts` | 7 | 2 đề Academic: 2 JSON, 4 audio và 1 ảnh |

Bucket IELTS đã được tạo và publish ngày 11/11/2026 theo layout người dùng chốt.
Hai đề đã import vào DB và được FE phục vụ qua luồng IELTS thật (làm bài, autosave, nộp bài, review).

## Cây thư mục

Các tên trong dấu `<...>` biểu diễn quy tắc; ví dụ cụ thể bên dưới đã được kiểm tra.

```text
MinIO
├── lesson-audio/                         # bucket
│   ├── listening/
│   │   └── <lesson-activity-uuid>.mp3
│   └── speaking/
│       └── <lesson-activity-uuid>/
│           ├── 0.mp3
│           ├── 1.mp3
│           └── ...                      # chỉ số item bắt đầu từ 0
│
├── stories/                              # bucket
│   └── <story-slug>/                     # ví dụ: bearskin, cinderella
│       ├── story.json
│       └── assets/
│           ├── narration.mp3
│           ├── cover.webp
│           ├── scene-01.webp
│           └── scene-02.webp             # số cảnh tùy truyện
│
├── toeic/                                # bucket
│   └── <category-slug>/
│       └── <test-slug>/
│           ├── content.json             # full đề, chứa dữ liệu đáp án
│           ├── audio/
│           │   └── <test-slug>-audio.mp3
│           └── images/
│               ├── <test-slug>-part-1-q01.png
│               ├── <test-slug>-part-3-q62-64.png
│               └── ...
└── ielts/                                # bucket
    └── ielts-academic/
        ├── ielts-collection-1-listening-test-1/
        │   ├── content.json
        │   ├── audio/                   # section-1.mp3 đến section-4.mp3, có prefix test slug
        │   └── images/                  # sơ đồ section 3
        └── ielts-collection-1-reading-test-1/
            └── content.json
```

Các đuôi `.webp`/`.png` trong cây là ví dụ. Publisher giữ hoặc chọn đuôi theo content type/file đầu vào.

## Bucket crawl đã xóa

Ngày 11/11/2026, theo yêu cầu người dùng, đã xóa toàn bộ 275 object và bucket
`study4-crawl` (~273 MB), gồm cache, media nguồn, lịch sử chạy, artifact import
và 71 object của đề nguồn `6861` chưa import. Không tạo backup trước khi xóa.
Đã xác nhận bucket không còn tồn tại; ba bucket có sẵn tại thời điểm xóa giữ nguyên số object.
Bucket `ielts` được tạo sau lần đối chiếu đó.

Trước khi xóa đã đối chiếu 29 bảng trong DB `engflex` và 73 JSON đã publish:
không có reference tới bucket hoặc object key của nó. Đã kiểm tra đủ 128 media
của 10 đề TOEIC trong bucket `toeic` bằng SHA-256; 10 `content.json` khớp DB.
Thư mục backup dự kiến `docs/crawl/study4` không tồn tại lúc đối chiếu.
Nếu cần tiếp tục đề `6861` hoặc xem lại nguồn crawl đã xóa, cần nguồn khác hoặc crawl lại.

## Quy tắc và liên kết với ứng dụng

### Nguồn nội dung đề thi

PostgreSQL `exam_tests.content` là nguồn chính khi đọc đề, làm bài và chấm điểm.
MinIO lưu ảnh/audio; `content.json` là bản export tại thời điểm publish, phục vụ
lưu trữ hoặc import lại, không phải nguồn được ứng dụng đọc khi làm bài.
Sửa riêng `content.json` trên MinIO không cập nhật đề trong DB.

### Lesson audio

`apps/api/scripts/seed_audio.py` đọc `lesson_activities.config` trong DB:
listening lấy `audioKey`, speaking lấy từng `items[].modelAudioKey`.
Script upload đúng key được cấu hình, không tự suy ra key mới từ ID.
Chạy lại sẽ ghi đè cùng key. Bucket lấy từ `MINIO_BUCKET`, mặc định `lesson-audio`.

Ví dụ object đã có:

```text
lesson-audio/listening/44444444-0000-4000-8000-000000000112.mp3
lesson-audio/speaking/44444444-0000-4000-8000-000000000114/0.mp3
```

### Stories

`apps/api/scripts/publish_story.py` đặt key theo `<story-slug>/assets/<name>.<ext>`:
audio là `narration`, thumbnail là `cover`, ảnh cảnh là `scene-01`, `scene-02`, ...
Manifest được ghi tại `<story-slug>/story.json`. Manifest chứa URL media cùng origin.

### TOEIC

Nguồn quy tắc: `apps/api/scripts/toeic_storage.py`.
Category/test slug phải là chữ thường, số và dấu gạch nối.
Ảnh được đặt theo lần đầu asset xuất hiện trong nhóm câu hỏi; nhiều ảnh trong
một nhóm có hậu tố `-image-01`, `-image-02`. SHA-256 dùng để kiểm tra nội dung,
không phải quy tắc tên file mới. Validator web vẫn chấp nhận một số tên hash cũ.

Prefix đang có:

- `toeic/ets-2024/ets-2024-test-02/` — 18 object.
- `toeic/toeic-lr-collection-1/toeic-lr-collection-1-test-1/` đến `...-test-9/` — tổng 120 object.

Ví dụ phân biệt identity với object key:

```text
Bucket:       toeic
Object key:   ets-2024/ets-2024-test-02/audio/ets-2024-test-02-audio.mp3
Media ID:     toeic/ets-2024/ets-2024-test-02/audio/ets-2024-test-02-audio.mp3
URL web:      /api/media/toeic/ets-2024/ets-2024-test-02/audio/ets-2024-test-02-audio.mp3
```

Không thêm lại prefix `toeic/` vào object key bên trong bucket `toeic`.
Layout cũ `lesson-audio/exams/toeic/...` không có trong inventory hiện tại.
Xem thêm [hướng dẫn import TOEIC](../../apps/api/scripts/README-toeic-import.md).

### IELTS — layout đã chốt

Quyết định của người dùng ngày 11/11/2026:

```text
ielts/                           # bucket riêng
  <category-slug>/
    <test-slug>/
      content.json               # bản export đầy đủ của đề
      audio/
      images/
```

- `ielts` là tên bucket; object key bắt đầu từ `<category-slug>/<test-slug>/`.
- Category slug dùng slug trong DB; hiện có `ielts-academic` và `ielts-general`.
- Test slug là identity ổn định của đề, dùng cùng slug khi import DB.
- Audio đặt `<test-slug>-section-<number>.mp3`; ảnh đặt
  `<test-slug>-section-<number>-image-<source-asset-index>.<ext>` theo bytes đã kiểm tra.
- Nội dung phục vụ làm bài vẫn đọc từ `exam_tests.content`; ảnh/audio lấy từ MinIO.
- `content.json` private; policy chỉ cho anonymous đọc `*/*/audio/*` và `*/*/images/*`.
- Publisher: `apps/api/scripts/publish_ielts.py`; adapter: `scripts/study4_ielts_content.py`.
  Chạy bằng uv environment của `scripts/toeic_pipeline` với `--group study4`.
- Tổng 7 object (~45.1 MB). Đã đọc lại và xác minh hash cả 7 object; anonymous
  đọc JSON trả 403, Range request tới media trả 206. Reading không có media nên
  không tạo folder rỗng `audio/` hoặc `images/`.
- Schema IELTS đã giữ audio theo section, transcript review, context dạng cây
  có bảng/merged cells/ảnh theo nhóm câu hỏi và đáp án nguồn/biến thể tường minh.
- Route media web đã hỗ trợ IELTS qua `isIeltsMediaKey`; JSON bị chặn (404), media hỗ trợ Range (206).
- DB `exam_tests`: Listening `59dae7d8-7c9f-42da-b237-d6c7db39e8cd`, Reading
  `e5a01334-993f-4efc-ab17-335c83301014`; source keys `study4_6842`, `study4_6832`.
  Cùng category Academic `af4f06b2-520d-47c9-83df-8b1555f25e25`.
- Importer `apps/api/scripts/import_ielts.py` kiểm tra lại Go parser, đủ 40 câu,
  category/subtype và hash media trước transaction. Mặc định dry-run; `--apply`
  mới ghi DB. Chạy lại dữ liệu giống hệt là no-op, dữ liệu khác bị từ chối ghi đè.
- Taking API không trả đáp án, giải thích hay transcript; review sau nộp mới có.
  FE render passage/context/bảng/ảnh, phát audio theo section; tiến độ autosave
  theo revision, reload tiếp tục bài, có reset, practice, hủy và chấm band.
- Đã kiểm tra Go integration trên PostgreSQL thật và browser Clerk đăng nhập:
  cả hai đề qua start → trả lời → autosave → reload → submit → review, không
  có lỗi runtime. Report: `docs/crawl/study4/ielts-browser-report.json`.
  Các attempt/profile tạm của bài kiểm tra đã được dọn.

## Cách web đọc object và quyền truy cập hiện tại

Route: `apps/web/src/routes/api/media/$.ts`.

| URL web | Bucket | Object key |
| --- | --- | --- |
| `/api/media/listening/<id>.mp3` | `MEDIA_BUCKET` (mặc định `lesson-audio`) | `listening/<id>.mp3` |
| `/api/media/speaking/<id>/0.mp3` | `MEDIA_BUCKET` | `speaking/<id>/0.mp3` |
| `/api/media/stories/<slug>/assets/narration.mp3` | `stories` | `<slug>/assets/narration.mp3` |
| `/api/media/toeic/<category>/<test>/audio/<name>.mp3` | `toeic` | `<category>/<test>/audio/<name>.mp3` |
| `/api/media/ielts/<category>/<test>/audio/<name>.mp3` | `ielts` | `<category>/<test>/audio/<name>.mp3` |

Web forward HTTP từ MinIO và chuyển tiếp Range/206 để tua audio. Route media
không yêu cầu đăng nhập. IELTS và TOEIC chỉ nhận media key hợp lệ trong `audio/` hoặc `images/`; không phục vụ `content.json` hay bucket `study4-crawl`.
Go API không nằm trong đường truyền media này.

Policy đã đọc trực tiếp:

- `lesson-audio`: anonymous `GetObject` toàn bucket, có `ListBucket`/`GetBucketLocation`.
- `stories`: anonymous `GetObject` toàn bucket, gồm cả `story.json`.
- `toeic`: anonymous `GetObject` chỉ với `*/*/audio/*` và `*/*/images/*`.
  `content.json` không được policy này công khai; web cũng chặn qua `isToeicMediaKey`.
- `ielts`: anonymous `GetObject` chỉ với `*/*/audio/*` và `*/*/images/*`; JSON private.

## Script và cấu hình cần đọc khi thay đổi

- `apps/api/docker-compose.yaml`: khởi tạo `lesson-audio`, `stories` bằng `minio-init`.
- `apps/api/scripts/seed_audio.py`: tạo/upload audio bài học.
- `apps/api/scripts/publish_story.py`: đóng gói và publish truyện.
- `apps/api/scripts/toeic_storage.py`: key, hash verification, JSON và policy bucket TOEIC.
- `apps/api/scripts/publish_ielts.py`, `scripts/study4_ielts_content.py`: export/publish IELTS có kiểm tra Go parser và SHA-256.
- `apps/api/scripts/publish_toeic_media.py`, `import_toeic.py`: publish/import đề.
- `apps/api/scripts/organize_toeic_storage.py`, `rename_toeic_media.py`: bảo trì layout TOEIC.
- `scripts/STUDY4.md`: luồng export và import nguồn Study4.
- `apps/web/src/routes/api/media/$.ts`, `apps/web/src/features/exams/media-key.ts`: resolve và kiểm tra key public.

Trước khi đổi tên hoặc di chuyển object, đối chiếu reference trong DB và JSON
manifest; đổi key trên MinIO riêng lẻ sẽ làm hỏng media đang được tham chiếu.
Khi thêm bucket/prefix hoặc đổi quy tắc, cập nhật README này theo dữ liệu thực tế.

## Kiểm tra lại inventory (chỉ đọc)

Chạy PowerShell từ `apps/api`, dùng client chung và `.env` hiện có:

```powershell
@'
from pathlib import Path
from collections import Counter
import sys
from dotenv import load_dotenv
sys.path.insert(0, "scripts")
from toeic_storage import minio_client
load_dotenv(Path(".env"))
client = minio_client()
for bucket in client.list_buckets():
    keys = [o.object_name for o in client.list_objects(bucket.name, recursive=True)]
    print(bucket.name, len(keys), dict(Counter(k.split("/")[0] for k in keys)))
    print("\n".join(keys[:10]))
'@ | uv run --no-project --with minio --with python-dotenv python -
```

Tài liệu này được đưa vào Git theo yêu cầu người dùng ngày 12/11/2026 để các
phiên làm việc sau tra cứu. Media nguồn, dữ liệu crawl và báo cáo kiểm thử vẫn
được lưu local; không nằm trong commit tài liệu.
