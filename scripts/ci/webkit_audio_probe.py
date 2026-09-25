#!/usr/bin/env python3
"""Thử WebKitGTK 4.1 phát MP3 như Sano: qua scheme tự định nghĩa (giống Wails
đưa file) và qua file://. In một dòng kết quả cho mỗi cách rồi thoát.

    xvfb-run python3 webkit_audio_probe.py <file.mp3>
"""
import http.server
import os
import sys
import threading

import gi

gi.require_version("Gtk", "3.0")
gi.require_version("WebKit2", "4.1")
from gi.repository import Gio, GLib, Gtk, WebKit2  # noqa: E402

MP3 = os.path.abspath(sys.argv[1])
DATA = open(MP3, "rb").read()
class H(http.server.SimpleHTTPRequestHandler):
    def log_message(self, *a):
        pass


httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), lambda *a, **k: H(*a, directory=os.path.dirname(MP3), **k))
threading.Thread(target=httpd.serve_forever, daemon=True).start()
HTTP_URL = f"http://127.0.0.1:{httpd.server_address[1]}/{os.path.basename(MP3)}"
TESTS = [("scheme", "sano://media/a.mp3"), ("blob", "sano://media/a.mp3"), ("file", GLib.filename_to_uri(MP3)), ("http", HTTP_URL)]
PAGE = """<!doctype html><title>...</title><script>
const a = new Audio(%r);
const log = [];
const t0 = Date.now();
const note = (s) => { log.push(((Date.now() - t0) / 1000).toFixed(1) + 's ' + s); document.title = 'LOG ' + log.join(' | '); };
const done = (s) => { note(s); document.title = s + ' || ' + log.join(' | '); };
for (const ev of ['loadstart', 'loadedmetadata', 'canplay', 'playing', 'stalled', 'suspend', 'waiting', 'abort', 'emptied'])
  a.addEventListener(ev, () => note(ev + ' rs=' + a.readyState + ' ns=' + a.networkState));
a.addEventListener('error', () => done('ERR media error code=' + (a.error && a.error.code) + ' ' + (a.error && a.error.message)));
a.play().then(() => setTimeout(() => done('OK currentTime=' + a.currentTime.toFixed(2)), 1500))
        .catch(e => done('ERR ' + e.name + ': ' + e.message));
setTimeout(() => done('ERR treo rs=' + a.readyState + ' ns=' + a.networkState + ' paused=' + a.paused), 12000);
</script>"""


BLOB_PAGE = PAGE.replace("const a = new Audio(%r);", "const a = new Audio();").replace(
    "a.play().then(",
    "fetch(%r).then(r => r.blob()).then(b => { note('fetch ' + b.size + 'B'); a.src = URL.createObjectURL(b); return a.play(); }).then(",
)


def handler(req):
    path = req.get_path()
    if path.endswith(".mp3"):
        body, ctype = DATA, "audio/mpeg"
    elif path.endswith("blob.html"):
        body, ctype = (BLOB_PAGE % TESTS[1][1]).encode(), "text/html"
    else:
        body, ctype = (PAGE % TESTS[0][1]).encode(), "text/html"
    stream = Gio.MemoryInputStream.new_from_bytes(GLib.Bytes.new(body))
    resp = WebKit2.URISchemeResponse.new(stream, len(body))
    resp.set_content_type(ctype)
    req.finish_with_response(resp)


ctx = WebKit2.WebContext.get_default()
ctx.register_uri_scheme("sano", handler)
ctx.get_security_manager().register_uri_scheme_as_secure("sano")
results = {}
win = Gtk.Window()
win.set_default_size(400, 300)
# Trong app người dùng bấm nút → có "user gesture"; ở đây cho tự phát.
policies = WebKit2.WebsitePolicies(autoplay=WebKit2.AutoplayPolicy.ALLOW)
view = WebKit2.WebView(web_context=ctx, website_policies=policies)
s = view.get_settings()
s.set_media_playback_requires_user_gesture(False)
s.set_allow_file_access_from_file_urls(True)
win.add(view)
win.show_all()
order = iter(TESTS)
current = [None]


def next_test():
    try:
        name, url = next(order)
    except StopIteration:
        Gtk.main_quit()
        return
    current[0] = name
    if name == "scheme":
        view.load_uri("sano://media/index.html")
    elif name == "blob":
        view.load_uri("sano://media/blob.html")
    elif name == "file":
        view.load_html(PAGE % url, "file:///")
    else:
        view.load_html(PAGE % url, "http://127.0.0.1/")
    GLib.timeout_add_seconds(20, timeout, name)


def timeout(name):
    if name not in results:
        results[name] = "ERR hết giờ"
        print(f"{name}: {results[name]}", flush=True)
        next_test()
    return False


def on_title(v, _p):
    t = v.get_title() or ""
    name = current[0]
    if name and name not in results and (t.startswith("OK") or t.startswith("ERR")):
        results[name] = t
        print(f"{name}: {t}", flush=True)
        next_test()


view.connect("notify::title", on_title)
next_test()
Gtk.main()
sys.exit(0 if all(r.startswith("OK") for r in results.values()) else 1)
