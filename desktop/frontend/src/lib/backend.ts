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
  voice: string
  rightsConfirmedAt: string // lúc tick xác nhận quyền dùng tài liệu (bắt buộc để render)
  introText: string
  keepHeadingNumbers: boolean
  dropStems: string[]
  coverPath: string
  readingEdits: Record<string, ReadingEdit>
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
}

export interface LibraryBook {
  slug: string
  title: string
  author: string
  category: string // trống = chưa phân loại
  cover: string
  coverUrl: string
  zip: string
  chapters: number
  sections: number
  durationSec: number
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
}
export interface BookDetail extends LibraryBook {
  dir: string
  tracks: Track[]
}

/** Thông tin sửa được của một cuốn — khớp library.Info bên Go. */
export interface BookInfo {
  title: string
  author: string
  category: string
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
  SaveSampleDocx(): Promise<string>
  InspectDocx(path: string, keepHeadingNumbers: boolean): Promise<Outline>
  Voices(): Promise<Voice[]>
  PreviewClips(s: BookSettings, stems: string[]): Promise<Clip[]>
  SpeakSample(voice: string, text: string): Promise<string>
  StartRender(s: BookSettings): Promise<RenderStatus>
  CancelRender(): Promise<void>
  RenderStatus(): Promise<RenderStatus | null>
  Library(): Promise<LibraryInfo>
  LibrarySize(): Promise<number>
  Book(slug: string): Promise<BookDetail>
  OpenBookFolder(slug: string): Promise<void>
  DeleteBook(slug: string): Promise<string>
  UpdateBookInfo(slug: string, info: BookInfo): Promise<LibraryBook>
  RevealBookZip(slug: string): Promise<void>
  OpenLibraryFolder(): Promise<void>
  ChooseCover(): Promise<CoverFile | null>
  ExportM4B(slug: string): Promise<M4BStatus | null>
  CancelM4B(): Promise<void>
  M4BStatus(): Promise<M4BStatus | null>
  RevealM4B(): Promise<void>
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

export async function checkTTS(): Promise<TTSStatus> {
  const app = goApp()
  if (!app) return { ...mockTTS }
  return app.CheckTTS()
}

const mockSetupInfo: SetupInfo = {
  os: 'macOS', arch: 'arm64', cpu: 'Apple M2', cores: 8, ramBytes: 16 * 2 ** 30,
  freeBytes: 120 * 2 ** 30, needBytes: 1500 * 2 ** 20, requiredFree: 2500 * 2 ** 20,
  downloadBytes: 1000 * 2 ** 20, dataDir: '~/Library/Application Support/Sano/tts',
  usedBytes: 0, enough: true, supported: true, note: '',
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

/** Mở hộp chọn file .docx của hệ điều hành. Huỷ → null. */
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

/** Hỏi nơi lưu rồi xuất M4B trong nền (tiến độ qua m4b:progress / m4b:finished). Huỷ hộp lưu → null. */
export async function exportM4B(slug: string): Promise<M4BStatus | null> {
  return need().ExportM4B(slug)
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
