// Cầu nối tới phần Go (Wails). Gọi thẳng window.go / window.runtime thay vì
// thư mục wailsjs/ sinh tự động → frontend typecheck + build được cả khi máy
// chưa cài Wails CLI. Mở bằng trình duyệt thường (vite dev) thì dùng dữ liệu giả
// cho các màn xem được; phần cần máy thật (nghe thử, render) báo lỗi rõ.
import { mockLibrary, mockOutline, mockVoices } from './mock'

export interface TTSStatus {
  ready: boolean
  python: string
  pythonFound: boolean
  pythonVersion: string
  scriptsDir: string
  modelsOk: boolean
  ffmpeg?: string
  source?: 'env' | 'app' | 'legacy' | ''
  dataDir?: string
  /** Bộ đọc app cài vẫn chạy (ready) nhưng nên sync lại thư viện Python (bản vá bảo mật). */
  update?: boolean
  message: string
  detail: string
}

/** Một dòng tiến độ cài bộ đọc (khớp setup.Step bên Go). */
export interface SetupStep {
  key: 'python' | 'vieneu' | 'models' | 'ffmpeg' | 'verify'
  label: string
  state: 'pending' | 'running' | 'done' | 'skipped' | 'error'
  pct: number
  detail: string
}

export interface SetupStatus {
  running: boolean
  done: boolean
  cancelled: boolean
  error: string
  hint: string
  steps: SetupStep[]
  elapsedSec: number
  etaSec: number
  logFile: string
  seq: number // tăng dần; sự kiện tới không theo thứ tự thì bỏ bản cũ hơn
}

/** Máy người dùng + dung lượng/thời gian cài (khớp setup.Info). */
export interface SetupInfo {
  os: string
  arch: string
  cpu: string
  cores: number
  ramBytes: number
  freeBytes: number
  needBytes: number
  requiredFree: number
  downloadBytes: number
  dataDir: string
  usedBytes: number
  enough: boolean
  supported: boolean
  blocked: boolean
  osVersion: string
  note: string
}

export interface UninstallResult {
  cancelled: boolean
  freedBytes: number
  dir: string
}

export interface DocxFile {
  path: string
  name: string
  size: number
}

export interface OutlineSection {
  stem: string
  title: string
  chars: number
  images: number
  toc: boolean
  tocReason: string
}
export interface OutlineChapter {
  title: string
  sections: OutlineSection[]
}
export interface Outline {
  title: string
  fileTitle: string
  chapters: OutlineChapter[]
  sections: number
  chars: number
  sampleSentence: string
  warnings: {
    images: number
    skippedImages?: number // hình bỏ qua vì quá lớn sau khi giải nén
    tables: number
    fakeHeadings: string[]
    unknownAcronyms: { word: string; count: number }[]
  }
}

export interface Voice {
  name: string
  desc: string
  featured: boolean
}

export interface ReadingEdit {
  from: string
  to: string
}

/** Lựa chọn ở các bước Tạo sách — khớp BookSettings bên Go. */
export interface BookSettings {
  path: string
  title: string
  author: string
  category: string
  series: string // tên bộ sách; trống = sách lẻ
  volume: number // số tập; 0 = tập kế tiếp
  voice: string
  rightsConfirmedAt: string // lúc tick xác nhận quyền dùng tài liệu (bắt buộc để render)
  introText: string
  keepHeadingNumbers: boolean
  dropStems: string[]
  coverPath: string
  readingEdits: Record<string, ReadingEdit>
  pronunciations: Record<string, string> // từ điển cách đọc riêng của cuốn (D12)
}

export interface Clip {
  stem: string
  title: string
  text: string
  full: boolean
  file: string
  durationSec: number
  url: string
}

export interface Progress {
  phase: 'render' | 'package' | 'done'
  done: number
  total: number
  doneChars: number
  totalChars: number
  stem: string
  title: string
  elapsedSec: number
}

export interface RenderStatus {
  running: boolean
  done: boolean
  cancelled: boolean
  error: string
  title: string
  slug: string
  progress: Progress
  source: string // '' = tạo trong app, 'mcp' = AI tạo qua MCP
}

/** Một cuốn AI nhờ tạo qua MCP (mcp_queue.go, wireframe D22). */
export interface MCPJob {
  id: string
  title: string
  author?: string
  voice: string
  client: string
  sections: number
  listenMin: number
  status: 'queued' | 'rendering' | 'staged' | 'saved' | 'failed' | 'cancelled'
  percent: number
  slug?: string
  error?: string
  autoSave: boolean
  pledgedAt?: string
  createdAt: number
  stagedAt?: number
}

export interface MCPLogEntry {
  at: number
  client: string
  tool: string
  text: string
  edit: boolean
  error?: string
  slug?: string
  undo?: string // mã bản cũ giữ trước khi AI sửa
  canUndo?: boolean
  undone?: boolean
}

