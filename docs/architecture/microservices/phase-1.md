# Phase 1 — Chuẩn bị vận hành nhiều service

Nghiệm thu trong môi trường local cô lập. [Roadmap](README.md) quy định phạm vi. File này lưu kết quả rà soát và checklist nghiệm thu.

## Bản đồ dependency

Đối chiếu [composition root](../../../apps/api/internal/composition/root.go) và [ownership manifest](../../../apps/api/ownership.json). Số lần gọi dưới đây là số lần đọc theo đường chạy được mô tả. Ngân sách độ trễ của cả ba luồng cần đo trước khi chuyển luồng qua HTTP.

| Luồng | Caller và owner | Dữ liệu và số lần đọc hiện tại | Hành vi cần giữ khi tách |
|---|---|---|---|
| Premium trên protected action | Lessons, Videos, Vocabulary và Scenarios (gồm policy cho Conversation) gọi `common.PremiumChecker`. Billing sở hữu subscriptions. | `SubscriptionService.IsPremium` đọc subscription active chưa hết hạn một lần cho mỗi lần kiểm tra premium. Một action kiểm tra cả deck lẫn entry có thể gọi nhiều hơn một lần. | `403` khi thiếu quyền, `503` khi lookup lỗi. Kiểm tra theo từng action và giữ các lần kiểm tra lại hiện có. |
| Free content và admin preview | Cùng policy với protected action. Quyền preview được caller xác định. | `RequirePremiumAccess` bỏ qua Billing khi nội dung free hoặc caller có quyền admin. Không có lookup entitlement tại bước này. | Billing outage không chặn các đường đi vốn bỏ qua lookup. |
| Admin directory | Learner `UserService` gọi `subscriptionLookup.Enrich`. `UserDirectoryRepository` đọc dữ liệu thuộc Billing. | Một truy vấn subscription cho cả trang Clerk hoặc user vừa cập nhật. Batch rỗng bỏ qua truy vấn. | Giữ batch và subscription đang hiệu lực. Lỗi lookup được trả về thay vì hiển thị mọi user là không premium. |
| Learner context khi provision Conversation | `ConversationService.Start` gọi `promptInputs` rồi gửi payload qua HTTP cho Voice. Learner sở hữu hồ sơ. | Không có lần đọc profile. `promptInputs` đọc scenario và persona khi có ID tương ứng. Payload chưa điền `VoiceSessionBody.Learner`. | Giữ hành vi hiện tại trong Phase 1. Scoped learner context là dependency dự kiến nếu Conversation được tách. Cần chốt trường dữ liệu và hành vi khi Learner unavailable trước khi đưa vào đường provision. |
| Dashboard progress | Dashboard gọi `useProgress` → `GET /api/v1/progress` → Learner `ProgressService`. | Khi thành công có năm lần đọc tuần tự. Learning cung cấp ngày có attempt hoàn thành và số attempt hoàn thành. Conversation cung cấp ngày kết thúc và tổng thời lượng hôm nay. Learner cung cấp daily goal. | Lỗi nguồn dữ liệu làm rollup trả lỗi. Hồ sơ chưa tồn tại hoặc chưa có goal dương dùng 15 phút. Giữ ranh giới ngày UTC. Giao diện chỉ hiển thị hàng thống kê khi pending hoặc thành công. |

### Điểm nối để kiểm chứng

- [Policy premium](../../../apps/api/internal/common/premium.go), [Billing checker](../../../apps/api/internal/modules/payments/services/subscription_service.go) và [batch directory enrichment](../../../apps/api/internal/database/repositories/user_directory_repository.go). Khi tách Billing, kiểm chứng entitlement theo subject và giữ batch cho admin directory.
- [Conversation provisioning](../../../apps/api/internal/modules/conversations/services/conversation_service.go) và [Voice client](../../../apps/api/internal/voice/client.go). Nếu tách Conversation, đo provisioning và callbacks trước khi đổi routing.
- [Dashboard](../../../apps/web/src/routes/_app/home.tsx), [query](../../../apps/web/src/features/attempts/queries.ts) và [progress rollup](../../../apps/api/internal/modules/attempts/services/progress_service.go). Nếu cần tách reporting, đo thời gian tổng và từng nguồn trước khi chọn cách đọc dữ liệu.

## Nghiệm thu theo thứ tự roadmap

- [ ] Kiểm tra mở nội dung, lưu hoặc nộp attempt, đọc profile và xem progress trên local. Ghi số đo ban đầu và kiểm tra log, readiness, smoke.
- [ ] Khi triển khai Billing ở Phase 6, kiểm tra contract, timeout, outage, smoke và rollback trên local.
