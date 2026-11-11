# Output khớp MinIO

Lệnh `release` xuất package vào `output/toeic/<category-slug>/<test-slug>/`:

```text
content.json
images/<test-slug>-part-1-q01.png
images/<test-slug>-part-3-q62-64.png
audio/<test-slug>-audio.mp3
```

Cấu hình qua `--category-slug`, `--test-slug`, `--output-root`. Mặc định là
`ets-2024`, `ets-2024-test-NN`, và `scripts/toeic_pipeline/output`.
Quy tắc đặt tên và kiểm tra hash dùng chung với publisher MinIO.

```powershell
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py release --test 2 --category-slug ets-2024 --test-slug ets-2024-test-02 --output-root scripts/toeic_pipeline/output
```

Import qua `apps/api/scripts/import_toeic.py --package-root <output-root>` kèm
`--release <data/test_NN/release>` và các tham số category/test. Release vẫn phải
qua validation và review; provenance, bằng chứng và correction giữ ở workspace
pipeline. Chi tiết lệnh: `apps/api/scripts/README-toeic-import.md`.

# TOEIC 2024 Pipeline — Kế hoạch ưu tiên độ chính xác

## 1. Mục tiêu và trạng thái

Chuyển PDF đề, sách giải và audio trong `Topic Test/Toeic 2024/` thành dữ liệu có cấu trúc, giữ nguyên nội dung và truy ngược từng trường về nguồn.

**Plan bên dưới là tiêu chuẩn nghiệm thu. Source Python phiên bản 0.1 đã có để chạy pilot và tạo draft; chưa đồng nghĩa dữ liệu đã được duyệt. Xem mục 12 để chạy implementation hiện tại và biết các giới hạn.**

Ưu tiên: đúng bộ đề → đúng nội dung/đáp án → đúng quan hệ/bố cục → đúng đoạn audio → tốc độ/chi phí. Không cam kết tự động chính xác tuyệt đối. Release yêu cầu kiểm tra máy và người đối chiếu nguồn; confidence của OCR/LLM không chứng minh dữ liệu đúng.

Đầu ra gồm `draft/` (có thể thiếu nhưng ghi rõ lỗi), `review/` (bằng chứng và correction), `release/` (đầy đủ, đã duyệt, không còn issue mở). Pipeline không tự đưa draft lên frontend hoặc ghi vào DB.

## 2. Baseline nguồn đã kiểm tra

Kiểm tra trực tiếp ngày 2026-10-30. Mỗi lần chạy vẫn phải kiểm kê lại.

| Nguồn | Quan sát |
| --- | --- |
| LC | 10 PDF × 13 trang, không có text layer |
| RC | 10 PDF × 29 trang, không có text layer |
| Ảnh đề | Khoảng 3.2k × 4.4k px, kích thước thay đổi; có watermark và trang có ghi chú viết tay |
| Answer LC ETS.pdf | 320 trang digital text, nhiều font và bố cục |
| Answer RC Test.pdf | 415 trang digital text; lấy đáp án/dịch/giải thích, không dùng thay đề tiếng Anh gốc |
| Answer.pdf | 4 trang, bảng đáp án LC nhiều test/cột; chưa có bảng RC độc lập đã xác nhận |
| Audio | 10 MP3, 44.1 kHz, 128 kbps, thời lượng thay đổi |

| Test | Thời lượng giây, làm tròn 3 chữ số |
| --- | ---: |
| 01 | 2752.521 |
| 02 | 2752.521 |
| 03 | 2698.545 |
| 04 | 2749.211 |
| 05 | 2728.569 |
| 06 | 2736.597 |
| 07 | 2729.820 |
| 08 | 2752.651 |
| 09 | 2770.243 |
| 10 | 2751.895 |

Không dùng 2752 giây làm giới hạn chung. Audio 01/02 cùng thời lượng nhưng khác SHA-256; duration không chứng minh đúng danh tính file.

Rủi ro đã thấy:

- Header sách giải có chỗ ghi ETS 2023, chỗ ghi 2024. Phải xác nhận bằng nội dung, chưa kết luận sai bộ chỉ từ header.
- Font đậm gồm Calibri, TimesNewRoman, Arial… và còn dùng cho tiêu đề, số câu, dịch, bằng chứng. Không bắt đáp án chỉ bằng `Calibri-Bold`.
- Đề có ảnh ví dụ trong directions, hai cột, bảng và nhóm bài đọc xuyên trang; text phẳng không đủ giữ nghĩa.
- Ví dụ README cũ chưa được xác thực như fixture. Test 02 RC câu 147–148 thực tế ở trang PDF 9 là bài “Focus Your Social Media Presence”; bỏ các ví dụ mang ID thật nhưng nội dung chưa đối chiếu.

## 3. Nguồn sự thật