/** Dữ liệu màn MCP (mcp_settings.go). */
export interface MCPInfo {
  localOn: boolean
  running: boolean
  allowEdit: boolean
  bridge: string
  bridgeFound: boolean
  claudeCode: string
  codex: string
  configJson: string
  claudeDesktop: '' | 'found' | 'added'
  active: string[]
  log: MCPLogEntry[]
}

export interface LibraryBook {
  slug: string
  title: string
  author: string
  translator?: string // dịch giả (D16)
  publisher?: string // nhà xuất bản (D16)
  category: string // trống = chưa phân loại
  series: string // tên bộ sách; trống = sách lẻ
  volume: number // số tập trong bộ (0 khi là sách lẻ)
  cover: string
  coverUrl: string
  thumbUrl?: string // bìa thu nhỏ cho kệ sách (trống = dùng coverUrl)
  zip: string
  chapters: number
  sections: number
  durationSec: number
  voice: string // giọng đọc; trống = không rõ (gói cũ không ghi)
  createdAt: string
}
export interface LibraryInfo {
  dir: string
  books: LibraryBook[]
}
export interface Track {
  chapter: string
  title: string
  file: string
  url: string
  durationSec: number
  chapterStart?: boolean // tiểu mục đầu của một chương
}
export interface BookDetail extends LibraryBook {
  dir: string
  tracks: Track[]
}

/** Thông tin sửa được của một cuốn — khớp library.Info bên Go. */
export interface BookInfo {
  title: string
  author: string
  translator: string
  publisher: string
  category: string
  series: string
  volume: number // 0 = tập kế tiếp
}

/** Một danh mục hoặc bộ sách kèm số cuốn — khớp library.Group bên Go. */
export interface Group {
  name: string
  count: number
}

export interface CoverFile {
  path: string
  dataUrl: string
}

/** Tiến độ xuất M4B — khớp m4b.Progress bên Go. */
export interface M4BProgress {
  phase: 'encode' | 'mux' | 'done'
  track: number
  tracks: number
  doneSec: number
  totalSec: number
  percent: number
}

/** Trạng thái lượt xuất M4B — khớp M4BStatus bên Go. */
export interface M4BStatus {
  running: boolean
  done: boolean
  cancelled: boolean
  error: string
  slug: string
  title: string
  path: string
  size: number
  durationSec: number
  chapters: number
  progress: M4BProgress
}

/** Kết quả kiểm tra bản mới trên GitHub Releases. */
export interface UpdateInfo {
  available: boolean
  version: string
  published: string
  notes: string[]
  url: string
  /** Tải + thay được ngay trong app; false thì `manual` nói lý do (hiện nút Mở trang tải). */
  autoUpdate: boolean
  manual: string
  /** Dung lượng file cài sẽ tải (byte, 0 nếu không rõ). */
  size: number
}

/** Tiến độ tự cập nhật (sự kiện update:progress). */
export interface UpdateStatus {
  phase: 'idle' | 'downloading' | 'ready' | 'error'
  version: string
  done: number
  total: number
  verified: boolean
  error: string
  applyOnQuit: boolean
}

/** Video chia sẻ: ảnh thẻ từng câu + đoạn tiếng [start, end) của tiểu mục (khớp ShareVideoRequest bên Go). */
export interface ShareVideoRequest {
  slug: string
  file: string
  start: number
  end: number
  frames: { png: string; sec: number }[]
  /** Ảnh các cột sóng âm sáng (base64) — ffmpeg tô dần từ trái sang theo thời gian. */
  wavePNG: string
  waveX: number
  waveY: number
  waveW: number
  waveH: number
  title: string
}

/** Video cả cuốn (D15) — khớp BookVideoPlan / BookVideoStatus bên Go. */
export interface BookVideoSeg {
  file: string
  extra?: string // mã âm thanh thêm (câu đọc thêm, nhạc hiệu, chuông)
  start: number
  dur: number
  silence: boolean
}
export interface BookVideoOverlay {
  png: string
  x: number
  y: number
  w: number
  h: number
  t0: number
  t1: number
}
export interface BookVideoPlan {
  frames: number[]
  segs: BookVideoSeg[]
  overlays: BookVideoOverlay[]
  width: number
  height: number
  title: string
  thumb: string
  srt: string
  desc: string
}
export interface BookVideoStatus {
  running: boolean
  done: boolean
  error: string
  slug: string
  title: string
  phase: 'frames' | 'audio' | 'video' | 'files'
  pct: number
  dir: string
  video: string
  bytes: number
  durSec: number
  started: number
}

/** Âm thanh thêm cho video (D16) — khớp VideoExtra bên Go. */
export interface VideoExtra {
  id: string
  key: 'brand' | 'info' | 'end' | 'music' | 'chime'
  url: string
  durSec: number
}
export interface VideoExtrasRequest {
  slug: string
  lines: { key: string; text: string }[]
  music: 'sano' | 'file' | 'none'
  musicPath: string
  chime: boolean
}

