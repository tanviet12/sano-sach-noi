<script setup lang="ts">
// Khung cửa sổ Sano (bám wireframe WfDesktop): thanh tiêu đề, màn cài bộ đọc
// hoặc thanh bên + nội dung, hộp cập nhật phủ trên cùng.
import { onMounted, ref } from 'vue'
import { platform } from './lib/backend'
import { init, state } from './lib/store'
import TitleBar from './components/TitleBar.vue'
import AppSidebar from './components/AppSidebar.vue'
import UpdateDialog from './components/UpdateDialog.vue'
import SetupView from './views/SetupView.vue'
import TermsView from './views/TermsView.vue'
import LibraryView from './views/LibraryView.vue'
import PlayerView from './views/PlayerView.vue'
import SettingsView from './views/SettingsView.vue'
import AboutView from './views/AboutView.vue'
import CreateView from './views/CreateView.vue'

const os = ref('')
onMounted(async () => {
  os.value = await platform()
  await init()
})
</script>

<template>
  <div class="relative h-dvh w-full overflow-hidden bg-background text-foreground flex flex-col">
    <TitleBar :os="os" />

    <SetupView v-if="state.view === 'setup'" />
    <TermsView v-else-if="state.view === 'terms'" />

    <div v-else class="flex-1 flex min-h-0">
      <AppSidebar />
      <main class="flex-1 min-w-0 flex flex-col">
        <LibraryView v-if="state.view === 'library'" />
        <PlayerView v-else-if="state.view === 'player'" />
        <SettingsView v-else-if="state.view === 'settings'" />
        <AboutView v-else-if="state.view === 'about'" />
        <CreateView v-else />
      </main>
    </div>

    <UpdateDialog v-if="state.update !== 'closed'" />
  </div>
</template>