| Trường | Nguồn chính | Đối chiếu |
| --- | --- | --- |
| Stem/options/passages in trên đề | PDF LC/RC đúng test | Người duyệt xem ảnh gốc; OCR thứ hai hỗ trợ phát hiện lỗi |
| Ảnh/graphic/bảng/bố cục | Vùng ảnh PDF đề | So crop với trang gốc, giữ đủ nhãn/chú thích |
| Nội dung lời nói và thời gian | Audio đúng test | Sách giải là transcript ứng viên; nghe giải quyết bất đồng |
| Đáp án LC 1–100 | Bảng Answer.pdf đã xác nhận test | Đáp án sách giải LC và người duyệt |
| Đáp án RC 101–200 | Cột đáp án sách giải RC đã xác nhận test | Người duyệt đúng hàng và đúng đề; ghi rõ chỉ có một nguồn đáp án độc lập |
| Giải thích/dịch/từ vựng | Sách giải tương ứng | Đối chiếu đúng câu và đoạn |

Không có phép cross-check ba nguồn chung cho LC/RC. OCR hai lần cùng một bảng không tạo hai nguồn độc lập.

Bất đồng tạo issue với giá trị ứng viên và bằng chứng, chặn release. Người duyệt quyết định và ghi lý do. Nếu nguồn sai, giữ raw extraction và lưu correction riêng. Chưa giải quyết được thì giữ draft; không chọn theo đa số, không nhờ LLM giải đề hoặc sửa answer cho khớp explanation.

Dịch/giải thích/vocabulary không có trong nguồn được để null hoặc mảng rỗng có lý do. Nếu cần biên soạn bổ sung, dùng luồng riêng có nhãn `authored`, bằng chứng và người duyệt; không trình bày như trích từ sách.

## 4. Luồng xử lý

```text
Inventory + xác nhận danh tính nguồn
→ trích sách giải theo vùng + OCR đề có tọa độ + crop
→ ghép part/group/question/passage qua trang
→ cấu trúc hóa có evidence
→ ASR + alignment theo thứ tự thời gian
→ validate + đối chiếu toàn bộ nội dung/đáp án/audio
→ draft hoặc release đã duyệt có checksum
```

### 4.1. Inventory và manifest

- Ghi path, SHA-256, số trang, kích thước/rotation trang và ảnh; audio ghi codec/channels/sample rate/duration đầy đủ.
- Map tường minh theo `collectionId/testNumber/section`; xử lý tên có/không có `_`, không dựa vào directory order hoặc sort chuỗi `1,10,2`.
- Lập mapping vùng/trang cho mỗi test trong ba sách đáp án. `pageNumber` đếm từ 1 theo PDF; `printedPageLabel` lưu riêng.
- Người duyệt xác nhận bằng câu/nhóm đầu-cuối mỗi part và nhóm graphic; audio nghe đầu/cuối từng part. Nếu lệch, dừng ghép và kiểm tra toàn phần liên quan.
- Tách directions/example khỏi câu thi. Nhãn năm không phải bằng chứng duy nhất để xác nhận bộ đề.

### 4.2. Sách giải và đáp án

- PyMuPDF lấy span, bbox, font, flags; phân vùng cột trước khi nối hàng. Giữ raw và xử lý câu/hàng ngắt trang.
- Parser phân biệt test, part, question, transcript, answer, translation, vocabulary. Không dùng một regex/font cố định cho toàn bộ sách.
- Answer.pdf: map nhãn test với vùng cột bằng tọa độ trước khi đọc cặp số câu/đáp án. Không tin thứ tự text extraction; đặc biệt kiểm tra thứ tự test 6/7. Đủ 100 câu duy nhất/test và Part 2 không có D.
- LC: chỉ xét nhấn mạnh trong vùng lựa chọn đúng câu, dùng nhiều tín hiệu bố cục/font. Thiếu hoặc nhiều ứng viên đưa đi review.
- RC: đọc cột đáp án theo hàng 101–200, gắn đúng lời giải/dịch; không lấy chữ A–D trong phần giảng làm đáp án.
- Speaker/accent giữ như nguồn nếu có bằng chứng; không suy từ ASR. Marker số câu chèn trong sách giải không phải lời nói.

### 4.3. OCR và crop

