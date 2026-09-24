# Sano Design System

> Baseline giao diện phần mềm Sano (`desktop/frontend`). Mọi wireframe + màn mới bám doc này.
> **KHÔNG tự thiết kế lại tokens, không inline màu/font ngoài design system.**

---

## 0. Tinh thần

- Gọn, rõ, yên tĩnh — người dùng đến để làm sách và nghe, không để ngắm giao diện.
- Accent **đỏ thương hiệu** `#c60505` (xem [`BRAND.md`](../BRAND.md)), còn lại là nền trung tính slate.
- Sáng / tối theo hệ điều hành, polish ngang nhau, không invert thô.
- Mọi action có `focus-visible:ring-2`, mọi nút chỉ có icon có `aria-label` hoặc `title`.
- Motion: transition 150–200ms ease-out, không bounce.

**KHÔNG chấp nhận:** empty state "Không có dữ liệu" trống trải · lỗi bằng `alert()` hay "Có lỗi xảy ra" không nói gì · việc chạy lâu (cài bộ đọc, render, xuất M4B) không có tiến độ và không huỷ được.

---

## 1. Design tokens

### 1.1. Màu (HSL, qua CSS variable)

Định nghĩa trong `desktop/frontend/src/style.css`, map sang Tailwind ở `desktop/frontend/tailwind.theme.ts`. Component dùng qua class: `bg-primary`, `text-foreground`, `border-border`...

| Token | Sáng | Tối | Dùng cho |
|---|---|---|---|
| `--background` | `0 0% 100%` | `222.2 47.4% 6%` | Nền cửa sổ |
| `--foreground` | `222.2 47.4% 11.2%` | `210 40% 98%` | Chữ mặc định |
| `--card` / `--popover` | `0 0% 100%` | `222.2 47.4% 8%` | Thẻ, hộp thoại |
| `--primary` | `0 94% 40%` (`#c60505`) | `0 80% 55%` | Nút chính, link, mục đang chọn |
| `--primary-foreground` | `0 0% 100%` | `0 0% 100%` | Chữ trên nền primary |
| `--secondary` / `--muted` / `--accent` | `210 40% 96.1%` | `217.2 32.6% 17.5%` | Nền phụ, hover |
| `--muted-foreground` | `215.4 16.3% 46.9%` | `215 20.2% 65.1%` | Chữ phụ, gợi ý |
| `--destructive` | `0 84.2% 60.2%` | `0 62.8% 50%` | Xoá, lỗi (lệch sắc nhẹ so với primary) |
| `--border` / `--input` | `214.3 31.8% 91.4%` | `217.2 32.6% 17.5%` | Viền, ô nhập |
| `--ring` | `0 94% 40%` | `0 80% 55%` | Focus ring |
| `--radius` | `0.5rem` | | Bo góc |

Dark mode: class `.dark` trên `<html>`, `main.ts` bật/tắt theo `prefers-color-scheme` của hệ điều hành.

### 1.2. Chữ

Font `Inter, IBM Plex Sans, system-ui, sans-serif`; bìa sách dùng chữ có chân (serif).

| Cấp | Tailwind | Dùng cho |
|---|---|---|
| Tiêu đề màn | `text-xl font-semibold tracking-tight` | "Thư viện", "Cài đặt"... |
| Tiêu đề khối | `text-base font-medium` | Tiêu đề thẻ, hộp thoại |
| Nhãn nhóm | `text-xs uppercase tracking-wider text-muted-foreground font-semibold` | Nhóm trong Cài đặt |
| Thân | `text-sm` | Mặc định |
| Phụ | `text-xs` | Gợi ý, metadata, badge |

- Tối đa `text-2xl`. Số đếm, phần trăm, thời lượng dùng `tabular-nums`.

### 1.3. Khoảng cách, bo góc, bóng

- Bội số 4px của Tailwind. Hay dùng: `gap-2` (hàng nút), `gap-3`/`gap-4` (lưới sách, trường nhập), `space-y-5` (các khối trong màn), padding màn `px-6 pt-6`.
- Bo góc: `rounded-md` (nút, ô nhập, mục thanh bên), `rounded-lg` (thẻ), `rounded-full` (badge, chip lọc), `rounded-xl` chỉ cho logo.
- Bóng: `shadow-sm` cho thẻ/nút, `shadow-xl` cho hộp thoại. Thanh bên, thanh tiêu đề phẳng.

---

## 2. Component

Nằm ở `desktop/frontend/src/components/`:

