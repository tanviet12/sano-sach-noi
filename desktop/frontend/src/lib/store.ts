// Trạng thái dùng chung của cửa sổ: màn đang mở, các bước Tạo sách (dữ liệu
// thật từ phần Go), render nền, thư viện, bộ đọc.
// Một cuốn render một lúc: đang render thì bấm "Tạo sách mới" mở màn tiến độ.
import { computed, reactive } from 'vue'
import {
  cancelRender as goCancelRender, checkTTS, describeDocx, errText, inspectDocx, library as goLibrary,
  listVoices, onEvent, previewClips, renderStatus, startRender as goStartRender, version as goVersion,
  cancelSetup as goCancelSetup, mockSetupStatus, setupInfo as goSetupInfo, setupStatus as goSetupStatus,
  startSetup as goStartSetup, acceptTermsVersion, termsStatus, checkUpdate,
  startUpdate as goStartUpdate, cancelUpdate as goCancelUpdate, updateStatus as goUpdateStatus,
  applyUpdate as goApplyUpdate, applyUpdateOnQuit as goApplyUpdateOnQuit,
  type BookSettings, type TermsStatus, type Clip, type DocxFile, type LibraryInfo, type Outline, type ReadingEdit,
  type RenderStatus, type SetupInfo, type SetupStatus, type TTSStatus, type UpdateInfo, type UpdateStatus, type Voice,
} from './backend'
import { TERMS_VERSION } from './terms'

export type View = 'setup' | 'terms' | 'library' | 'create' | 'player' | 'settings' | 'about'
export type UpdateState = 'closed' | 'info'
/** Kiểm tra bản mới: chưa kiểm / đang kiểm / đang dùng bản mới nhất / có bản mới / lỗi. */
export type UpdateCheck = 'idle' | 'checking' | 'latest' | 'available' | 'error'

export interface TocSection {
  stem: string
  title: string
  chars: number
  on: boolean
  note: string
}
export interface TocEntry {
  title: string
  kind: 'skip' | 'chapter'
  chars: number
  on: boolean
  note: string
  stems: string[] // mọi tiểu mục của chương
  sections: TocSection[] // trống khi chương chỉ có 1 tiểu mục trùng tên (ô chương điều khiển luôn)
}

/** Mã lời mở đầu khi nghe thử (khớp bookmaker.IntroStem). */
export const INTRO_STEM = 'intro'
const DEFAULT_VOICE = 'Hải Đăng'
const LAST_VOICE_KEY = 'sano.lastVoice'

/** Giọng của cuốn tạo gần nhất (tiện ích riêng của máy, mất thì về giọng mặc định). */
export function lastVoice(): string {
  try {
    return localStorage.getItem(LAST_VOICE_KEY) || ''
  } catch {
    return ''
  }
}
function saveLastVoice(v: string) {
  try {
    localStorage.setItem(LAST_VOICE_KEY, v)
  } catch {
    // không lưu được thì thôi
  }
}

// Cho phép mở thẳng một màn qua ?screen=...&step=...&update=1 (giống wireframe) —
// tiện chụp màn hình so với wireframe khi phát triển. Chỉ khi chạy dev: bản phát
// hành bỏ qua để không nhảy qua được màn điều khoản.
const q = new URLSearchParams(import.meta.env.DEV ? window.location.search : '')
const views: View[] = ['setup', 'terms', 'library', 'create', 'player', 'settings', 'about']
const initialView = views.includes(q.get('screen') as View) ? (q.get('screen') as View) : 'library'

