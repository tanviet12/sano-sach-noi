---
layout: home
title: Nghe thử sách nói tạo bằng AI – Sano
titleTemplate: false
description: 'Nghe thử 5 cuốn sách nói và 12 giọng đọc tiếng Việt do Sano tạo bằng AI từ file Word, ngay trên máy tính. Miễn phí, mã nguồn mở.'
markdownStyles: false
---

<script setup>
import { withBase } from 'vitepress'
</script>

<div class="sano-home">

<section class="sano-section sano-ambient demo-hero">
  <div class="sano-container">
    <p class="demo-eyebrow">Nghe thử Sano</p>
    <h1 class="demo-h1">Sách nói tạo bằng AI, <span class="accent">ngay trên máy tính</span></h1>
    <p class="sano-lead demo-lead">5 cuốn sách mẫu, 12 giọng đọc tiếng Việt của bộ đọc mã nguồn mở VieNeu-TTS. Tất cả tạo từ file Word bằng Sano ngay trên máy tính: không cần phòng thu, không cần API key, không tốn tiền token, không gửi tài liệu lên mạng.</p>
    <div class="demo-actions">
      <a href="#tu-sach" class="sano-btn brand block-sm">Nghe ngay</a>
      <a :href="withBase('/tai-ve')" class="sano-btn outline block-sm">Tải Sano miễn phí</a>
    </div>
  </div>
</section>

<section id="tu-sach" class="sano-section">
  <div class="sano-container">
    <h2 class="sano-h2">Tủ sách nghe thử</h2>
    <p class="sano-lead">Chọn một cuốn, mỗi cuốn đọc bằng một giọng khác nhau. Bìa sách do Sano tự vẽ.</p>
    <DemoShelf />
  </div>
</section>

<section id="giong-doc" class="sano-section alt">
  <div class="sano-container">
    <h2 class="sano-h2">Một đoạn, 12 giọng</h2>
    <p class="sano-lead">Cùng một đoạn văn, bấm từng giọng để so sánh. Sano có 25 giọng, chọn được ngay khi tạo sách.</p>
    <VoiceGallery />
    <VoiceCredit />
  </div>
</section>

<section class="sano-section">
  <div class="sano-container demo-normalize">
    <div>
      <h2 class="sano-h2">Đọc tự nhiên, không đọc máy móc</h2>
      <p class="sano-lead">Số, chữ viết tắt, ký hiệu được đọc thành lời như người thật đọc. Bấm nghe bản chưa và đã chuẩn hoá để so.</p>
    </div>
    <NormalizeCompare />
  </div>
</section>

<section class="sano-section alt demo-cta">
  <div class="sano-container">
    <h2 class="sano-h2">Làm sách nói từ tài liệu của bạn</h2>
    <p class="sano-lead">Miễn phí, mã nguồn mở, không cần API key, chạy trên Windows, macOS, Linux. Nghe trên máy tính, trên điện thoại và trên ô tô qua CarPlay, Android Auto.</p>
    <div class="demo-actions center">
      <a :href="withBase('/tai-ve')" class="sano-btn brand block-sm">Tải Sano</a>
      <a :href="withBase('/cai-dat')" class="sano-btn outline block-sm">Xem hướng dẫn</a>
    </div>
  </div>
</section>

</div>
