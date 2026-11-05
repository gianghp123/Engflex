# Kế hoạch tách module Learning và Learner

Đợt này bắt đầu bằng việc tổ chức lại code Go thành hai module nghiệp vụ trong `apps/api/internal/modules/`. Hai module tiếp tục chạy trong **một Go API, dùng PostgreSQL hiện tại**. Đây là đợt tách module trong code. Deploy riêng sẽ được quyết định sau. Billing, Conversation, Vocabulary, Voice và Pronunciation nằm ngoài phạm vi đợt này.

| Module mới | Code chuyển vào | Trách nhiệm |
|---|---|---|
| **Learning** | `lessons`, `videos`, `exams`, `stories` | Nội dung học, bài tập, tạo và cập nhật attempt, chấm điểm. |
| **Learner** | `profiles`, `bookmarks`, `users`, phần progress trong `attempts` | Hồ sơ, bookmark, user directory và tổng hợp tiến độ. |

Theo [ownership manifest](../../../apps/api/ownership.json), dữ liệu `Attempt` thuộc Learning. Package `attempts` hiện đăng ký route progress và nằm phía Learner. Learner đọc kết quả làm bài qua interface để tổng hợp progress.

## Phase 1 — Tạo và chuyển module Learning

- [ ] **1.1 — Tạo `internal/modules/learning/` và chuyển lessons.** Giữ lessons là package con với controllers, services và DTOs hiện có. Cập nhật imports và wiring trong composition root ngay khi chuyển. Giữ nguyên đường dẫn API, middleware và response.
- [ ] **1.2 — Chuyển videos, exams và stories lần lượt.** Sau mỗi lần chuyển, build và chạy test của package vừa chuyển trước khi tiếp tục.
- [ ] **1.3 — Kiểm tra các luồng ghi của Learning.** Thử mở nội dung free/premium, tạo và cập nhật attempt, nộp bài và chấm điểm. Giữ các transaction, locking, video revision và TOEIC deadline hiện có.

**Hoàn thành khi** bốn nhóm nội dung chạy từ cây package `learning/` qua các route hiện tại.

## Phase 2 — Tạo và chuyển module Learner

- [ ] **2.1 — Tạo `internal/modules/learner/` và chuyển profiles, bookmarks, users.** Giữ từng package nghiệp vụ riêng và cập nhật wiring sau mỗi lần chuyển.
- [ ] **2.2 — Chuyển package progress.** Đưa package `attempts` hiện đăng ký progress vào cây `learner/`. Giữ route `GET /api/v1/progress` và tiếp tục đọc qua `LearningProgressReader`, `ConversationProgressReader` và `ProfileReader` hiện có.
- [ ] **2.3 — Kiểm tra ranh giới.** Giữ thao tác ghi attempt ở Learning. Learner đọc số liệu cần thiết qua interface, gồm cả luồng admin xem attempt. Composition root tiếp tục nối dependencies. Giữ hành vi Billing đang cung cấp cho các luồng hiện tại.

**Hoàn thành khi** profile, bookmarks, user directory và progress chạy từ cây package `learner/`, còn các luồng ghi attempt vẫn chạy qua Learning.

## Phase 3 — Dọn wiring và nghiệm thu

- [ ] **3.1 — Đồng bộ đường dẫn package.** Cập nhật ownership manifest, cấu hình sinh contracts và các cấu hình khác có tham chiếu đến package đã chuyển. Kiểm tra import, route registration và đường dẫn cũ trước khi dọn.
- [ ] **3.2 — Kiểm tra Go và contracts.** Chạy Go tests, `go vet` và các check trong CI. Sinh lại contracts nếu cấu hình đầu vào đổi đường dẫn và kiểm tra public DTOs vẫn giữ nguyên.
- [ ] **3.3 — Chạy thử qua HTTP local.** Kiểm tra readiness, [smoke probes](../../../infrastructure/scripts/smoke.py), mở nội dung, lưu/nộp attempt, profile, bookmark và progress. Dùng schema và dữ liệu kiểm thử hiện có.

**Hoàn thành khi** Go API chạy với hai cây module mới, các route và contracts hiện tại vẫn dùng được, và hai nhóm có ranh giới đọc/ghi rõ ràng.