- LC: lấy 6 ảnh câu thật Part 1; OCR toàn bộ câu hỏi/options và graphic Part 3–4. Part 2 lấy nội dung nghe từ audio/sách giải, không tự dựng câu hỏi in trên đề.
- RC: OCR đầy đủ Part 5–7, gồm email/chat metadata, số liệu, bảng, blanks, nhãn chèn câu và lựa chọn.
- Xác định ảnh scan chính theo vị trí/kích thước, không nhầm watermark. Giữ render gốc và phép biến đổi xoay/deskew/scale.
- Bbox lưu trong hệ tọa độ trang PDF đã chuẩn hóa rotation, gốc trên trái, đơn vị point, dạng `[x0,y0,x1,y1]`; lưu rotation và mapping pixel OCR về trang.
- OCR theo vùng/cột, giữ reading order. Bỏ header/footer bằng quy tắc có kiểm chứng; không xóa text passage chỉ vì lặp lại.
- Đối chiếu OCR thứ hai cho vùng khó: số tiền/ngày, tên riêng, phủ định, dấu câu, blank, A–D. Engine/model khác chỉ tạo tín hiệu bất đồng, không tự chọn kết quả theo confidence.
- Vùng khó đọc phải render lại/xem nguồn/chép tay có duyệt; không để LLM đoán chữ mất.
- Crop giữ đủ legend/caption, bỏ ghi chú ngoài vùng đề, không lấy ảnh ví dụ. Nếu nét viết tay che nội dung thì tạo issue, không tự phục dựng phần bị che.
- Master crop PNG; ảnh phân phối WebP lossless chỉ sau kiểm tra đọc rõ. Giữ ảnh nguồn bảng/passage làm fallback bố cục.
- Benchmark trên scan thật; bỏ cam kết 0.3 giây/trang khi chưa đo.

### 4.4. Ghép cấu trúc và LLM

- Khóa danh tính bằng `(collectionId,testNumber,partNumber,questionNumber)`. Group mang trực tiếp passages/questions; không xuất flat list để frontend tự regroup.
- Ghép theo question range in trên đề và tiếp nối trang; không coi mỗi trang là một group.
- Parser lấy các trường xác định chắc chắn. LLM hỗ trợ gán block/quan hệ và chuẩn hóa cấu trúc, không giải đề hoặc tự thêm nội dung.
- Gửi từng group đầy đủ với block IDs, đúng vùng sách giải và ảnh liên quan nếu model hỗ trợ; không đưa cả sách vào một prompt. Nếu text không đủ biểu đạt bố cục, phải xem ảnh/review.
- LLM trả nguồn block/span cho nội dung đã lấy; thiếu bằng chứng trả null + issue. Không thêm missing option, sửa đáp án hoặc viết lại tiếng Anh.
- Giữ text gốc. Chuẩn hóa Unicode/line-wrap theo quy tắc; sửa chính tả OCR, nối từ qua dòng hoặc đổi punctuation phải có diff được kiểm tra.
- Validate schema và đối chiếu nội dung với block nguồn. JSON hợp lệ không chứng minh nội dung đúng. Truncated/parse lỗi retry giới hạn rồi đưa review, không bỏ group.
- Cấu hình API/model qua environment. DeepSeek chính thức dùng `deepseek-flash` theo docs tại thời điểm kiểm tra; lưu returned model, ngày chạy, prompt/schema version và usage. Alias thay đổi không phải snapshot bất biến.

### 4.5. Audio và alignment

- Faster-whisper lấy segment và `word_timestamps`, lưu raw output. So transcript sách giải với lời thực tế, không ép ASR thành bản sách.
- Anchor bằng số câu/nhóm và nội dung đặc trưng; alignment đơn điệu trong từng part. Không fuzzy-search từng câu độc lập trên toàn file.
- Có thể thử thư viện forced alignment với transcript đã kiểm tra nếu môi trường hỗ trợ. Forced alignment cũng cần nghe xác nhận.
- Timestamp luôn theo MP3 nguồn; chunk/VAD/resample phải lưu offset/mapping, kiểm tra không cộng offset hai lần hoặc làm mất khoảng lặng.
- Part 1–2: `itemAudio` gồm số câu và toàn bộ prompt/options được đọc, kết thúc sau lời cuối trước câu kế tiếp.
- Part 3–4: `stimulusAudio` chỉ hội thoại/bài nói; `itemAudio` gồm giới thiệu nhóm, stimulus và phần câu hỏi được đọc, dừng trước nhóm kế tiếp. Có thể lưu `questionAudio` riêng nếu cần.
- Range `[startTime,endTime)` tính bằng giây. Padding chỉ ở khoảng trống, không cắt lời/lấn item sau; stimulus nằm trong item.
- Đủ 54 item LC/test: 6 + 25 + 13 + 10. Kiểm tra coverage/thứ tự, không gán nhiều nhóm vào cùng đoạn.
- Nghe đầy đủ từng item và hai ranh giới; sửa transcript theo audio có correction. Confidence/similarity chỉ xếp thứ tự review, không tự cho release.

## 5. Hợp đồng dữ liệu

Đây là **import format có version**, chưa tương thích trực tiếp với models/contracts Engflex hiện tại. Tạo Pydantic models, export JSON Schema và khóa fixture đã duyệt trước khi viết extractor đầy đủ. Không hand-edit generated schema.

Implementation phân biệt schema extraction (`draft.json`, có evidence) với schema server import (`content.json`, allowlist). Thiết kế DB kết hợp bảng/JSONB và quy tắc export ở mục 13; không nhập nguyên draft vào DB.