export const state = reactive({
  view: initialView as View,
  step: Math.min(6, Math.max(1, Number(q.get('step')) || 1)),
  update: (q.get('update') ? 'info' : 'closed') as UpdateState,
  // Bản mới trên GitHub Releases (chỉ có khi CheckUpdate báo). ?update=1 lúc dev
  // mở sẵn hộp cập nhật với dữ liệu mẫu để xem giao diện.
  updateInfo: (q.get('update') ? devUpdateInfo() : null) as UpdateInfo | null,
  updateCheck: 'idle' as UpdateCheck,
  autoUpdateCheck: readAutoUpdateCheck(),
  // Tự cập nhật: tiến độ tải/kiểm (từ phần Go), lỗi khi bấm (vd đang render),
  // hẹn cập nhật khi render xong. ?update=downloading|ready|error lúc dev để xem giao diện.
  upd: devUpdateStatus(q.get('update')),
  updError: '',
  updateAfterRender: false,

  version: '',
  tts: null as TTSStatus | null,
  ttsChecking: false,

  // Cài bộ đọc lần đầu
  setup: mockSetupStatus() as SetupStatus,
  setupInfo: null as SetupInfo | null,
  setupError: '', // lỗi khi bấm Cài (vd đang render) — khác lỗi trong lúc cài

  // B1 Nạp file
  file: null as DocxFile | null,
  fileError: '',
  loading: false,
  outline: null as Outline | null,
  title: '',
  author: '',
  category: '', // danh mục (trống = không phân loại)
  series: '', // bộ sách (trống = sách lẻ)
  volume: 0, // số tập (0 = tập kế tiếp)
  coverPath: '',
  coverDataUrl: '',

  // B2 Mục lục
  toc: [] as TocEntry[],
  keepHeadingNumbers: false,

  // B3 Giọng đọc
  voices: [] as Voice[],
  voicesError: '',
  voice: lastVoice() || DEFAULT_VOICE,
  sampleSentence: '',

  // B4 Lời mở đầu
  introEnabled: true,
  introText: '',
  introTouched: false,

  // B5 Nghe thử
  clips: [] as Clip[],
  clipsKey: '', // lựa chọn lúc render nghe thử; đổi lựa chọn → nghe thử lại
  previewing: false,
  previewError: '',
  heard: [] as string[], // stem các đoạn đã bấm nghe
  rightsConfirmedAt: '', // lúc tick "có quyền dùng tài liệu này" ở bước Nghe thử
  origText: {} as Record<string, string>, // lời đọc gốc của đoạn nghe thử (trước khi sửa)
  edits: {} as Record<string, ReadingEdit>,

  // B6 Render
  render: null as RenderStatus | null,
  renderError: '',

  // Thư viện + trình phát
  library: null as LibraryInfo | null,
  libraryError: '',
  playerSlug: '',
  terms: null as TermsStatus | null, // lần đồng ý điều khoản gần nhất
  playerAutoplay: false, // mở trình phát là phát tiếp luôn (nút phát ở hàng Nghe tiếp)
})

// ── Suy ra từ lựa chọn ────────────────────────────────────────────────────

const norm = (s: string) => s.toLocaleLowerCase('vi').replace(/[\s.:]+/g, ' ').trim()

/** Dựng mục lục có ô tick từ kết quả nạp file. Trang mục lục gốc bỏ tick sẵn. */
export function buildToc(o: Outline): TocEntry[] {
  return o.chapters.map((ch) => {
    const allToc = ch.sections.length > 0 && ch.sections.every((s) => s.toc)
    const single = ch.sections.length === 1 && norm(ch.sections[0].title) === norm(ch.title)
    return {
      title: ch.title || 'Nội dung',
      kind: allToc ? 'skip' : 'chapter',
      chars: ch.sections.reduce((n, s) => n + s.chars, 0),
      on: !allToc,
      note: allToc ? 'Gợi ý bỏ: mục lục gốc' : '',
      stems: ch.sections.map((s) => s.stem),
      sections: single ? [] : ch.sections.map((s) => ({
        stem: s.stem, title: s.title, chars: s.chars, on: !s.toc, note: s.toc ? 'Gợi ý bỏ: mục lục' : '',
      })),
    }
  })
}

/** Các tiểu mục sẽ đọc, theo thứ tự sách. */
export const selectedStems = computed(() => {
  const out: string[] = []
  for (const c of state.toc) {
    if (!c.on) continue
    if (c.sections.length === 0) out.push(...c.stems)
    else out.push(...c.sections.filter((s) => s.on).map((s) => s.stem))
  }
  return out
})

const charsByStem = computed(() => {
  const m: Record<string, number> = {}
  for (const ch of state.outline?.chapters ?? []) for (const s of ch.sections) m[s.stem] = s.chars
  return m
})

const titleByStem = computed(() => {
  const m: Record<string, string> = {}
  for (const ch of state.outline?.chapters ?? []) for (const s of ch.sections) m[s.stem] = s.title
  return m
})