interface GoApp {
  Version(): Promise<string>
  CheckUpdate(): Promise<UpdateInfo>
  StartUpdate(): Promise<UpdateStatus>
  CancelUpdate(): Promise<void>
  UpdateStatus(): Promise<UpdateStatus>
  ApplyUpdate(): Promise<void>
  ApplyUpdateOnQuit(): Promise<UpdateStatus>
  CheckTTS(): Promise<TTSStatus>
  SetupInfo(): Promise<SetupInfo>
  SetupStatus(): Promise<SetupStatus>
  StartSetup(): Promise<SetupStatus>
  CancelSetup(): Promise<void>
  UninstallTTS(): Promise<UninstallResult>
  ThirdPartyNotices(): Promise<string>
  TermsStatus(): Promise<TermsStatus>
  AcceptTerms(version: number): Promise<TermsStatus>
  QuitApp(): Promise<void>
  ChooseDocx(): Promise<DocxFile | null>
  DescribeDocx(path: string): Promise<DocxFile>
  SampleDocx(): Promise<DocxFile>
  PastedText(text: string): Promise<DocxFile>
  SaveAIGuide(kind: string): Promise<string>
  SaveSampleDocx(): Promise<string>
  InspectDocx(path: string, keepHeadingNumbers: boolean): Promise<Outline>
  Voices(): Promise<Voice[]>
  PreviewClips(s: BookSettings, stems: string[]): Promise<Clip[]>
  SpeakSample(voice: string, text: string): Promise<string>
  StartRender(s: BookSettings): Promise<RenderStatus>
  CancelRender(): Promise<void>
  RenderStatus(): Promise<RenderStatus | null>
  MCPJobs(): Promise<MCPJob[]>
  MCPCommit(ids: string[]): Promise<MCPJob[]>
  MCPDiscard(id: string): Promise<MCPJob[]>
  MCPDismiss(): Promise<MCPJob[]>
  MCPInfo(): Promise<MCPInfo>
  SetMCPLocal(on: boolean): Promise<MCPInfo>
  SetMCPAllowEdit(on: boolean): Promise<MCPInfo>
  AddToClaudeDesktop(): Promise<MCPInfo>
  MCPUndo(slug: string, id: string): Promise<MCPInfo>
  Library(): Promise<LibraryInfo>
  LibrarySize(): Promise<number>
  Book(slug: string): Promise<BookDetail>
  BookTexts(slug: string): Promise<SectionText[]>
  ChooseBookZip(): Promise<string>
  PreviewBookZip(path: string): Promise<ImportPreview>
  ImportBookZip(path: string, replaceSlug: string): Promise<string>
  CancelImport(): Promise<void>
  OpenBookFolder(slug: string): Promise<void>
  DeleteBook(slug: string): Promise<string>
  UpdateBookInfo(slug: string, info: BookInfo): Promise<LibraryBook>
  SeriesList(): Promise<Group[]>
  RenameCategory(old: string, name: string): Promise<number>
  DeleteCategory(name: string): Promise<number>
  RenameSeries(old: string, name: string): Promise<number>
  DeleteSeries(name: string): Promise<number>
  LibraryOrder(): Promise<string[]>
  RecordListening(slug: string, listenSec: number, audioSec: number): Promise<void>
  MarkFinished(slug: string): Promise<void>
  ListenLog(): Promise<ListenLog>
  ClearListenLog(): Promise<string>
  SetLibraryOrder(items: string[]): Promise<void>
  RevealBookZip(slug: string): Promise<void>
  OpenLibraryFolder(): Promise<void>
  ChooseCover(): Promise<CoverFile | null>
  ExportM4B(slug: string, sectionSec: number, chapterSec: number, ask: boolean): Promise<M4BStatus | null>
  CanAirDrop(): Promise<boolean>
  AirDropM4B(): Promise<void>
  SaveShareImage(png: string, title: string): Promise<string>
  CopyShareImage(png: string): Promise<void>
  CanCopyImage(): Promise<boolean>
  AirDropShareImage(png: string, title: string): Promise<void>
  MakeShareVideo(req: ShareVideoRequest): Promise<number>
  CancelShareVideo(): Promise<void>
  SaveShareVideo(): Promise<string>
  AirDropShareVideo(): Promise<void>
  BookVideoBegin(slug: string, title: string): Promise<string>
  BookVideoFrame(id: string, index: number, img: string): Promise<void>
  BookVideoFinish(id: string, plan: BookVideoPlan): Promise<void>
  BookVideoCancel(): Promise<void>
  BookVideoState(): Promise<BookVideoStatus | null>
  OpenBookVideoFolder(): Promise<void>
  BookVideoExtras(req: VideoExtrasRequest): Promise<VideoExtra[]>
  PickMusicFile(): Promise<string>
  CancelM4B(): Promise<void>
  M4BStatus(): Promise<M4BStatus | null>
  RevealM4B(): Promise<void>
  EditBook(slug: string): Promise<EditView>
  StartReread(slug: string, edits: SectionEdit[]): Promise<EditStatus>
  StartVoiceChange(slug: string, voice: string): Promise<EditStatus>
  CancelEdit(discard: boolean): Promise<void>
  DiscardVoiceChange(slug: string): Promise<void>
  EditStatus(): Promise<EditStatus | null>
  SaveBookInfo(slug: string, info: BookInfo): Promise<SaveInfoResult>
  SetBookCover(slug: string, path: string): Promise<LibraryBook>
  UseAutoCover(slug: string): Promise<LibraryBook>
  Pronunciations(): Promise<DictEntry[]>
  SetGlobalPronunciation(word: string, reading: string): Promise<void>
  DeleteGlobalPronunciation(word: string): Promise<void>
  BookPronunciations(slug: string): Promise<Record<string, string>>
  SetBookPronunciation(slug: string, word: string, reading: string): Promise<void>
  DeleteBookPronunciation(slug: string, word: string): Promise<void>
  CountWords(path: string, words: string[]): Promise<Record<string, number>>
}

