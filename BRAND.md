# Brand — Sano

**Màu nhấn chính: đỏ thương hiệu `#c60505`.**

| Token | Value |
|---|---|
| Hex | `#c60505` |
| RGB | `198, 5, 5` |
| HSL | `0 94% 40%` |
| Tailwind CSS var (light) | `--primary: 0 94% 40%` / `--ring: 0 94% 40%` |
| Tailwind CSS var (dark) | `--primary: 0 80% 55%` / `--ring: 0 80% 55%` |

### Usage
- CTA primary: nền `#c60505`, text trắng, hover đậm hơn (~`#a30404`), focus ring 2px `#c60505`.
- Link / active nav: text `#c60505` + indicator.
- **KHÔNG hard-code hex** trong component — luôn qua CSS var `--primary` (đã set ở `desktop/frontend/src/style.css`) để dark mode auto-switch.
- Semantic giữ nguyên: success emerald, warning amber, danger red. Vì primary đã là đỏ, dùng `--destructive` lệch sắc nhẹ để phân biệt với primary khi cần.
- Phần mềm: thanh bên trái, thẻ sách có bìa lớn, trình phát rộng. Áp brand đỏ.

## Logo

Trang giấy gấp góc (file Word) chứa sóng âm (giọng đọc): từ tài liệu thành sách nói.

| File | Dùng cho |
|---|---|
| `desktop/frontend/src/assets/logo.svg` | Logo chính (nền đỏ bo góc, 5 vạch sóng âm) — màn cài bộ đọc, điều khoản, Giới thiệu |
| `desktop/frontend/src/assets/favicon.svg` | Bản rút gọn 3 vạch dày cho cỡ nhỏ (thanh bên) |
| `desktop/build/appicon.png` | Icon ứng dụng 1024×1024 (Wails sinh icon Windows, macOS, Linux từ file này) |

Sửa logo: sửa `logo.svg`, xuất lại `appicon.png` 1024×1024 từ cùng hình, giữ `favicon.svg` cùng kiểu.

