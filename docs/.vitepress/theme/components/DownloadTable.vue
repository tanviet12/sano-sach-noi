<script setup lang="ts">
// Bảng tải thẳng file cài (trang Cài đặt). Link theo phiên bản mới nhất (xem release.ts).
import { REPO, useRelease } from './release'

const { version, links } = useRelease()
</script>

<template>
  <table>
    <thead>
      <tr><th>Máy</th><th>Tải về (bản {{ version }})</th></tr>
    </thead>
    <tbody>
      <tr>
        <td>Windows 10/11</td>
        <td>
          <a :href="links.win.url">{{ links.win.name }}</a> — bộ cài, không cần quyền quản trị<br />
          <a :href="links['win-zip'].url">{{ links['win-zip'].name }}</a> — bản portable, giải nén là chạy
        </td>
      </tr>
      <tr><td>macOS 10.13+ (Apple Silicon, Intel)</td><td><a :href="links.mac.url">{{ links.mac.name }}</a></td></tr>
      <tr><td>Linux x86_64</td><td><a :href="links.linux.url">{{ links.linux.name }}</a></td></tr>
    </tbody>
  </table>
  <p>Bấm là tải về ngay. Phiên bản cũ và file <code>SHA256SUMS</code>: <a :href="REPO + '/releases'" target="_blank" rel="noopener">trang Releases</a>.</p>
</template>