interface WailsRuntime {
  BrowserOpenURL(url: string): void
  Environment(): Promise<{ buildType: string; platform: string; arch: string }>
  OnFileDrop(cb: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean): void
  OnFileDropOff(): void
  EventsOn(name: string, cb: (...data: unknown[]) => void): () => void
  ClipboardSetText(text: string): Promise<boolean>
}

declare global {
  interface Window {
    go?: { main?: { App?: GoApp } }
    runtime?: WailsRuntime
  }
}

function goApp(): GoApp | undefined {
  return window.go?.main?.App
}

/** Phần Go bắt buộc cho việc này (nghe thử, render...). */
function need(): GoApp {
  const app = goApp()
  if (!app) throw new Error('Cần mở trong phần mềm Sano (đang xem bằng trình duyệt, không có phần Go)')
  return app
}

/** Đang chạy trong cửa sổ Wails (có phần Go) hay trình duyệt thường. */
export function isDesktop(): boolean {
  return !!goApp()
}

/** Lỗi từ Go về dạng chuỗi dễ đọc. */
export function errText(e: unknown): string {
  if (e instanceof Error) return e.message
  return String(e)
}

export async function version(): Promise<string> {
  return (await goApp()?.Version()) ?? 'dev'
}

/** Hỏi GitHub có bản mới không. Trình duyệt thường (không có phần Go): null. */
export async function checkUpdate(): Promise<UpdateInfo | null> {
  const app = goApp()
  if (!app) return null
  return app.CheckUpdate()
}

/** Bắt đầu tải bản mới (tiến độ qua sự kiện update:progress). */
export async function startUpdate(): Promise<UpdateStatus> {
  return need().StartUpdate()
}

export async function cancelUpdate(): Promise<void> {
  await goApp()?.CancelUpdate()
}

export async function updateStatus(): Promise<UpdateStatus | null> {
  return (await goApp()?.UpdateStatus()) ?? null
}

/** Thay bản mới rồi thoát; bản mới tự mở lại. */
export async function applyUpdate(): Promise<void> {
  return need().ApplyUpdate()
}

/** "Khởi động lại sau": thay bản mới khi thoát Sano. */
export async function applyUpdateOnQuit(): Promise<UpdateStatus> {
  return need().ApplyUpdateOnQuit()
}

const mockTTS: TTSStatus = {
  ready: true,
  python: '~/VieNeu-TTS-v3/.venv/bin/python',
  pythonFound: true,
  pythonVersion: '3.12',
  scriptsDir: 'scripts/tts',
  modelsOk: true,
  message: 'Sẵn sàng',
  detail: 'Dữ liệu giả — mở trong trình duyệt, không có phần Go',
}
// ?tts=update lúc dev: xem màn "Cập nhật bộ đọc".
if (import.meta.env.DEV && new URLSearchParams(window.location.search).get('tts') === 'update') {
  Object.assign(mockTTS, {
    ready: true, update: true, message: 'Bộ đọc cần cập nhật thư viện',
    detail: 'Bản Sano này vá lỗi bảo mật trong thư viện Python của bộ đọc. Chỉ tải lại vài thư viện, mô hình giọng đọc giữ nguyên.',
  })
}

export async function checkTTS(): Promise<TTSStatus> {
  const app = goApp()
  if (!app) return { ...mockTTS }
  return app.CheckTTS()
}

const mockSetupInfo: SetupInfo = {
  os: 'macOS', arch: 'arm64', cpu: 'Apple M2', cores: 8, ramBytes: 16 * 2 ** 30,
  freeBytes: 120 * 2 ** 30, needBytes: 1500 * 2 ** 20, requiredFree: 2500 * 2 ** 20,
  downloadBytes: 1000 * 2 ** 20, dataDir: '~/Library/Application Support/Sano/tts',
  usedBytes: 0, enough: true, supported: true, blocked: false, osVersion: '', note: '',
}

export function mockSetupStatus(): SetupStatus {
  const step = (key: SetupStep['key'], label: string): SetupStep => ({ key, label, state: 'pending', pct: 0, detail: '' })
  return {
    running: false, done: false, cancelled: false, error: '', hint: '', elapsedSec: 0, etaSec: -1, logFile: '', seq: 0,
    steps: [
      step('python', 'Python'), step('vieneu', 'Bộ đọc VieNeu-TTS'), step('models', 'Mô hình giọng đọc (580 MB)'),
      step('ffmpeg', 'ffmpeg (ghi file MP3)'), step('verify', 'Kiểm tra đọc thử'),
    ],
  }
}

