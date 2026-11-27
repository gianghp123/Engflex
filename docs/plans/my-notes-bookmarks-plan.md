# Hoàn thành My Notes

My Notes là nơi xem nội dung đã bookmark, gồm từ vựng, bài học, video và truyện. Người dùng lưu từ màn nội dung và có thể thêm ghi chú cá nhân vào mục đã lưu. Từ vựng tiếp tục dùng `user_vocabulary`. Bài học, video và truyện dùng `bookmarks`.

## 1. Backend và contracts

- Chạy GitNexus impact trước khi sửa symbol hiện có. Xác định caller của API bookmark, query Story và thao tác cập nhật từ vựng. Báo trước khi impact là HIGH hoặc CRITICAL.
- Thêm cột `note` nullable vào migration `20261005193714_create_bookmarks.sql` theo quy ước một migration cho mỗi bảng. Giữ khóa chính `(user_id, content_type, content_id)`, giới hạn loại bằng enum Go và CHECK hiện có. Áp dụng lại chain trên PostgreSQL dev theo AGENTS.md.
- Giữ GET, POST và DELETE bookmark hiện có cho các caller như Story. Khi tạo bookmark, kiểm tra nội dung đích tồn tại và áp dụng cùng quy tắc hiển thị, quyền truy cập với màn chi tiết. Lưu lặp phải trả bookmark đã có và giữ ghi chú cũ. Cho phép chủ sở hữu bỏ lưu cả mục có nội dung đích đã bị gỡ.
- Thêm PATCH cập nhật riêng trường ghi chú của bookmark theo loại và ID. Thêm đường cập nhật riêng `user_vocabulary.note` theo item ID để giữ nguồn, mastery và lịch ôn. Cả hai chỉ cập nhật hàng thuộc Clerk subject hiện tại. Chuẩn hóa ghi chú rỗng thành null và khai báo giới hạn độ dài dùng chung giữa request và web.
- Bổ sung đường đọc cho My Notes theo tab. Danh sách bookmark trả quan hệ lồng `Lesson`, `VideoExercise` hoặc `Story` với metadata tối thiểu để dựng thẻ và điều hướng. Danh sách từ đã lưu trả quan hệ `VocabularyItem` và ghi chú có sẵn. Dùng quan hệ GORM read-only, repo trả model có quan hệ đã điền, DTO phản chiếu model và controller map một lần. Dùng JOIN phù hợp hoặc số query cố định. Xử lý nội dung mất quyền hoặc đã gỡ mà vẫn cho chủ sở hữu bỏ lưu, đồng thời tránh trả metadata bị giới hạn.
- Tab Tất cả cần phân trang và sắp xếp ổn định theo ngày lưu trên cả hai nguồn. Chốt truy vấn hợp nhất khóa trước khi triển khai, rồi nạp model và các quan hệ theo lô tại repo. Đếm tổng, số có ghi chú và số theo loại trên tập dữ liệu đầy đủ theo phạm vi của từng chỉ số. Kiểm thử dữ liệu lớn hơn một trang và các bản ghi có cùng ngày lưu.
- Giữ Swagger, response envelope, validation và service logging theo mẫu dự án. Chạy `make contracts`, `make mocks` khi đổi interface và `make swagger` từ `apps/api`. Sinh contract từ Go DTO, giữ enum dạng union.

## 2. Frontend

- Nối My Notes với TanStack Query. Đặt query key theo Clerk user và tham số lọc. Dùng các API mutation riêng cho sửa ghi chú và bỏ lưu. Sau mutation, đồng bộ danh sách, thống kê và trạng thái bookmark trên màn nội dung. Dùng chung query và invalidation với Story và Vocabulary.
- Nối Story và Vocabulary trước để kiểm tra dữ liệu thật đã có. Sau đó thêm nút lưu/bỏ lưu ở màn chi tiết bài học và các màn video dùng chung ID bài video. Trạng thái nút phản ánh dữ liệu đã lưu của tài khoản hiện tại.
- Điều hướng bằng `APP_ROUTES` và params thật. Dùng slug cho Story và Lesson, item ID cho Vocabulary, video ID cho route luyện video. Tập trung API paths ở `API_ROUTES`.
- Đưa tab, từ khóa và trang vào URL. Đổi tab hoặc từ khóa thì về trang đầu. Dùng `Pager`, `StatCard`, `CardShell`, `SubmitButton` và trạng thái riêng cho sổ tay rỗng, không có kết quả lọc, lỗi tải và mục mất quyền truy cập. Giữ bản nháp khi lưu lỗi, chỉ đóng editor sau khi lưu thành công. Dùng Paraglide với key tĩnh và đồng bộ EN/VI.
- Sau khi cả bốn loại hoạt động với API thật, xóa `features/notes/fixtures.ts`, các thao tác thêm/sửa/xóa trên store demo và `NoteCreateDialog`. Dọn types, imports, message IDs chỉ dùng cho mock và link về trang danh mục của fixture. Xác nhận My Notes khởi đầu rỗng với tài khoản mới và không còn đường dẫn nào nạp fixture. Giữ mocks do mockery sinh để chạy unit tests.

## 3. Nghiệm thu

- Kiểm tra Go với `gofmt`, `go vet ./...`, `go test ./... -count=1`. Chạy truy vấn và migration trên PostgreSQL thật, kiểm tra hàng trả về và các trường hợp JOIN không có quan hệ.
- Chạy `pnpm compile:messages`, `pnpm check`, `pnpm exec tsc --noEmit`, `pnpm check:conventions` và `pnpm build` từ web. Probe HTTP/SSR khi đăng xuất và khi đã đăng nhập.
- Với từng loại, lưu từ màn nội dung, mở My Notes, sửa ghi chú, tải lại, mở đúng đích, xóa ghi chú và cuối cùng bỏ lưu. Kiểm tra thống kê, tìm kiếm, phân trang và trạng thái hai màn sau mỗi thay đổi.
- Thử tài khoản thứ hai, bookmark trùng, UUID lỗi, nội dung không tồn tại hoặc bị gỡ, mất quyền truy cập và lỗi mạng. Xác nhận cập nhật ghi chú từ vựng giữ nguyên toàn bộ tiến độ và lịch ôn. Xác nhận Story Saved vẫn dùng được sau khi My Notes chuyển sang API thật.
