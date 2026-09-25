#!/usr/bin/env python3
"""Thử phát MP3 sau cú bấm chuột THẬT (xdotool) với chính sách tự phát mặc định
của WebKitGTK (như Wails): bấm → fetch qua scheme riêng → blob → play().

    xvfb-run python3 webkit_click_probe.py <file.mp3> <nounlock|unlock>
"""
import os
import subprocess
import sys

import gi

gi.require_version("Gtk", "3.0")
gi.require_version("WebKit2", "4.1")
from gi.repository import Gio, GLib, Gtk, WebKit2  # noqa: E402

DATA = open(os.path.abspath(sys.argv[1]), "rb").read()
MODE = sys.argv[2]
UNLOCK = "a.play().catch(() => {});" if MODE == "unlock" else ""
PAGE = f"""<!doctype html><title>...</title>
<button id=b style="position:fixed;inset:0;width:100%;height:100%">Nghe</button><script>
const a = new Audio();
const done = (s) => {{ document.title = s; }};
document.getElementById('b').onclick = () => {{
  {UNLOCK}
  fetch('sano://media/a.mp3').then(r => r.blob())
    .then(b => {{ a.src = URL.createObjectURL(b); return a.play(); }})
    .then(() => setTimeout(() => done('OK currentTime=' + a.currentTime.toFixed(2)), 1500))
    .catch(e => done('ERR ' + e.name + ': ' + e.message));
}};
document.title = 'READY';
</script>"""


def handler(req):
    if req.get_path().endswith(".mp3"):
        body, ctype = DATA, "audio/mpeg"
    else:
        body, ctype = PAGE.encode(), "text/html"
    stream = Gio.MemoryInputStream.new_from_bytes(GLib.Bytes.new(body))
    resp = WebKit2.URISchemeResponse.new(stream, len(body))
    resp.set_content_type(ctype)
    req.finish_with_response(resp)


ctx = WebKit2.WebContext.get_default()
ctx.register_uri_scheme("sano", handler)
ctx.get_security_manager().register_uri_scheme_as_secure("sano")
win = Gtk.Window()
win.set_default_size(400, 300)
win.move(0, 0)
view = WebKit2.WebView.new_with_context(ctx)  # chính sách tự phát mặc định
win.add(view)
win.show_all()
result = {"v": "ERR hết giờ"}


def on_title(v, _p):
    t = v.get_title() or ""
    if t == "READY":
        GLib.timeout_add(800, lambda: subprocess.call(["xdotool", "mousemove", "200", "150", "click", "1"]) and False)
    elif t.startswith("OK") or t.startswith("ERR"):
        result["v"] = t
        Gtk.main_quit()


view.connect("notify::title", on_title)
view.load_uri("sano://media/index.html")
GLib.timeout_add_seconds(20, lambda: Gtk.main_quit() or False)
Gtk.main()
print(f"click-{MODE}: {result['v']}", flush=True)
sys.exit(0 if result["v"].startswith("OK") else 1)