export async function setupInfo(): Promise<SetupInfo> {
  const app = goApp()
  if (!app) return { ...mockSetupInfo }
  return app.SetupInfo()
}

export async function setupStatus(): Promise<SetupStatus> {
  const app = goApp()
  if (!app) return mockSetupStatus()
  return app.SetupStatus()
}

/** Bắt đầu cài bộ đọc (tiến độ qua sự kiện setup:progress / setup:finished). */
export async function startSetup(): Promise<SetupStatus> {
  return need().StartSetup()
}

export async function cancelSetup(): Promise<void> {
  await goApp()?.CancelSetup()
}

/** Hỏi xác nhận rồi gỡ bộ đọc app đã cài. */
export async function uninstallTTS(): Promise<UninstallResult> {
  return need().UninstallTTS()
}

export interface TermsStatus {
  acceptedVersion: number // 0 = chưa đồng ý
  acceptedAt: string
}

/** Lần đồng ý điều khoản gần nhất. Xem bằng trình duyệt (không có Go) → coi như đã đồng ý. */
export async function termsStatus(): Promise<TermsStatus> {
  const app = goApp()
  if (!app) return { acceptedVersion: Number.MAX_SAFE_INTEGER, acceptedAt: '' }
  return app.TermsStatus()
}

export async function acceptTermsVersion(version: number): Promise<TermsStatus> {
  return need().AcceptTerms(version)
}

/** Đóng phần mềm (trình duyệt thì không làm gì). */
export async function quitApp(): Promise<void> {
  await goApp()?.QuitApp()
}

/** Toàn văn giấy phép bên thứ ba nhúng trong app ("" khi xem bằng trình duyệt). */
export async function thirdPartyNotices(): Promise<string> {
  return (await goApp()?.ThirdPartyNotices()) ?? ''
}

/** Mở hộp chọn file .docx / .epub của hệ điều hành. Huỷ → null. */
export async function chooseDocx(): Promise<DocxFile | null> {
  const app = goApp()
  if (!app) return { path: '/giả/ky-nang-giao-tiep.docx', name: 'ky-nang-giao-tiep.docx', size: 1_468_006 }
  return app.ChooseDocx()
}

export async function describeDocx(path: string): Promise<DocxFile> {
  return need().DescribeDocx(path)
}

/** Ghi file Word mẫu vào ~/Sano/.tam để thử tạo sách khi chưa có tài liệu. */
export async function sampleDocx(): Promise<DocxFile> {
  return need().SampleDocx()
}

/** Đổi văn bản AI trả về (% / # / ##) thành file Word tạm để nạp như file thường. */
export async function pastedText(text: string): Promise<DocxFile> {
  const app = goApp()
  if (!app) return { path: '/giả/Van-ban-dan-tu-AI.docx', name: 'Văn bản dán từ AI', size: text.length }
  return app.PastedText(text)
}

/** Lưu skill Claude (kind "claude" → zip) hoặc file hướng dẫn cho ChatGPT / Gemini. Huỷ → "". */
export async function saveAIGuide(kind: string): Promise<string> {
  const app = goApp()
  if (!app) return kind === 'claude' ? '~/Downloads/sano-sach-noi.zip' : '~/Downloads/sano-huong-dan-ai.txt'
  return app.SaveAIGuide(kind)
}

/** Lưu file Word mẫu về máy (hộp lưu file, mặc định thư mục Tải về). Huỷ → "". */
export async function saveSampleDocx(): Promise<string> {
  const app = goApp()
  if (!app) return '~/Downloads/Mau-sach-noi-Sano.docx'
  return app.SaveSampleDocx()
}

export async function inspectDocx(path: string, keepHeadingNumbers: boolean): Promise<Outline> {
  const app = goApp()
  if (!app) return mockOutline()
  return app.InspectDocx(path, keepHeadingNumbers)
}

export async function listVoices(): Promise<Voice[]> {
  const app = goApp()
  if (!app) return mockVoices()
  return app.Voices()
}

export async function previewClips(s: BookSettings, stems: string[]): Promise<Clip[]> {
  return need().PreviewClips(s, stems)
}

export async function speakSample(voice: string, text: string): Promise<string> {
  return need().SpeakSample(voice, text)
}

export async function startRender(s: BookSettings): Promise<RenderStatus> {
  return need().StartRender(s)
}

export async function mcpJobs(): Promise<MCPJob[]> {
  return (await goApp()?.MCPJobs()) ?? []
}

export async function mcpCommit(ids: string[]): Promise<MCPJob[]> {
  return need().MCPCommit(ids)
}

export async function mcpDiscard(id: string): Promise<MCPJob[]> {
  return need().MCPDiscard(id)
}

export async function mcpDismiss(): Promise<MCPJob[]> {
  return (await goApp()?.MCPDismiss()) ?? []
}

