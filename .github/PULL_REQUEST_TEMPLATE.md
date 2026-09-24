# Pull Request

## Tóm tắt thay đổi

<!-- 1-3 câu: feat/fix/refactor gì, module nào, lý do. -->

## Loại commit

- [ ] `feat(module): ...` — feature mới
- [ ] `fix(module): ...` — bug fix
- [ ] `refactor(module): ...` — không đổi behavior
- [ ] `chore: ...` — config / dep / gitignore
- [ ] `docs: ...` — chỉ doc
- [ ] `test: ...` — chỉ test

## Checklist

### Build + test
- [ ] `go vet ./...` + `go test ./internal/...` pass
- [ ] `cd web && npx vue-tsc -b && npm run build` pass (nếu sửa FE)
- [ ] Smoke test liên quan pass (`make smoke-all`)

### Quyền truy cập (nếu thêm endpoint mới)
- [ ] Endpoint admin nằm dưới `/admin/api/*` (middleware chỉ cho admin)
- [ ] Endpoint người nghe kiểm tra quyền truy cập sách (public / private + gán sách)

### Giao diện mới
- [ ] Bám `docs/design-system.md` (token màu, spacing, component)
- [ ] Dark mode, trạng thái rỗng / đang tải / lỗi đầy đủ

### Migration
- [ ] Có cặp `.up.sql` + `.down.sql`, đã thử up → down → up

### Bảo mật
- [ ] Không hardcode secret / API key
- [ ] Endpoint đăng nhập mới có chống dò mật khẩu (khoá IP sau 5 lần sai)

## Test plan

- [ ] ...

## Screenshot / video (nếu là UI)