| Component | Ghi chú |
|---|---|
| `ui/button` | Variants `default \| destructive \| outline \| secondary \| ghost \| link`; sizes `default (h-9) \| sm \| lg \| icon`. Mỗi màn **một** nút chính. Đang chạy: thay icon bằng `Loader2 animate-spin` + disable |
| `ui/badge` | Variants `default \| secondary \| outline \| destructive` — trạng thái bộ đọc, nhãn nhỏ |
| `sano/BookCover.vue` | Bìa mặc định vẽ theo tên sách (màu băm từ tên, gáy sách, sóng âm, tên sách chữ có chân). Cùng thiết kế với `internal/cover` (bìa PNG ghi vào thư mục sách) |
| `TitleBar`, `AppSidebar` | Khung cửa sổ, xem mục 3 |
| `TermsDialog`, `EditBookDialog`, `UpdateDialog` | Hộp thoại: nền mờ, Esc / bấm nền để đóng, nút chính bên phải |

Helper `cn()` ở `@/lib/utils` (clsx + tailwind-merge) để gộp class.

---

## 3. Bố cục cửa sổ

```
┌──────────────────────────── TitleBar (36px) ────────────────────────────┐
│ AppSidebar (w-56)  │  main: tiêu đề màn + hành động chính bên phải      │
│  Thư viện          │        nội dung (lưới sách / các bước / trình phát) │
│  Tạo sách mới      │                                                     │
│  Cài đặt           │                                                     │
│  Giới thiệu        │                                                     │
│  ...               │                                                     │
│  phiên bản, tác giả│                                                     │
└────────────────────┴─────────────────────────────────────────────────────┘
```

- Màn cài bộ đọc (`SetupView`) và điều khoản (`TermsView`) chiếm cả cửa sổ, không có thanh bên.
- Tạo sách: thanh 6 bước ở đầu (Nạp file → Mục lục → Giọng đọc → Lời mở đầu → Nghe thử → Render).
- Cửa sổ tối thiểu ~1100×720; lưới sách tự dàn cột theo bề rộng.

---

## 4. Icon

**lucide-vue-next** — nguồn duy nhất. Mặc định `w-4 h-4` trong nút và thanh bên, `w-5 h-5` cho nút icon lớn, `w-6 h-6` cho empty state.

---

## 5. Tương tác

- Hover/màu: `transition-colors`; thanh tiến độ: `transition-all duration-300`.
- Focus ring **bắt buộc**: `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background` (hoặc utility `.focus-ring`). Chỉ `focus-visible:`, không xoá outline mà không thêm ring.
- Việc chạy lâu: thanh tiến độ + ước lượng thời gian còn lại + nút huỷ; đang render thì thanh bên hiện phần trăm.

---

## 6. Truy cập (WCAG 2.1 AA)

- Nút/link không có chữ phải có `title` hoặc `aria-label`. Mục thanh bên đang chọn: `aria-current="page"`.
- Hộp thoại: Esc đóng, focus vào nút chính.
- Tương phản: `text-muted-foreground` trên nền trắng ≈ 4.6:1.
- HTML ngữ nghĩa: `<header>`, `<nav>`, `<main>`, `<aside>`, `<button>` thay vì `<div @click>`.

---

## 7. Thư mục

| Thư mục (`desktop/frontend/src/`) | Nội dung |
|---|---|
| `components/ui/` | Primitive kiểu shadcn-vue (Button, Badge) |
| `components/sano/` | Component nghiệp vụ dùng lại (BookCover) |
| `components/` | Khung cửa sổ, hộp thoại |
| `views/`, `views/create/` | Màn, các bước tạo sách |
| `lib/` | Gọi Go (`backend.ts`), trạng thái (`store.ts`), dữ liệu giả (`mock.ts`), tiện ích |
| `wireframes/` | Wireframe đã duyệt (chỉ xem ở bản dev) |

---

## 8. Wireframe trước, code sau

1. Màn mới / đổi bố cục lớn → làm wireframe tĩnh (dữ liệu giả) trong `src/wireframes/`, dùng token + component doc này.
2. Duyệt xong mới làm màn thật trong `src/views/`, bám 1:1 wireframe.
3. So ảnh chụp wireframe (`?wireframe=desktop|library|terms` ở bản dev) với màn thật (`?screen=...`). Lệch = lỗi.
4. Cần đổi spacing/màu/bố cục → sửa wireframe trước.

---

## 9. DO / DON'T

- ✅ Dùng lại `components/ui/`, `components/sano/` trước khi tạo mới.
- ✅ Empty state có icon + tiêu đề + mô tả + nút hành động.
- ✅ Lỗi nói rõ chuyện gì, gợi ý cách sửa, có nút thử lại khi hợp lý.
- ❌ Hard-code màu hex/rgb trong template — dùng token (`bg-primary`...).
- ❌ `style="..."` inline trừ giá trị động (`width: ${pct}%`).
- ❌ Thêm thư viện icon/font khác.
- ❌ Tắt focus ring vì "thấy xấu".

---

## 10. Tham chiếu

- shadcn-vue: https://www.shadcn-vue.com/
- Tailwind v3.4: https://tailwindcss.com/docs
- Lucide icons: https://lucide.dev/icons/
- Wails v2: https://wails.io/docs/introduction