export async function mcpInfo(): Promise<MCPInfo | null> {
  return (await goApp()?.MCPInfo()) ?? null
}

export async function setMCPLocal(on: boolean): Promise<MCPInfo> {
  return need().SetMCPLocal(on)
}

export async function setMCPAllowEdit(on: boolean): Promise<MCPInfo> {
  return need().SetMCPAllowEdit(on)
}

export async function mcpUndo(slug: string, id: string): Promise<MCPInfo> {
  return need().MCPUndo(slug, id)
}

export async function addToClaudeDesktop(): Promise<MCPInfo> {
  return need().AddToClaudeDesktop()
}

export async function cancelRender(): Promise<void> {
  await goApp()?.CancelRender()
}

export async function renderStatus(): Promise<RenderStatus | null> {
  return (await goApp()?.RenderStatus()) ?? null
}

export async function library(): Promise<LibraryInfo> {
  const app = goApp()
  if (!app) return mockLibrary()
  return app.Library()
}

export async function book(slug: string): Promise<BookDetail> {
  return need().Book(slug)
}

/** Xem trước gói sách trước khi nhập — khớp library.ImportPreview bên Go. */
export interface ImportPreview {
  path: string
  fileName: string
  title: string
  author: string
  category: string
  voice: string
  chapters: number
  sections: number
  durationSec: number
  sizeBytes: number
  hasCover: boolean
  coverDataUrl: string // ảnh bìa đã kiểm (PNG/JPEG/WebP), trống nếu không có
  existingSlug: string // trống = chưa có trong thư viện
  existingTitle: string
  existingCreatedAt: string
}
export interface ImportProgress {
  done: number
  total: number
}

/** Mở hộp chọn gói sách (.zip); huỷ → chuỗi rỗng. */
export async function chooseBookZip(): Promise<string> {
  return need().ChooseBookZip()
}
export async function previewBookZip(path: string): Promise<ImportPreview> {
  return need().PreviewBookZip(path)
}
/** Nhập gói; trả mã sách đã nhập. replaceSlug: mã cuốn trùng người dùng chọn Thay thế
 *  (cuốn cũ vào Thùng rác); rỗng = giữ cả hai nếu trùng. */
export async function importBookZip(path: string, replaceSlug: string): Promise<string> {
  return need().ImportBookZip(path, replaceSlug)
}
export async function cancelImport(): Promise<void> {
  return need().CancelImport()
}

/** Chữ một tiểu mục — khớp library.SectionText bên Go. */
export interface SectionText {
  text: string // để hiện (như trong file Word)
  script: string // lời đã đọc thật (có tên tiểu mục ở đầu) — để ước lượng thời điểm
}

/** Chữ của từng tiểu mục, cùng thứ tự `tracks` (rỗng = không có). */
export async function bookTexts(slug: string): Promise<SectionText[]> {
  return need().BookTexts(slug)
}

export async function openBookFolder(slug: string): Promise<void> {
  return need().OpenBookFolder(slug)
}

/** Hỏi xác nhận rồi chuyển sách vào Thùng rác. Trả nơi đã chuyển tới, "" nếu người dùng huỷ. */
export async function deleteBook(slug: string): Promise<string> {
  return need().DeleteBook(slug)
}

/** Sửa tên, tác giả, danh mục (ghi metadata.json + gói zip). */
export async function updateBookInfo(slug: string, info: BookInfo): Promise<LibraryBook> {
  return need().UpdateBookInfo(slug, info)
}

/** Đổi tên / xoá danh mục, bộ sách cho mọi cuốn; trả số cuốn đã sửa. */
export async function renameCategory(old: string, name: string): Promise<number> {
  return need().RenameCategory(old, name)
}
export async function deleteCategory(name: string): Promise<number> {
  return need().DeleteCategory(name)
}
export async function renameSeries(old: string, name: string): Promise<number> {
  return need().RenameSeries(old, name)
}
export async function deleteSeries(name: string): Promise<number> {
  return need().DeleteSeries(name)
}

/** Thứ tự "Tự sắp xếp": khoá "b:<slug>" (sách lẻ) / "s:<tên bộ, chữ thường>" (bộ sách). */
let mockOrder: string[] = []
export async function libraryOrder(): Promise<string[]> {
  const app = goApp()
  if (!app) return [...mockOrder]
  return app.LibraryOrder()
}
// Nhật ký nghe (~/Sano/.nghe.json) cho thống kê. Trình duyệt (không có phần Go): bỏ qua.
export interface BookListen { listen: number; audio: number }
export interface DayListen { books: Record<string, BookListen>; hours: number[] }
export interface ListenLog { version: number; days: Record<string, DayListen>; finished: Record<string, string> }
export async function recordListening(slug: string, listenSec: number, audioSec: number): Promise<void> {
  const app = goApp()
  if (app) await app.RecordListening(slug, listenSec, audioSec)
}
export async function markFinished(slug: string): Promise<void> {
  const app = goApp()
  if (app) await app.MarkFinished(slug)
}
export async function listenLog(): Promise<ListenLog> {
  const app = goApp()
  if (!app) return { version: 1, days: {}, finished: {} }
  return app.ListenLog()
}
/** Xoá số liệu Hành trình nghe (Go hỏi lại, chuyển vào Thùng rác). Trả nơi chuyển tới; "" = huỷ. */
export async function clearListenLog(): Promise<string> {
  return need().ClearListenLog()
}
export async function setLibraryOrder(items: string[]): Promise<void> {
  const app = goApp()
  if (!app) {
    mockOrder = [...items]
    return
  }
  return app.SetLibraryOrder(items)
}

