#!/usr/bin/env python3
"""
make_demo_gif.py - Compiles authentic UI screenshots into an optimized animated demo GIF
"""

import os
from PIL import Image

def main():
    images_dir = "docs/images"
    frames_names = [
        "distillforge_dashboard.png",
        "distillforge_pipeline.png",
        "distillforge_cockpit.png",
        "distillforge_curation.png",
        "distillforge_finops.png",
        "distillforge_playground.png",
    ]

    frames = []
    target_width = 1200
    target_height = 750

    for name in frames_names:
        p = os.path.join(images_dir, name)
        if not os.path.exists(p):
            print(f"Skipping missing: {p}")
            continue
        im = Image.open(p).convert("RGBA")
        im_resized = im.resize((target_width, target_height), Image.Resampling.LANCZOS)
        # Convert to P mode with 256 colors adaptive palette
        im_palette = im_resized.convert("RGB").convert("P", palette=Image.Palette.ADAPTIVE, colors=256)
        frames.append(im_palette)
        print(f"Processed frame: {name}")

    if not frames:
        print("No frames found!")
        return

    out_gif = os.path.join(images_dir, "distillforge_demo.gif")
    frames[0].save(
        out_gif,
        save_all=True,
        append_images=frames[1:],
        duration=2200,  # 2.2 seconds per slide
        loop=0,
        optimize=True,
    )
    size_mb = os.path.getsize(out_gif) / (1024 * 1024)
    print(f"Generated demo GIF: {out_gif} ({size_mb:.2f} MB, {len(frames)} slides)")

if __name__ == "__main__":
    main()