Quy ước: camelCase, UTF-8, timestamp số hữu hạn; ID chứa collection/test, ví dụ `toeic_2024_test_02_q147`. Root gồm `schemaVersion`, `collectionId`, `testId`, `testNumber`, `title`, `audioAssetId`, `assets`, `parts`. Manifest/evidence/review là sidecar có hash liên kết payload.

| Node | Cấu trúc |
| --- | --- |
| Part | partNumber, section, items; counts tính từ questions |
| Single (Part 1,2,5) | id, type=single, question; media/transcript/audio khi áp dụng |
| Group (Part 3,4,6,7) | id, type=group, passages, questions; media/transcript/audio khi áp dụng |
| Question | id, questionNumber, kind, stem, options, correctKey, explanation, vocabulary; target/media khi áp dụng |
| Option | key, text; Part 1–2 text=null, lời đọc nằm trong transcript |
| Transcript line | id, role=prompt/option/utterance, speaker, optionKey, text, vietnameseText; trường không có bằng chứng để null |
| Passage | id, format, blocks, media; bản dịch lồng theo passage/block tương ứng |
| Media reference | assetId, role, caption; gắn ở đúng item/passage/question sở hữu |
| Audio range | assetId, startTime, endTime; dùng chung cho item/stimulus/question |

Part 1–2 không đặt lời đọc vào options hoặc lấy hướng dẫn chung làm stem. Part 2 có prompt riêng trong transcript. Part 3–4 giữ nguyên options in trên đề.

### 5.1. Passage và vị trí câu hỏi

- Block là discriminated union: paragraph, heading, table, image, emailHeader, chatMessage. Giữ paragraph breaks, reading order và nhấn mạnh có ý nghĩa với câu hỏi.
- Table giữ rows/cells/spans nếu có và ảnh nguồn đối chiếu; không chuyển thành prose.
- Paragraph dùng inline nodes: text, blank, insertionPoint, markedText.
- Part 6: mỗi blank có ID và questionId; câu 131–146 trỏ đúng một blank, kể cả điền cả câu. Không điền đáp án đúng vào passage.
- Part 7: target chèn câu trỏ passage và insertion points [1]–[4]; từ/câu được gạch chân trỏ marked span. UI không tự đoán vị trí từ text.
- Single/double/triple passage xác định bằng 1/2/3 tài liệu thực tế, không bằng số trang. Một passage có thể nhiều blocks/vùng nguồn qua trang.
- Closed sets dùng typed union/enum và ràng buộc theo part. Khi tích hợp Go, enum sản phẩm dùng enums.* và DB CHECK theo AGENTS.md, không ép model/repository để khớp import DTO.

### 5.2. Evidence, correction và review

Mỗi leaf nội dung/đáp án và mỗi quyết định group/crop/audio có evidence:

- `target`: JSON Pointer tới field hoặc ID quan hệ.
- `sourceRefs`: file ID/hash, pageNumber+bbox đối với PDF; asset ID/hash+range đối với audio. Một field có thể dẫn nhiều vùng/trang.
- `rawValue`, `normalizedValue`, method, extractor/model/config version và confidence gốc nếu có; không gộp các confidence thành xác suất đúng.
- Issues gồm loại, mức độ, các giá trị ứng viên và trạng thái giải quyết.
- Review gồm người duyệt, thời điểm, quyết định, lý do correction và hash nội dung đã duyệt.

Trạng thái: extracted, needsReview, verified, rejected. Release yêu cầu nội dung bắt buộc verified. Optional null phải ghi notInSource/notApplicable; unreadable vẫn chặn nếu cần để làm/chấm bài. Không thay missing bằng chuỗi rỗng hay đáp án mặc định.

Assets dùng key tương đối và checksum, không hardcode URL frontend hoặc công khai path local. Importer/storage giải quyết URL. Sách giải/evidence không nằm trong payload làm bài.

## 6. Validation và cổng release

### 6.1. Kiểm tra máy

| Part | Số câu | Cấu trúc |
| --- | --- | --- |
| 1 | 1–6 | 6 single, A–D, mỗi câu một ảnh thật |
| 2 | 7–31 | 25 single, A–C |
| 3 | 32–70 | 13 group × 3, A–D |
| 4 | 71–100 | 10 group × 3, A–D |
| 5 | 101–130 | 30 single, A–D |
| 6 | 131–146 | 4 group × 4, blank liên kết đầy đủ |
| 7 | 147–200 | 54 câu, group range/passages lấy từ đề |

