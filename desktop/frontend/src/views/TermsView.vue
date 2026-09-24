<script setup lang="ts">
// Màn đồng ý điều khoản (wireframe D3): hiện sau khi cài xong bộ đọc, trước khi vào app;
// điều khoản đổi phiên bản thì hiện lại. Phải tick đồng ý mới vào được.
import { ref } from 'vue'
import { ChevronRight, FileText, Loader2, Ban, ShieldCheck } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import logoUrl from '@/assets/logo.svg'
import TermsText from '../components/TermsText.vue'
import { errText, quitApp } from '../lib/backend'
import { TERMS_VERSION } from '../lib/terms'
import { acceptTerms, state } from '../lib/store'

const agreed = ref(false)
const saving = ref(false)
const error = ref('')
const updated = (state.terms?.acceptedVersion ?? 0) > 0

const summary = [
  { icon: FileText, title: 'Tài liệu bạn có quyền dùng', desc: 'Tài liệu của bạn, sách hết bảo hộ, hoặc được tác giả cho phép.' },
  { icon: Ban, title: 'Không dùng sách còn bản quyền', desc: 'Trừ khi được tác giả hoặc chủ sở hữu cho phép.' },
  { icon: ShieldCheck, title: 'Chạy trên máy bạn', desc: 'Tài liệu không gửi đi đâu. Bạn tự chịu trách nhiệm nội dung.' },
]

async function agree() {
  error.value = ''
  saving.value = true
  try {
    await acceptTerms()
  } catch (e) {
    error.value = errText(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="flex-1 min-h-0 overflow-auto">
    <div class="mx-auto flex h-full max-w-2xl flex-col px-6 py-8">
      <div class="flex items-center gap-3">
        <img :src="logoUrl" alt="" class="h-10 w-10 rounded-xl" />
        <div>
          <h1 class="text-xl font-semibold tracking-tight">Điều khoản sử dụng</h1>
          <p class="text-sm text-muted-foreground">
            <template v-if="updated">Điều khoản vừa được cập nhật lên phiên bản {{ TERMS_VERSION }}. Vui lòng đọc và đồng ý lại để tiếp tục.</template>
            <template v-else>Một bước cuối trước khi bắt đầu.</template>
          </p>
        </div>
      </div>

      <div class="mt-5 grid grid-cols-3 gap-2.5">
        <div v-for="s in summary" :key="s.title" class="flex gap-3 rounded-lg border border-border p-3">
          <span class="grid h-8 w-8 shrink-0 place-items-center rounded-md bg-primary/10 text-primary"><component :is="s.icon" class="h-4 w-4" /></span>
          <div class="min-w-0">
            <p class="text-sm font-medium">{{ s.title }}</p>
            <p class="text-xs leading-relaxed text-muted-foreground">{{ s.desc }}</p>
          </div>
        </div>
      </div>

      <div class="mt-4 min-h-[180px] flex-1 overflow-auto rounded-lg border border-border bg-muted/30 px-5 py-4">
        <TermsText />
      </div>

      <label class="mt-4 flex cursor-pointer items-start gap-2.5 text-sm">
        <input v-model="agreed" type="checkbox" class="mt-0.5 h-4 w-4 accent-[hsl(var(--primary))]" />
        <span>Tôi đã đọc và đồng ý với Điều khoản sử dụng. Tôi chỉ đưa vào Sano tài liệu tôi có quyền sử dụng và tự chịu trách nhiệm về nội dung mình tạo ra.</span>
      </label>
      <p v-if="error" class="mt-2 text-sm text-destructive">{{ error }}</p>
      <div class="mt-4 flex items-center justify-between">
        <Button variant="ghost" class="text-muted-foreground" @click="quitApp">Thoát Sano</Button>
        <Button :disabled="!agreed || saving" @click="agree">
          <Loader2 v-if="saving" class="h-4 w-4 animate-spin" /> Đồng ý và bắt đầu <ChevronRight class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
