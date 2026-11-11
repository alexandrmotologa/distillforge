#!/usr/bin/env python3
"""
capture_screenshots.py - Captures authentic real screenshots of distillforge Web Studio
Uses Chrome DevTools Protocol (CDP) over WebSockets:
- Desktop viewports (1440x900, Retina scale)
- Mobile responsive viewports (390x844)
- Animated demo GIF (docs/images/distillforge_demo.gif)
"""

import asyncio
import base64
import json
import os
import subprocess
import time
import urllib.request
from PIL import Image

CHROME_PATH = r"C:\Program Files\Google\Chrome\Application\chrome.exe"
STUDIO_URL = "http://localhost:8090"
DEBUG_PORT = 9222

async def cdp_command(ws, method, params=None, msg_id=1):
    payload = {"id": msg_id, "method": method, "params": params or {}}
    await ws.send(json.dumps(payload))
    while True:
        resp = await ws.recv()
        data = json.loads(resp)
        if data.get("id") == msg_id:
            return data.get("result", {})

async def capture_tab(ws, tab_index, output_path, msg_id, width=1440, height=900, scale=1.5):
    # Set device metrics
    await cdp_command(ws, "Emulation.setDeviceMetricsOverride", {
        "width": width,
        "height": height,
        "deviceScaleFactor": scale,
        "mobile": width < 600,
    }, msg_id=msg_id)
    msg_id += 1

    # Click navigation tab if tab_index >= 0
    if tab_index >= 0:
        click_script = f"""
        (() => {{
            const buttons = document.querySelectorAll('header nav button');
            if (buttons && buttons[{tab_index}]) {{
                buttons[{tab_index}].click();
                return buttons[{tab_index}].textContent;
            }}
            return null;
        }})()
        """
        await cdp_command(ws, "Runtime.evaluate", {"expression": click_script}, msg_id=msg_id)
        msg_id += 1
        await asyncio.sleep(0.8)  # wait for animations / charts
    else:
        await asyncio.sleep(0.5)

    # Capture screenshot
    res = await cdp_command(ws, "Page.captureScreenshot", {"format": "png"}, msg_id=msg_id)
    msg_id += 1

    b64_data = res.get("data", "")
    if b64_data:
        os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
        with open(output_path, "wb") as f:
            f.write(base64.b64decode(b64_data))
        print(f"Captured: {output_path} ({width}x{height})")

    return msg_id

async def run_capture():
    import websockets

    # 1. Start Chrome with remote debugging
    chrome_proc = subprocess.Popen([
        CHROME_PATH,
        "--headless=new",
        "--disable-gpu",
        "--remote-allow-origins=*",
        f"--remote-debugging-port={DEBUG_PORT}",
        STUDIO_URL
    ])

    time.sleep(2.0)

    try:
        # 2. Get WebSocket debugger URL
        req = urllib.request.urlopen(f"http://localhost:{DEBUG_PORT}/json")
        targets = json.loads(req.read().decode())
        ws_url = targets[0]["webSocketDebuggerUrl"]
        print(f"Connected to CDP: {ws_url}")

        async with websockets.connect(ws_url, max_size=20*1024*1024) as ws:
            msg_id = 1
            await cdp_command(ws, "Page.enable", {}, msg_id)
            msg_id += 1

            # Navigate to Studio
            await cdp_command(ws, "Page.navigate", {"url": STUDIO_URL}, msg_id)
            msg_id += 1
            await asyncio.sleep(1.5)

            tabs = [
                ("dashboard", 0, "docs/images/distillforge_dashboard.png"),
                ("new_pipeline", 1, "docs/images/distillforge_pipeline.png"),
                ("training_cockpit", 2, "docs/images/distillforge_cockpit.png"),
                ("curated_dataset", 3, "docs/images/distillforge_curation.png"),
                ("head_to_head_diff", 4, "docs/images/distillforge_diff.png"),
                ("finops_roi", 5, "docs/images/distillforge_finops.png"),
                ("edge_playground", 6, "docs/images/distillforge_playground.png"),
            ]

            captured_paths = []
            for name, idx, out_file in tabs:
                msg_id = await capture_tab(ws, idx, out_file, msg_id, width=1440, height=900)
                captured_paths.append(out_file)

            # Mobile captures (390x844 iPhone 14/15 viewport)
            msg_id = await capture_tab(ws, 0, "docs/images/distillforge_mobile_dashboard.png", msg_id, width=390, height=844, scale=2.0)
            msg_id = await capture_tab(ws, 1, "docs/images/distillforge_mobile_pipeline.png", msg_id, width=390, height=844, scale=2.0)

        # 3. Create animated demo GIF
        print("Compiling animated demo GIF...")
        gif_frames = []
        for p in captured_paths:
            if os.path.exists(p):
                im = Image.open(p).convert("RGBA")
                # Downsample to crisp web size (960x600)
                im_resized = im.resize((960, 600), Image.Resampling.LANCZOS)
                # Convert to palette with adaptive dithering
                im_palette = im_resized.convert("P", palette=Image.Palette.ADAPTIVE)
                gif_frames.append(im_palette)

        if gif_frames:
            gif_path = "docs/images/distillforge_demo.gif"
            gif_frames[0].save(
                gif_path,
                save_all=True,
                append_images=gif_frames[1:],
                duration=1800,  # 1.8s per tab
                loop=0,
                optimize=True,
            )
            print(f"Generated Demo GIF: {gif_path} ({len(gif_frames)} frames)")

    finally:
        chrome_proc.terminate()
        chrome_proc.wait()

if __name__ == "__main__":
    asyncio.run(run_capture())