/** Nhãn phần giới thiệu Sano tự chèn — khác "Lời mở đầu" vì sách hay có sẵn chương cùng tên. */
export const INTRO_LABEL = 'Giới thiệu sách'

export function stemTitle(stem: string) {
  return stem === INTRO_STEM ? INTRO_LABEL : titleByStem.value[stem] ?? stem
}

/** Lời mở đầu sẽ đọc: chưa tự sửa thì luôn theo tên sách + tác giả hiện tại. */
export function introText() {
  if (!state.introEnabled) return ''
  return state.introTouched ? state.introText.trim() : defaultIntro()
}

export const totalChars = computed(() =>
  selectedStems.value.reduce((n, s) => n + (charsByStem.value[s] ?? 0), 0) + introText().length,
)
// ~70 ký tự/giây render trên máy tham khảo; ~15 ký tự/giây nghe
export const estListen = computed(() => Math.max(1, Math.round(totalChars.value / 15 / 60)))
export const estRender = computed(() => Math.max(1, Math.round(totalChars.value / 70 / 60)))

export function settings(): BookSettings {
  const selected = new Set(selectedStems.value)
  const drop = (state.outline?.chapters ?? []).flatMap((c) => c.sections.map((s) => s.stem)).filter((s) => !selected.has(s))
  return {
    path: state.file?.path ?? '',
    title: state.title.trim(),
    author: state.author.trim(),
    category: state.category.trim(),
    series: state.series.trim(),
    volume: state.series.trim() ? Math.max(0, Math.floor(Number(state.volume) || 0)) : 0,
    voice: state.voice,
    introText: introText(),
    keepHeadingNumbers: state.keepHeadingNumbers,
    dropStems: drop,
    coverPath: state.coverPath,
    readingEdits: { ...state.edits },
    rightsConfirmedAt: state.rightsConfirmedAt,
  }
}

/** Lựa chọn ảnh hưởng tới lời đọc / giọng (không tính lời đã sửa ở B5). */
function previewKey() {
  const s = settings()
  return JSON.stringify([s.path, s.title, s.author, s.voice, s.introText, s.keepHeadingNumbers, s.dropStems])
}

/** Đủ điều kiện render cả cuốn: đã xác nhận quyền dùng tài liệu (nghe thử không bắt buộc). */
export const canRender = computed(() => !!state.rightsConfirmedAt)

// ── Render nền ────────────────────────────────────────────────────────────

export const rendering = computed(() => !!state.render?.running)
export const renderPct = computed(() => {
  const p = state.render?.progress
  if (!p || !p.totalChars) return 0
  if (p.phase === 'done') return 100
  return Math.min(99, Math.floor((p.doneChars / p.totalChars) * 100))
})
/** Phút còn lại: theo tốc độ thật khi đã xong ≥1 tiểu mục, trước đó theo ước lượng. */
export const remainMin = computed(() => {
  const p = state.render?.progress
  if (!p || !p.doneChars || !p.elapsedSec) return estRender.value
  const perChar = p.elapsedSec / p.doneChars
  return Math.max(1, Math.ceil(((p.totalChars - p.doneChars) * perChar) / 60))
})

// ── Điều hướng ────────────────────────────────────────────────────────────

export function go(v: View) {
  state.view = v
  if (v === 'library') void refreshLibrary()
  if (v !== 'create') return
  if (rendering.value) state.step = 6 // đang render → mở màn tiến độ
  else if (state.render?.done) resetCreate() // cuốn trước xong rồi → bắt đầu cuốn mới
}

export function openBook(slug: string, autoplay = false) {
  state.playerAutoplay = autoplay
  state.playerSlug = slug
  state.view = 'player'
}

export function resetCreate() {
  state.render = null
  state.renderError = ''
  state.step = 1
  clearFile()
}

// ── B1: nạp file thật ─────────────────────────────────────────────────────

export async function setFile(f: DocxFile) {
  state.file = f
  state.fileError = ''
  state.loading = true
  state.outline = null
  try {
    const o = await inspectDocx(f.path, state.keepHeadingNumbers)
    state.outline = o
    state.toc = buildToc(o)
    state.title = o.title || capitalize(o.fileTitle || f.name.replace(/\.docx$/i, '').replace(/[-_]+/g, ' ').trim())
    state.sampleSentence = o.sampleSentence
    state.clips = []
    state.heard = []
    state.rightsConfirmedAt = '' // file mới → xác nhận lại
    state.origText = {}
    state.edits = {}
    state.introTouched = false
    state.introText = defaultIntro()
  } catch (e) {
    state.fileError = errText(e)
    state.file = null
  } finally {
    state.loading = false
  }
}

