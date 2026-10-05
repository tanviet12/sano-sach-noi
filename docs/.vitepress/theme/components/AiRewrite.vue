<script setup lang="ts">
// Trang chủ: ba cách đọc + tặng kèm skill làm sách nói (anh Việt duyệt 27/09).
// Nghe mẫu dùng chung 3 đoạn với app (docs/public/audio/cach-doc-cap-*.mp3).
import { ref } from 'vue'
import { withBase } from 'vitepress'
import { Download, Github, Pause, Play, Sparkles } from 'lucide-vue-next'

const levels = [
  { n: 1, title: 'Đọc nguyên văn', desc: 'Sano đọc đúng từng chữ trong file Word, EPUB, PDF.' },
  { n: 2, title: 'Làm mượt', desc: 'AI đổi bảng, hình, danh sách, chữ viết tắt thành lời. Giữ nguyên ý.' },
  { n: 3, title: 'Viết lại thành văn sách nói', desc: 'AI viết lại như người kể: chuyện trước, lý thuyết sau, cuối chương có ba ý cần nhớ.' },
]
const playing = ref(0)
let audio: HTMLAudioElement | null = null
function toggle(n: number) {
  const on = playing.value === n
  audio?.pause()
  audio = null
  playing.value = 0
  if (on) return
  audio = new Audio(withBase(`/audio/cach-doc-cap-${n}.mp3`))
  audio.onended = () => (playing.value = 0)
  playing.value = n
  void audio.play().catch(() => (playing.value = 0))
}
</script>

<template>
  <section id="ai-viet-lai" class="sano-section">
    <div class="sano-container grid">
      <div>
        <h2 class="sano-h2">Nghe cuốn hơn nhờ AI viết lại</h2>
        <p class="sano-lead">Văn viết để đọc bằng mắt, đọc to lên thường nghe chán. Chọn một trong ba cách đọc, nhờ Claude, ChatGPT hoặc Gemini bản miễn phí. Nghe cùng một đoạn, ba cách viết:</p>
        <ul class="levels">
          <li v-for="l in levels" :key="l.n" class="sano-card level">
            <button class="play" :class="{ on: playing === l.n }" :aria-label="`${playing === l.n ? 'Dừng' : 'Nghe mẫu'} cấp ${l.n}`" @click="toggle(l.n)">
              <component :is="playing === l.n ? Pause : Play" :size="16" aria-hidden="true" />
            </button>
            <div>
              <p class="t"><span class="sano-muted n">Cấp {{ l.n }}</span> {{ l.title }}</p>
              <p class="d sano-muted">{{ l.desc }}</p>
            </div>
          </li>
        </ul>
        <p class="more"><a :href="withBase('/lam-muot-tai-lieu')">Xem cách làm từng cấp →</a></p>
      </div>

      <div class="sano-card gift">
        <p class="tag"><Sparkles :size="14" aria-hidden="true" /> Tặng kèm miễn phí</p>
        <h3 class="gt">Skill làm sách nói cho Claude và ChatGPT</h3>
        <p class="sano-muted gd">Nạp một lần. Lần sau chỉ cần đính kèm file Word hoặc PDF và gõ một câu, AI trả về file Word đã viết lại, có sẵn chương, mục để nạp vào Sano.</p>
        <p class="say">"Làm file sách nói dùng skill sano-sach-noi (cấp độ 3)"</p>
        <div class="btns">
          <a class="sano-btn brand" :href="withBase('/skill/sano-sach-noi.zip')" download><Download :size="16" aria-hidden="true" /> Tải skill</a>
          <a class="sano-btn outline" href="https://github.com/tanviet12/sano-sach-noi/tree/main/skills" target="_blank" rel="noopener"><Github :size="16" aria-hidden="true" /> Xem trên GitHub</a>
        </div>
        <p class="sano-muted note">Dùng được với Claude mọi gói, ChatGPT (qua Dự án hoặc mục Skills) và Claude Code. <a :href="withBase('/lam-muot-tai-lieu#nap-skill-cho-ai-lam-mot-lan-dung-mai')">Cách nạp</a></p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.grid {
  display: grid;
  gap: 32px;
  align-items: start;
}
@media (min-width: 960px) {
  .grid {
    grid-template-columns: 1.2fr 1fr;
    gap: 48px;
  }
}
.levels {
  margin: 24px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 12px;
}
.level {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  box-shadow: none;
}
.play {
  flex: none;
  width: 36px;
  height: 36px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  cursor: pointer;
}
.play.on {
  background: hsl(var(--sano-primary));
  border-color: hsl(var(--sano-primary));
  color: #fff;
}
.t {
  margin: 0;
  font-weight: 500;
}
.n {
  font-size: 12px;
  font-weight: 400;
  margin-right: 4px;
}
.d {
  margin: 2px 0 0;
  font-size: 14px;
  line-height: 1.6;
}
.more {
  margin: 16px 0 0;
  font-size: 14px;
}
.more a,
.note a {
  color: var(--vp-c-brand-1);
}
.gift {
  padding: 28px;
  background: linear-gradient(160deg, hsl(172 60% 96%), var(--vp-c-bg) 70%);
}
.dark .gift {
  background: linear-gradient(160deg, hsl(172 40% 14%), var(--vp-c-bg) 70%);
}
.tag {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: hsl(172 66% 30%);
  background: hsl(172 60% 90%);
}
.dark .tag {
  color: hsl(172 60% 70%);
  background: hsl(172 40% 20%);
}
.gt {
  margin: 14px 0 0;
  font-size: 22px;
  line-height: 1.3;
  font-weight: 600;
}
.gd {
  margin: 8px 0 0;
  font-size: 15px;
  line-height: 1.6;
}
.say {
  margin: 16px 0 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: hsl(var(--sano-muted));
  font-size: 14px;
  font-weight: 500;
}
.btns {
  margin-top: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.btns .sano-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.note {
  margin: 14px 0 0;
  font-size: 13px;
  line-height: 1.6;
}
</style>