- Đủ 7 part/200 số câu duy nhất liên tục, ID duy nhất toàn collection; không overlap hoặc mất group.
- Option keys đúng tập/thứ tự; correctKey thuộc options; stem/options bắt buộc không rỗng, không placeholder hoặc text bị cắt.
- LC đối chiếu đủ 100 cặp đáp án/test; RC đủ 100 hàng. Bất đồng dù một câu vẫn chặn đến khi correction được duyệt.
- Blank/target/reference resolve đúng group; không passage/crop/transcript gắn nhầm test.
- Audio: `0 <= startTime < endTime <= actualDuration`; đúng asset, số hữu hạn, thứ tự và quan hệ chứa hợp lệ. Tolerance rounding nếu cần phải ghi rõ trong config, không che lỗi cắt đoạn.
- Assets tồn tại, decode được, hash khớp; bbox nằm trong trang, crop đủ nhãn và không lấy ảnh ví dụ.
- Strict schema chặn unknown fields/version; kiểm tra evidence coverage và null reason. Không sửa answer hoặc bỏ câu để validator xanh.

### 6.2. Review bắt buộc cho từng test

1. Đối chiếu toàn bộ stem/options/passages với ảnh gốc, đặc biệt tên riêng, phủ định, ngày/tiền và bảng.
2. Đối chiếu toàn bộ đáp án, liên kết lời giải/dịch và vocabulary. Nguồn có lỗi tạo issue, không tự viết lại để che lỗi.
3. Nghe toàn bộ 54 item LC, kiểm tra transcript, đúng câu/test và hai ranh giới không cắt lời.
4. Xem toàn bộ crop/passage preview, gồm nhóm nối trang, bảng, graphic và insertion targets.
5. Lượt kiểm tra độc lập thứ hai cho 200 đáp án/test và mọi correction/bất đồng, bảng/số liệu, blank/insertion target và boundary/crop từng bị sửa. Thiếu lượt độc lập thì chưa release.

Review hiển thị nguồn-output song song, phát đúng audio và sửa từng field; không dùng nút duyệt tất cả thay đối chiếu. Người dùng có thể thực hiện review sau khi pipeline tạo gói bằng chứng cụ thể.

**Release = validators pass + review coverage đầy đủ + không còn issue mở + hash đã duyệt khớp payload.** Sửa nội dung làm mất hiệu lực review của field và các quan hệ phụ thuộc; phải duyệt lại trước release. Chỉ correction đã được duyệt mới đóng issue.

Báo cáo số lỗi/correction theo loại, trường thiếu, bất đồng nguồn, runtime/VRAM/chi phí và sai lệch audio boundary trên mẫu gán tay. Mục tiêu pilot: không phát hiện câu bị cắt hoặc lẫn câu sau; đo sai lệch từng boundary so với mốc người nghe xác nhận, mốc đề xuất tối đa 0.2 giây cần kiểm chứng bằng nghe thực tế. Không suy ra “100% chính xác” từ đủ 200 câu hoặc schema pass.

## 7. Công nghệ và môi trường

| Thành phần | Yêu cầu |
| --- | --- |
| Python + uv | Project riêng: pyproject.toml, uv.lock, .python-version; luôn uv run --project scripts/toeic_pipeline |
| Python version | Thử 3.13, chỉ khóa sau smoke test PDF/OCR/ASR trên Windows; nếu không tương thích chọn phiên bản hỗ trợ toàn bộ dependencies và ghi quyết định |
| PyMuPDF | Text spans, bbox, render, ảnh và coordinate transforms |
| RapidOCR + ONNX Runtime | Baseline OCR; khóa package/model/checksum sau benchmark |
| Pillow | Master PNG + WebP lossless sau kiểm tra |
| Faster-whisper/CTranslate2 | Word timestamps; pin CUDA/cuDNN/CTranslate2 tương thích, smoke test GPU thật |
| Forced alignment | Ưu tiên thư viện có sẵn nếu thử nghiệm môi trường/chất lượng đạt |
| Pydantic + JSON Schema | Strict typed schema và cross-field checks |
| LLM adapter | Structure có evidence, timeout/retry/budget, lưu raw response |

Máy kiểm tra có RTX 3050 Ti Laptop 4 GB VRAM. Benchmark model/quantization với batch thấp, có CPU fallback; không mặc định large FP16. Cài Torch CUDA không chứng minh CTranslate2 chạy được. Model nhỏ chỉ dùng nếu chất lượng pilot đạt; không giảm cổng duyệt để chạy nhanh hơn.