/** Đổi cách đọc số đầu tiêu đề → nạp lại để số ký tự khớp (giữ tick đang chọn). */
export async function reloadOutline() {
  if (!state.file) return
  const keep = new Set(selectedStems.value)
  try {
    const o = await inspectDocx(state.file.path, state.keepHeadingNumbers)
    state.outline = o
    state.toc = buildToc(o).map((c) => ({
      ...c,
      on: c.stems.some((s) => keep.has(s)),
      sections: c.sections.map((s) => ({ ...s, on: keep.has(s.stem) })),
    }))
  } catch (e) {
    state.fileError = errText(e)
  }
}

/** Đường thử luồng thật khi phát triển: ?docx=/đường/dẫn/file.docx (chỉ bản dev). */
export async function loadDocxPath(path: string) {
  try {
    await setFile(await describeDocx(path))
  } catch (e) {
    state.fileError = errText(e)
  }
}

export function clearFile() {
  state.file = null
  state.fileError = ''
  state.outline = null
  state.toc = []
  state.title = ''
  state.author = ''
  state.category = ''
  state.series = ''
  state.volume = 0
  state.coverPath = ''
  state.coverDataUrl = ''
  state.clips = []
  state.heard = []
  state.rightsConfirmedAt = ''
  state.origText = {}
  state.edits = {}
}

function capitalize(s: string) {
  return s ? s.charAt(0).toLocaleUpperCase('vi') + s.slice(1) : s
}

export function defaultIntro() {
  const lines = ['Bạn đang nghe sách nói.', `Cuốn sách: ${state.title.trim() || 'chưa đặt tên'}.`]
  if (state.author.trim()) lines.push(`Tác giả: ${state.author.trim()}.`)
  return lines.join('\n\n')
}

// ── B3: giọng đọc ─────────────────────────────────────────────────────────

export async function loadVoices() {
  if (state.voices.length) return
  state.voicesError = ''
  try {
    state.voices = await listVoices()
    if (!state.voices.some((v) => v.name === state.voice)) state.voice = state.voices[0]?.name ?? DEFAULT_VOICE
  } catch (e) {
    state.voicesError = errText(e)
  }
}

// ── B5: nghe thử ──────────────────────────────────────────────────────────

/** Đoạn nghe thử mặc định: lời mở đầu (nếu có) + 2 tiểu mục đầu sẽ đọc. */
function defaultPreviewStems() {
  const stems = selectedStems.value.slice(0, 2)
  return introText() ? [INTRO_STEM, ...stems] : stems
}

/** Vào B5: lựa chọn đổi từ lần nghe trước (hoặc chưa nghe) thì render nghe thử lại. */
export async function ensurePreview() {
  if (state.previewing) return
  if (state.clips.length && state.clipsKey === previewKey()) return
  state.clips = []
  state.heard = []
  state.origText = {}
  state.edits = {}
  await addPreview(defaultPreviewStems())
}

/** Render thêm / render lại các đoạn nghe thử (giữ thứ tự, thay đoạn trùng). */
export async function addPreview(stems: string[]) {
  if (!stems.length || !state.file) return
  state.previewing = true
  state.previewError = ''
  try {
    const got = await previewClips(settings(), stems)
    for (const c of got) {
      if (!state.edits[c.stem] && !(c.stem in state.origText)) state.origText[c.stem] = c.text
      const i = state.clips.findIndex((x) => x.stem === c.stem)
      if (i >= 0) state.clips.splice(i, 1, c)
      else state.clips.push(c)
      state.heard = state.heard.filter((h) => h !== c.stem) // đoạn mới render phải nghe lại
    }
    state.clipsKey = previewKey()
  } catch (e) {
    state.previewError = errText(e)
  } finally {
    state.previewing = false
  }
}

/** Sửa lời đọc một đoạn rồi render lại đoạn đó; bản cuối dùng lời đã sửa. */
export async function editClip(stem: string, text: string) {
  const from = state.origText[stem]
  if (from === undefined) return
  if (text.trim() === from.trim()) delete state.edits[stem]
  else state.edits[stem] = { from, to: text.trim() }
  await addPreview([stem])
}