export async function revealBookZip(slug: string): Promise<void> {
  return need().RevealBookZip(slug)
}

/** Tổng dung lượng sách đã tạo (byte); trình duyệt thường: null. */
export async function librarySize(): Promise<number | null> {
  const app = goApp()
  if (!app) return null
  return app.LibrarySize()
}

export async function openLibraryFolder(): Promise<void> {
  return need().OpenLibraryFolder()
}

export async function chooseCover(): Promise<CoverFile | null> {
  return need().ChooseCover()
}

/**
 * Xuất M4B trong nền (tiến độ qua m4b:progress / m4b:finished), nghỉ theo gaps.
 * ask: hỏi nơi lưu (huỷ hộp lưu → null); không ask: lưu thẳng vào Tải về.
 */
export async function exportM4B(slug: string, gaps: { section: number; chapter: number }, ask = true): Promise<M4BStatus | null> {
  return need().ExportM4B(slug, gaps.section, gaps.chapter, ask)
}

/** Máy có AirDrop (macOS) không. */
export async function canAirDrop(): Promise<boolean> {
  return (await goApp()?.CanAirDrop()) ?? false
}

/** Mở bảng AirDrop cho file M4B vừa xuất. */
export async function airDropM4B(): Promise<void> {
  return need().AirDropM4B()
}

export async function cancelM4B(): Promise<void> {
  await goApp()?.CancelM4B()
}

export async function m4bStatus(): Promise<M4BStatus | null> {
  return (await goApp()?.M4BStatus()) ?? null
}

/** Mở thư mục và chọn sẵn file M4B vừa xuất. */
export async function revealM4B(): Promise<void> {
  return need().RevealM4B()
}

/** Nghe sự kiện Go đẩy lên (render:progress, render:finished). Trả hàm huỷ. */
export function onEvent<T>(name: string, cb: (data: T) => void): () => void {
  if (!window.runtime?.EventsOn) return () => {}
  return window.runtime.EventsOn(name, (d) => cb(d as T))
}