Tài liệu: [Faster-whisper](https://github.com/SYSTRAN/faster-whisper), [RapidOCR](https://github.com/RapidAI/RapidOCR), [DeepSeek API](https://api-docs.deepseek.com/en/). Kiểm tra lại khi khóa dependencies/API.

## 8. Cache và chạy lại

- Cache key theo stage + source hashes + config + dependency/model + schema/prompt version. Upstream đổi invalid downstream liên quan; không chỉ cache theo tên test.
- Raw artifacts bất biến; normalized draft và correction overlay riêng. Correction gắn field ID/hash; nguồn đổi phải tạo conflict, không patch nhầm.
- Lưu OCR/ASR và request/response đã loại secrets. API key chỉ từ environment.
- Checkpoint từng group, atomic writes, khóa chống ghi đồng thời; retry giới hạn/backoff. Group lỗi chặn release, không bị bỏ qua.
- Export thứ tự ổn định; cùng raw artifacts/config/corrections cho cùng content hash. LLM gọi lại có thể khác nên version hóa response, không hứa deterministic inference.
- --skip-audio chỉ hợp lệ khi cache đủ và hash đúng audio/config; thiếu/stale phải lỗi.
- Ignore .venv, cache, review media, output và model weights. Không commit tài liệu nguồn/media; cleanup không xóa nguồn.

## 9. Cấu trúc dự kiến

```text
scripts/toeic_pipeline/
├── README.md
├── pyproject.toml / uv.lock / .python-version
├── run_pipeline.py / config.py
├── inventory/       # hashes + verified test/page mappings
├── schemas/         # Pydantic + generated JSON Schema
├── extractors/      # answer spans, OCR regions, crop/transforms
├── structure/       # grouping across pages + LLM adapter
├── audio/           # ASR, anchors, alignment/timeline mapping
├── validation/      # source/content/answer/relations/assets/audio
├── review/          # evidence renderer + correction handling
├── export/          # draft/release packaging
├── tests/           # reviewed fixtures + failure cases
└── data/            # ignored raw/cache/draft/review/release
```

## 10. Triển khai và nghiệm thu

### A. Khóa nguồn và schema

Inventory toàn bộ 10 test, duyệt source mapping, xây schema/evidence/correction format và fixture thật đã đối chiếu. Chưa gọi LLM hàng loạt.

### B. Pilot đủ 7 part

Test 02 là pilot đầy đủ, thêm mẫu test khác để phủ các font/bố cục sách giải. Mẫu phải có directions/example Part 1, Part 2 ba lựa chọn, LC graphic, hai cột RC, bảng, nhóm nhiều trang, single/double/triple passages, Part 6 sentence blank, Part 7 insertion target. Thiếu dạng nào trong Test 02 thì lấy đúng test khác và giữ ID/provenance riêng.

Duyệt đủ 200 câu, 54 item nghe, passages và media; đo lỗi/runtime/VRAM/chi phí thật. Sửa lỗi hệ thống trước khi khóa engine/model và mở rộng.

### C. Kiểm thử lỗi dữ liệu

Expected fixtures được người xem nguồn xác nhận, không lấy output extractor làm đáp án test. Kiểm tra: cột test không theo text order, hàng ngắt trang, nhiều font đậm, mất option, duplicate question, ghép nhầm test, OCR sai số/negation, blank sai câu, bảng lệch hàng, offset audio sai, ASR bỏ câu, stale cache, LLM truncated và sửa sau review. Mỗi trường hợp phải tạo lỗi/issue đúng và không lọt release.

### D. Mở rộng 10 test

Chạy từng test có checkpoint, vẫn review toàn bộ như pilot. Release gồm payload, asset manifest/checksums, validation report và review audit. Build staging rồi chuyển atomic sang release khi đủ điều kiện; không ghi đè release đã duyệt bằng draft.

### E. Tích hợp Engflex sau pipeline

Thiết kế importer idempotent theo source identity/hash; model/DTO/repository giữ quan hệ lồng nhau, contracts sinh từ Go theo AGENTS.md. Không ép TOEIC vào lesson config hiện tại. Tách payload làm bài khỏi correctKey, transcript/dịch và lời giải; chỉ cung cấp hỗ trợ đúng chế độ/thời điểm. Khi thay DB phải kiểm tra importer/query trên Postgres thật.

## 11. CLI dự kiến

Đây là acceptance contract cho CLI sẽ xây, chạy từ repository root. Dry-run chỉ inventory/cache/config/plan, không inference hoặc API; muốn thử OCR/ASR dùng stage tương ứng.

```powershell
uv sync --project scripts/toeic_pipeline --frozen
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py inventory
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py run --test 2 --dry-run
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py run --test 2 --stage extract
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py run --test 2
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py review --test 2
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py validate --test 2
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py release --test 2
```

Run chỉ tạo draft. Review tạo gói bằng chứng/nhập quyết định, không tự phê duyệt. Release không bypass validation/review. Stage lỗi hoặc release chưa đạt trả exit code khác 0, giữ artifacts và chỉ rõ report để sửa tiếp.

## 12. Implementation 0.1 và vận hành pilot

Source nằm trực tiếp trong thư mục này: `schemas.py`, `extractors.py`, `structure.py`, `llm.py`, `audio.py`, `book_tables.py`, `media.py`, `validation.py`, `corrections.py`, `pipeline_io.py`, `run_pipeline.py`. `schema.json` được sinh từ Pydantic; cập nhật bằng command `schema`.

Đã khóa Python 3.13 và dependencies qua uv.lock. PyAV khóa dưới 16 vì thử nghiệm thực tế cho thấy PyAV 19 không còn argument mà faster-whisper 1.2.1 sử dụng. Baseline ASR là small.en/CPU/int8; kết quả là ứng viên cần nghe lại.

```powershell
uv sync --project scripts/toeic_pipeline --frozen
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py run --test 2
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py review --test 2
uv run --project scripts/toeic_pipeline python scripts/toeic_pipeline/run_pipeline.py validate --test 2
uv run --project scripts/toeic_pipeline pytest scripts/toeic_pipeline/tests -q
uv run --project scripts/toeic_pipeline ruff check scripts/toeic_pipeline
```

`--stage extract` chạy inventory/PDF/OCR không gọi LLM/ASR. `--no-audio` bỏ bước ASR trong run và vẫn để validation báo thiếu audio. Chưa cung cấp `--skip-audio`; khi chạy bình thường ASR tự dùng cache đúng source/config. `--dry-run` không inference/API, nhưng vẫn kiểm kê hash nguồn và ghi manifest.

`--env-file PATH` hoặc mặc định `apps/api/.env` chỉ nạp DEEPSEEK_API_KEY, DEEPSEEK_BASE_URL, DEEPSEEK_MODEL; không log secrets. Tên mô tả DeepSeek-V4.1-Flash-1017 được map sang API ID deepseek-flash, không sửa .env gốc. Full run có gửi OCR và sách giải tới API đã cấu hình; không dùng full run cho tài liệu khác nếu chưa được phép gửi ra dịch vụ đó. Part 7 chia request theo nhóm để tránh output truncation; câu thiếu/trùng/sai phạm vi bị từ chối.

Artifacts trong `data/test_02/`: draft.json, sources.json, answer_candidates.json, mapping_candidates.json, raw/, pages/, crops/, validation.json và review.html. Mở review.html cục bộ để xem JSON, trang đề và nghe MP3 nguồn. Đây là viewer bằng chứng, chưa phải giao diện ký duyệt có tài khoản.

Ảnh Part 1 Test 02 có ROI fixture gắn SHA-256 nguồn sau khi xem trang gốc. Câu 3 có nét ghi chú đè lên ảnh; giữ nguyên bằng chứng, không tẩy hoặc đoán phục dựng. Crop và graphic khác vẫn phải kiểm tra thủ công. OCR baseline chạy hai nửa trang; chưa có engine độc lập thứ hai, nên không coi output là OCR consensus. Bảng/inline layout còn unresolved sẽ được báo lỗi thay vì tự chứng nhận.

RC solutions lấy trực tiếp table cells và continuation qua trang; `raw/rc_solutions.json` giữ sourceRefs cho explanation/translation/vocabulary. Vocabulary còn ở dạng raw source text, chưa chuẩn hóa POS/word entries. LLM có thể để null cho nội dung không lấy chắc chắn; validator yêu cầu review null reason. LC book-answer cross-check cần xác nhận độc lập, không coi đáp án LLM sao chép từ key là nguồn thứ hai.

Correction có thể nhập bằng command `correct --test 2 --corrections FILE`. File JSON gồm payloadHash và changes; mỗi change cần op=replace, JSON Pointer path, expectedValueHash, value, reviewer, reason. Hash là SHA-256 của canonical JSON theo pipeline_io.digest. Correction stale bị từ chối và overlay được lưu riêng; review cũ tự mất hiệu lực khi payload hash đổi.

Review decisions nhập qua `data/test_02/reviews.json`, gồm payloadHash và decisions. Mỗi decision có code, target đúng validation report, decision=verified, reviewer, reviewedAt, reason. Chỉ issue yêu cầu review mới được đóng bằng hai reviewer khác tên; lỗi cấu trúc/nội dung thiếu phải sửa dữ liệu. Đây là audit khai báo cục bộ, chưa có xác thực danh tính/chữ ký; không dùng thay quy trình review độc lập thực tế. Không tự sinh quyết định xác nhận thay người duyệt.

Validate/release trả exit code 2 nếu chưa đủ điều kiện. Run trả 0 khi xử lý kỹ thuật hoàn tất nhưng draft còn issue review; trả 1 khi stage lỗi. Luôn đọc validation.json, không suy releaseReady từ exit code run.

### Kết quả pilot Test 02

Chạy toàn luồng với `run --test 2 --repair`: OCR 42 trang, 200 câu/đáp án ứng viên, 54 item có transcript/audio range, 100 lời giải RC và 16 blank nodes. `--repair` chạy OCR toàn trang cho nhóm có layout/blank/số passage chưa đạt, rồi cấu trúc hóa lại bằng nguồn; raw và cache cũ được giữ để đối chiếu. Đây là lượt OCR khác bố cục/độ phân giải, không phải engine OCR độc lập.

33 kiểm thử unit/integration pass; Ruff pass. Các kiểm thử artifact kiểm tra coverage, một số ô đáp án đã xem nguồn, audio ranges, blanks, double/triple passages và release gate. Chúng không thay việc đọc/nghe toàn bộ để xác nhận chính xác nội dung. Release bị chặn đúng khi chưa có review; chưa phát hành bộ đề.

Kết quả và giới hạn cụ thể nằm trong `data/test_02/pilot_report.json`, lỗi/review còn mở trong `validation.json`, dữ liệu trong `draft.json`, viewer trong `review.html`. Khi chạy lại hãy dùng `--repair` để tái sử dụng cả cache sửa lỗi; chạy không có flag này sẽ tạo lại draft từ lượt cấu trúc hóa đầu tiên và báo lại các lỗi đã phát hiện.

## 13. Lưu trữ một bảng đề và mở rộng attempts — phương án được chọn

Đặc tả chi tiết hiện hành: [Thiết kế DB đề thi và lượt làm bài](database/README.md). Phương án đã chốt chỉ thêm `exam_tests`, mở rộng `attempts`; thay thế đề xuất tách bảng version/part/item/question/answer trước đó. Đây là thiết kế cho bước tích hợp sau; chưa tạo migration, model Go hoặc ghi vào DB.

| Đơn vị dự kiến | Cột quan hệ/tra cứu | Nội dung phức tạp |
| --- | --- | --- |
| exam_tests mới | id, source_key UNIQUE, slug UNIQUE, title, type, timestamps | content JSONB giữ parts/items/passages/questions và media references; không có version/status/content_hash/evidence_key |
| attempts hiện có | User, exam_test_id, mode, selected_parts, status, score tổng, timestamps/deadline | result JSONB giữ answers, kết quả từng part/section và summary; không lặp totalScore |

Các closed set cấp row dùng Go enum + column CHECK theo AGENTS.md; cấu trúc JSONB dùng typed parser. Mỗi row exam_tests là một đề có sẵn, chống import trùng bằng UNIQUE(source_key). FK attempts.exam_test_id trỏ đúng đề; không ghi đè nội dung dùng để chấm bài khi đề đã có attempt. Không quản lý version/trạng thái xuất bản và không tạo bảng câu trả lời.

Ở runtime, Attempt mang relation ExamTest read-only; có thể đọc belongs-to bằng Joins. Cây JSONB giữ nguyên quan hệ lồng nhau để controller map sang DTO, không regroup passage/questions trong service/controller. Mỗi passage/transcript chỉ nằm một lần ở item. Listing/history chỉ lấy metadata cần thiết thay vì kéo content của mọi đề.

### 13.1. Export bằng allowlist

`content_export.py` định nghĩa schema nội dung riêng, chỉ xuất field đã khai báo. Thêm metadata vào extractor không tự làm field đó xuất hiện trong content. Không dùng thao tác xóa vài key khỏi draft làm hợp đồng import.

- Giữ ID/order/quan hệ, câu hỏi/options/đáp án/lời giải, transcript/dịch, bảng, inline blank/insertion/marked nodes, target IDs, audio ranges và asset references.
- Giữ nguyên null và mảng rỗng có ý nghĩa; không tự sửa nội dung hoặc flatten quan hệ để giảm dung lượng.
- Loại sourceRefs, answerRefs, bbox/font/confidence, metadata từng trang, raw OCR/ASR/LLM và review audit khỏi content.
- Asset manifest chỉ có sourceId/key và checksum của asset thực sự được tham chiếu; loại PDF sách giải, nguồn đề và asset không dùng. Asset thiếu hoặc trùng identity phải lỗi.
- `provenance.json` giữ draft hash, content hash và source manifest hash ở artifact của pipeline; bằng chứng đầy đủ vẫn nằm trong draft/raw/cache/review. Không import các hash hoặc liên kết evidence thành cột exam_tests; runtime không cần truy lại bằng chứng để làm/chấm bài.
- `content.json` là tài liệu import phía server **có đáp án**, không phải DTO đưa trực tiếp cho thí sinh. API làm bài phải có allowlist riêng theo chế độ/thời điểm để ẩn correctKey, explanation, transcript/dịch khi chưa được phép.

```powershell
# Xuất preview từ draft hiện tại, không OCR/ASR/API và không ghi DB
uv run --frozen --project scripts/toeic_pipeline python scripts/toeic_pipeline/content_export.py --test 2
```

Output trong `data/test_02/preview/`: content.json dễ đọc, content.compact.json cùng nội dung để truyền/đo dung lượng và provenance.json có trạng thái preview-needs-review. Giảm whitespace không phải tối ưu schema; dung lượng PostgreSQL thực tế cần đo sau import và tính cả index/overhead.

Lệnh release hiện có chỉ gọi exporter sau khi cổng validation/review pass; output content đã lọc nằm trong release/. Preview không phải release và không cho phép bỏ qua cổng duyệt. Exporter không ghi DB hay upload media.

Release/review là bước kiểm tra đầu ra của pipeline trước import; không có nghĩa cần cột status hoặc quy trình publish/version ở DB ứng dụng. JSON schemaVersion là định dạng hợp đồng dữ liệu, không phải version đề.