export function markHeard(stem: string) {
  if (!state.heard.includes(stem)) state.heard.push(stem)
}

// ── B6: render ────────────────────────────────────────────────────────────

export async function startRender() {
  state.renderError = ''
  try {
    state.render = await goStartRender(settings())
    saveLastVoice(state.voice)
    state.step = 6
  } catch (e) {
    state.renderError = errText(e)
  }
}

export async function cancelRender() {
  await goCancelRender()
}

/** Kết quả lượt render: xong → vào thư viện; hủy → về bước nghe thử. */
function onRenderFinished(st: RenderStatus) {
  if (state.updateAfterRender) {
    // Đã hẹn "Tự cập nhật khi render xong": tải luôn, xong thì hộp cập nhật hỏi khởi động lại.
    state.updateAfterRender = false
    state.update = 'info'
    void startUpdate()
  }
  state.render = st
  if (st.cancelled) {
    state.render = null
    state.step = 5
  } else if (st.error) {
    state.renderError = st.error
  }
  void refreshLibrary()
}

// ── Thư viện ──────────────────────────────────────────────────────────────

export async function refreshLibrary() {
  try {
    state.library = await goLibrary()
    state.libraryError = ''
  } catch (e) {
    state.libraryError = errText(e)
  }
}

// ── Bộ đọc + khởi động ────────────────────────────────────────────────────

export async function refreshTTS() {
  state.ttsChecking = true
  try {
    state.tts = await checkTTS()
  } catch (e) {
    state.tts = {
      ready: false, python: '', pythonFound: false, pythonVersion: '', scriptsDir: '', modelsOk: false,
      message: 'Không kiểm tra được bộ đọc', detail: String(e),
    }
  } finally {
    state.ttsChecking = false
  }
}

// ── Cài bộ đọc ────────────────────────────────────────────────────────────

export async function refreshSetupInfo() {
  try {
    state.setupInfo = await goSetupInfo()
  } catch {
    state.setupInfo = null
  }
}

/** Nhận trạng thái cài mới nhất; bỏ bản cũ hơn (sự kiện có thể tới lệch thứ tự). */
function applySetup(st: SetupStatus) {
  if (st.seq >= state.setup.seq) state.setup = st
}

export async function startSetup() {
  state.setupError = ''
  try {
    applySetup(await goStartSetup())
  } catch (e) {
    state.setupError = errText(e)
  }
}

export async function cancelSetup() {
  await goCancelSetup()
}

async function onSetupFinished(st: SetupStatus) {
  applySetup(st)
  await Promise.all([refreshTTS(), refreshSetupInfo()])
  state.voices = [] // bộ đọc đổi → hỏi lại danh sách giọng
}

export async function init() {
  onEvent<RenderStatus>('render:progress', (st) => { state.render = st })
  onEvent<RenderStatus>('render:finished', onRenderFinished)
  onEvent<UpdateStatus>('update:progress', (st) => { state.upd = st })
  onEvent<SetupStatus>('setup:progress', applySetup)
  onEvent<SetupStatus>('setup:finished', onSetupFinished)
  try {
    applySetup(await goSetupStatus()) // mở lại cửa sổ khi đang cài
  } catch {
    /* giữ trạng thái ban đầu */
  }
  state.version = await goVersion()
  const [st] = await Promise.all([renderStatus(), refreshLibrary()])
  if (st?.running) state.render = st // mở lại cửa sổ khi đang render
  const upd = await goUpdateStatus()
  if (upd && !q.get('update')) state.upd = upd
  await refreshTTS()
  state.terms = await termsStatus()
  // Lần mở đầu chưa có bộ đọc → màn cài bộ đọc; có rồi mà chưa đồng ý điều khoản
  // (hoặc điều khoản có bản mới) → màn điều khoản. Trừ khi đã chỉ định màn qua URL.
  // Chưa có bộ đọc, hoặc bộ đọc cần cập nhật thư viện (bản vá) → màn cài bộ đọc.
  if (!q.get('screen') && state.tts && (!state.tts.ready || state.tts.update)) state.view = 'setup'
  else if (!q.get('screen') && needTerms()) state.view = 'terms'
  if (state.autoUpdateCheck && !q.get('update')) void checkForUpdate()
  const dev = import.meta.env.DEV ? q.get('docx') : null
  if (dev) {
    state.view = 'create'
    state.step = 1
    await loadDocxPath(dev)
  }
}

