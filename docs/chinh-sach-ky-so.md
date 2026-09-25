---
title: Chính sách ký số
description: 'Chính sách ký số của Sano: file cài nào được ký, ký ở đâu, ai duyệt, chữ ký bản cập nhật và quyền riêng tư. Ký số Windows miễn phí qua SignPath Foundation.'
---

# Chính sách ký số

Trang này nói rõ file nào của Sano được ký, ký bằng cách nào và ai chịu trách nhiệm. Bản tiếng Anh ở [cuối trang](#code-signing-policy-english).

::: info Tình trạng
Bản cài Windows **đang xin ký số miễn phí** qua chương trình dành cho phần mềm mã nguồn mở của [SignPath Foundation](https://signpath.org). Khi được duyệt, từ bản phát hành kế tiếp:

*Free code signing provided by [SignPath.io](https://about.signpath.io), certificate by [SignPath Foundation](https://signpath.org).*

Trước đó, bản cài Windows và macOS chưa có chữ ký số của hệ điều hành — xem [cách mở app lần đầu](./mo-app-lan-dau).
:::

## File nào được ký

| File | Chữ ký |
|---|---|
| `Sano-<phiên bản>-windows-amd64-setup.exe` và `Sano.exe` trong bản portable `.zip` | Chữ ký số Windows (Authenticode) qua SignPath Foundation — khi được duyệt |
| `SHA256SUMS` (mã SHA256 của mọi file cài) | Chữ ký ed25519 riêng của dự án, file `SHA256SUMS.sig`. Sano kiểm chữ ký này trước khi tự cập nhật, sai là từ chối |
| Mọi file phát hành | Attestation nguồn gốc build của GitHub (`gh attestation verify`) |

Chỉ ký file build từ mã nguồn của chính repo [tanviet12/sano-sach-noi](https://github.com/tanviet12/sano-sach-noi). Không ký phần mềm của người khác. Bộ đọc giọng nói (VieNeu-TTS) và ffmpeg không nằm trong file cài — Sano tải chúng lúc cài, đúng phiên bản ghim và kiểm SHA256.

## Ký ở đâu

- Chỉ ký trong quy trình phát hành tự động trên GitHub Actions ([`desktop-release.yml`](https://github.com/tanviet12/sano-sach-noi/blob/main/.github/workflows/desktop-release.yml)), chạy khi gắn thẻ phiên bản `vX.Y.Z`.
- **Không bao giờ ký bản build trên máy cá nhân.**
- Mỗi lượt ký Windows phải được người duyệt ký bấm duyệt trên SignPath.
- Khoá ed25519 ký `SHA256SUMS` chỉ nằm trong secret của repo và bản sao cất riêng của người bảo trì, không có trong mã nguồn.

## Vai trò

| Vai trò | Người | Việc |
|---|---|---|
| Người viết mã (Committer) | [Bùi Tấn Việt](https://github.com/tanviet12) | Được sửa mã nguồn trong repo |
| Người duyệt (Reviewer) | [Bùi Tấn Việt](https://github.com/tanviet12) | Duyệt mọi đóng góp từ bên ngoài (pull request) trước khi gộp |
| Người duyệt ký (Approver) | [Bùi Tấn Việt](https://github.com/tanviet12) | Duyệt từng lượt ký bản phát hành |

Mọi thành viên bật xác thực hai lớp cho tài khoản GitHub và SignPath.

## Quyền riêng tư

Sano không thu thập, không gửi dữ liệu người dùng đi đâu. Tài liệu, sách nói và tiến độ nghe nằm trên máy bạn. Sano chỉ kết nối mạng để:

- **Tải bộ đọc giọng nói lần đầu** (uv, Python, VieNeu-TTS, mô hình, ffmpeg từ GitHub, Hugging Face và nguồn chính thức của ffmpeg) — chỉ tải file, không gửi dữ liệu của bạn.
- **Kiểm tra bản mới** trên GitHub Releases khi mở app — chỉ hỏi số phiên bản mới nhất, tắt được trong **Cài đặt**. Bấm **Cập nhật ngay** thì tải file cài từ GitHub Releases.

Không có telemetry, không có tài khoản, không quảng cáo. Xem thêm [Điều khoản sử dụng](./dieu-khoan-su-dung#_4-du-lieu-va-quyen-rieng-tu). Gỡ Sano: [Gỡ cài đặt](./go-cai-dat).

## Báo lỗi bảo mật

Thấy file cài có dấu hiệu bị sửa, chữ ký sai hoặc lỗi bảo mật: mở [báo cáo bảo mật riêng trên GitHub](https://github.com/tanviet12/sano-sach-noi/security/advisories/new) (không đăng công khai).

---

## Code signing policy (English)

::: info Status
Windows installers are **pending approval** for free code signing through the [SignPath Foundation](https://signpath.org) open-source program. Once approved, starting with the next release:

*Free code signing provided by [SignPath.io](https://about.signpath.io), certificate by [SignPath Foundation](https://signpath.org).*
:::

**What is signed**

- Windows installer `Sano-<version>-windows-amd64-setup.exe` and `Sano.exe` inside the portable `.zip` — Authenticode signature via SignPath Foundation (once approved).
- `SHA256SUMS` — signed with the project's own ed25519 key (`SHA256SUMS.sig`). The app verifies this signature before applying any in-app update and rejects mismatches.
- All release files carry GitHub build provenance attestations.

Only artifacts built from this project's own source code in [tanviet12/sano-sach-noi](https://github.com/tanviet12/sano-sach-noi) are signed. Third-party runtime components (VieNeu-TTS, ffmpeg) are not bundled; the app downloads pinned versions and verifies their SHA256.

**Where signing happens**

- Only in the automated GitHub Actions release workflow ([`desktop-release.yml`](https://github.com/tanviet12/sano-sach-noi/blob/main/.github/workflows/desktop-release.yml)), triggered by a `vX.Y.Z` version tag.
- Local builds are never signed.
- Every Windows signing request is manually approved by the approver in SignPath.

**Team roles**

- Committers and reviewers: [Bùi Tấn Việt](https://github.com/tanviet12) — all external contributions are reviewed via pull request before merge.
- Approvers: [Bùi Tấn Việt](https://github.com/tanviet12).

All team members use multi-factor authentication for GitHub and SignPath.

**Privacy policy**

This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it. Sano collects no user data and has no telemetry. It connects to the network only to (1) download the text-to-speech engine and models on first run (files only, from GitHub, Hugging Face and official ffmpeg sources), and (2) check GitHub Releases for a newer version on start-up (version number only; can be turned off in Settings), downloading the installer from GitHub Releases when the user clicks *Update now*.

**Reporting security issues**: please use [GitHub private vulnerability reporting](https://github.com/tanviet12/sano-sach-noi/security/advisories/new).