/** Chép chữ vào clipboard: API của Wails trước, trình duyệt sau. */
export async function copyText(text: string): Promise<boolean> {
  try {
    if (window.runtime?.ClipboardSetText) return await window.runtime.ClipboardSetText(text)
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

/** Chỉ link web http(s) mới được mở ra ngoài (không file:, javascript:, giao thức app lạ). */
export function isWebURL(url: string): boolean {
  try {
    const p = new URL(url).protocol
    return p === 'https:' || p === 'http:'
  } catch {
    return false
  }
}

/** Mở link bằng trình duyệt mặc định của máy (webview không tự mở tab mới). */
export function openURL(url: string) {
  if (!isWebURL(url)) return
  if (window.runtime) window.runtime.BrowserOpenURL(url)
  else window.open(url, '_blank', 'noopener')
}

/** 'darwin' | 'windows' | 'linux' | 'browser' */
export async function platform(): Promise<string> {
  if (!window.runtime) return 'browser'
  try {
    return (await window.runtime.Environment()).platform
  } catch {
    return 'browser'
  }
}

/** Nhận file kéo thả vào cửa sổ (chỉ trong Wails). Trả hàm huỷ đăng ký. */
export function onFileDrop(cb: (paths: string[]) => void): () => void {
  if (!window.runtime) return () => {}
  window.runtime.OnFileDrop((_x, _y, paths) => cb(paths), false)
  return () => window.runtime?.OnFileDropOff()
}

// ── Sửa sách (wireframe D11) ───────────────────────────────────────────────

/** Một tiểu mục để sửa — khớp library.EditSection bên Go. */
export interface EditSection {
  index: number
  chapterIndex: number
  chapter: string
  title: string
  text: string // chữ gốc (hiện khi nghe)
  script: string // lời đọc đang dùng
  intro: boolean // lời mở đầu "Bạn đang nghe sách nói…"
  file: string
  stem: string
  durationSec: number
}
export interface VoiceJob {
  voice: string
  done: string[]
}
export interface EditView extends Omit<LibraryBook, 'sections'> {
  coverAuto: boolean
  sections: EditSection[]
  voiceJob: VoiceJob | null
  urls: Record<string, string> // stem → URL MP3
}
export interface SectionEdit {
  index: number
  title: string
  text: string
}
/** Lượt sửa đang chạy / gần nhất — khớp EditStatus bên Go. */
export interface EditStatus {
  running: boolean
  done: boolean
  cancelled: boolean
  error: string
  kind: 'sections' | 'voice'
  slug: string
  bookTitle: string
  voice: string
  total: number
  finished: number
  current: number
  doneIdx: number[]
  queuedIdx: number[] | null
  startedAt: number
  remainSec: number
}
export interface SaveInfoResult {
  book: LibraryBook
  introEdit: SectionEdit | null
  coverRedrawn: boolean
}

export async function editBook(slug: string): Promise<EditView> {
  return need().EditBook(slug)
}
export async function startReread(slug: string, edits: SectionEdit[]): Promise<EditStatus> {
  return need().StartReread(slug, edits)
}
export async function startVoiceChange(slug: string, voice: string): Promise<EditStatus> {
  return need().StartVoiceChange(slug, voice)
}
export async function cancelEdit(discard: boolean): Promise<void> {
  await goApp()?.CancelEdit(discard)
}
export async function discardVoiceChange(slug: string): Promise<void> {
  return need().DiscardVoiceChange(slug)
}
export async function editStatus(): Promise<EditStatus | null> {
  return (await goApp()?.EditStatus()) ?? null
}
/** Lưu tên, tác giả, danh mục, bộ sách; bìa tự vẽ thì vẽ lại theo tên mới. */
export async function saveBookInfo(slug: string, info: BookInfo): Promise<SaveInfoResult> {
  return need().SaveBookInfo(slug, info)
}
export async function setBookCover(slug: string, path: string): Promise<LibraryBook> {
  return need().SetBookCover(slug, path)
}
export async function useAutoCover(slug: string): Promise<LibraryBook> {
  return need().UseAutoCover(slug)
}

// ── Từ điển cách đọc (wireframe D12) ─────────────────────────────────────

/** Một dòng từ điển chung — khớp DictEntry bên Go. */
export interface DictEntry {
  word: string
  reading: string
  builtin: boolean // có trong bộ chuẩn
  mine: boolean // tự thêm / ghi đè (xoá được)
  default: string // cách đọc chuẩn khi đã ghi đè
}
export async function pronunciations(): Promise<DictEntry[]> {
  return (await goApp()?.Pronunciations()) ?? []
}
export async function setGlobalPronunciation(word: string, reading: string): Promise<void> {
  return need().SetGlobalPronunciation(word, reading)
}
export async function deleteGlobalPronunciation(word: string): Promise<void> {
  return need().DeleteGlobalPronunciation(word)
}
export async function bookPronunciations(slug: string): Promise<Record<string, string>> {
  return (await need().BookPronunciations(slug)) ?? {}
}
export async function setBookPronunciation(slug: string, word: string, reading: string): Promise<void> {
  return need().SetBookPronunciation(slug, word, reading)
}
export async function deleteBookPronunciation(slug: string, word: string): Promise<void> {
  return need().DeleteBookPronunciation(slug, word)
}
export async function countWords(path: string, words: string[]): Promise<Record<string, number>> {
  if (!words.length) return {}
  return (await goApp()?.CountWords(path, words)) ?? {}
}

// ── Chia sẻ đoạn hay (D14) ──
export async function saveShareImage(png: string, title: string): Promise<string> {
  return need().SaveShareImage(png, title)
}
/** Chép ảnh vào clipboard: phần Go trên Mac, API trình duyệt ở máy khác. */
export async function copyShareImage(png: string, blob: Blob): Promise<void> {
  const app = goApp()
  if (app && (await app.CanCopyImage())) return app.CopyShareImage(png)
  await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })])
}
export async function airDropShareImage(png: string, title: string): Promise<void> {
  return need().AirDropShareImage(png, title)
}
export async function makeShareVideo(req: ShareVideoRequest): Promise<number> {
  return need().MakeShareVideo(req)
}
export async function cancelShareVideo(): Promise<void> {
  await goApp()?.CancelShareVideo()
}
export async function saveShareVideo(): Promise<string> {
  return need().SaveShareVideo()
}
export async function airDropShareVideo(): Promise<void> {
  return need().AirDropShareVideo()
}

// ── Video cả cuốn (D15) ──
export async function bookVideoBegin(slug: string, title: string): Promise<string> {
  return need().BookVideoBegin(slug, title)
}
export async function bookVideoFrame(id: string, index: number, img: string): Promise<void> {
  return need().BookVideoFrame(id, index, img)
}
export async function bookVideoFinish(id: string, plan: BookVideoPlan): Promise<void> {
  return need().BookVideoFinish(id, plan)
}
export async function bookVideoCancel(): Promise<void> {
  await goApp()?.BookVideoCancel()
}
export async function bookVideoState(): Promise<BookVideoStatus | null> {
  return (await goApp()?.BookVideoState()) ?? null
}
export async function openBookVideoFolder(): Promise<void> {
  return need().OpenBookVideoFolder()
}
export async function bookVideoExtras(req: VideoExtrasRequest): Promise<VideoExtra[]> {
  return (await need().BookVideoExtras(req)) ?? []
}
export async function pickMusicFile(): Promise<string> {
  return need().PickMusicFile()
}