const AUTO_UPDATE_KEY = 'sano.autoUpdateCheck'

function readAutoUpdateCheck(): boolean {
  try {
    return localStorage.getItem(AUTO_UPDATE_KEY) !== 'off'
  } catch {
    return true
  }
}

export function setAutoUpdateCheck(on: boolean) {
  state.autoUpdateCheck = on
  try {
    localStorage.setItem(AUTO_UPDATE_KEY, on ? 'on' : 'off')
  } catch {
    /* không lưu được thì chỉ áp cho lần chạy này */
  }
}

function devUpdateInfo(): UpdateInfo {
  return {
    available: true, version: '0.2.0', published: '', notes: ['Ghi chú phát hành mẫu (chỉ khi chạy dev)'],
    url: 'https://github.com/tanviet12/sano-sach-noi/releases', autoUpdate: q.get('manual') !== '1',
    manual: q.get('manual') === '1' ? 'Sano đang chạy thẳng từ file .dmg — kéo Sano vào thư mục Applications rồi mở lại để tự cập nhật được' : '',
    size: 18 << 20,
  }
}

function devUpdateStatus(mode: string | null): UpdateStatus {
  const st: UpdateStatus = { phase: 'idle', version: '0.2.0', done: 0, total: 18 << 20, verified: false, error: '', applyOnQuit: false }
  if (mode === 'downloading') return { ...st, phase: 'downloading', done: 11 << 20 }
  if (mode === 'ready') return { ...st, phase: 'ready', done: st.total, verified: true }
  if (mode === 'error') return { ...st, phase: 'error', error: 'chữ ký bản phát hành không hợp lệ — đây không phải bản chính thức của Sano, đã huỷ cập nhật' }
  return st
}

// ── Tự cập nhật ───────────────────────────────────────────────────────────

/** Tải bản mới (kiểm chữ ký + SHA256 ở phần Go). */
export async function startUpdate() {
  state.updError = ''
  try {
    state.upd = await goStartUpdate()
  } catch (e) {
    state.updError = errText(e)
  }
}

export async function cancelUpdate() {
  await goCancelUpdate()
}

/** Thay bản mới + khởi động lại. Đang render / cài bộ đọc / xuất M4B thì phần Go từ chối. */
export async function applyUpdate() {
  state.updError = ''
  try {
    await goApplyUpdate()
  } catch (e) {
    state.updError = errText(e)
  }
}

/** "Khởi động lại sau": thay khi thoát Sano. */
export async function applyUpdateLater() {
  try {
    state.upd = await goApplyUpdateOnQuit()
  } catch (e) {
    state.updError = errText(e)
    return
  }
  state.update = 'closed'
}

/** Đóng hộp cập nhật; lỗi lần trước không giữ lại cho lần mở sau. */
export function closeUpdate() {
  state.update = 'closed'
  state.updError = ''
  if (state.upd.phase === 'error') state.upd = { ...state.upd, phase: 'idle', error: '' }
}

/** Hỏi GitHub có bản mới không. Không có mạng / repo chưa công khai → 'error', không làm phiền. */
export async function checkForUpdate() {
  if (state.updateCheck === 'checking') return
  state.updateCheck = 'checking'
  try {
    const info = await checkUpdate()
    if (!info) {
      state.updateCheck = 'idle'
      return
    }
    state.updateInfo = info.available ? info : null
    state.updateCheck = info.available ? 'available' : 'latest'
  } catch {
    state.updateCheck = 'error'
  }
}

/** Chưa đồng ý điều khoản phiên bản hiện tại. */
export function needTerms() {
  return (state.terms?.acceptedVersion ?? 0) < TERMS_VERSION
}

/** Rời màn cài bộ đọc: chưa đồng ý điều khoản thì sang màn điều khoản trước. */
export function leaveSetup() {
  state.view = needTerms() ? 'terms' : 'library'
}

export async function acceptTerms() {
  state.terms = await acceptTermsVersion(TERMS_VERSION)
  state.view = 'library'
}
