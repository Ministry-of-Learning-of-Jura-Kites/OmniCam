import os
import re
import json
import requests
import tempfile
import shutil
from pathlib import Path
from urllib.parse import urlparse
from pygltflib import GLTF2, BufferFormat, ImageFormat


def download_file(base_url, file_path, tmp_dir):
    """Fetches a resource and saves it to the local temp directory."""
    full_url = base_url + file_path
    local_path = Path(tmp_dir) / file_path

    # Create subdirectories if the glTF references textures in folders
    local_path.parent.mkdir(parents=True, exist_ok=True)

    with requests.get(full_url, stream=True) as r:
        r.raise_for_status()
        with open(local_path, "wb") as f:
            shutil.copyfileobj(r.raw, f)


def download_all_assets(page_url, tmp_dir, capture_id):
    """Discovers asset URLs from page metadata and downloads everything."""
    # 1. Get the base asset URL from og:video tag
    resp = requests.get(page_url)
    resp.raise_for_status()

    video_match = re.search(r'meta property="og:video" content="([^"]+)"', resp.text)
    if not video_match:
        raise Exception("Could not find og:video metadata in page")

    video_url = video_match.group(1)
    hostname = urlparse(video_url).hostname
    base_asset_url = f"https://{hostname}/captures/{capture_id}/"

    # 2. Download and parse raw.gltf
    gltf_path = Path(tmp_dir) / "raw.gltf"
    download_file(base_asset_url, "raw.gltf", tmp_dir)

    with open(gltf_path, "r") as f:
        data = json.load(f)

    # 3. Download referenced images
    if "images" in data:
        for img in data["images"]:
            uri = img.get("uri")
            if uri and not uri.startswith("data:"):
                download_file(base_asset_url, uri, tmp_dir)

    # 4. Download geometry buffer
    download_file(base_asset_url, "raw_geometry.bin", tmp_dir)

    return base_asset_url


def gltf_to_glb(tmp_dir, output_dir, capture_id):
    """Converts loose glTF files into a single self-contained GLB."""
    gltf_input = Path(tmp_dir) / "raw.gltf"
    output_path = Path(output_dir).absolute() / f"{capture_id}.glb"

    # Load the glTF
    # pygltflib automatically loads external buffers/images if they are in the same dir
    scene = GLTF2.load(gltf_input)

    for img in scene.images:
        if not img.mimeType:
            if img.uri and img.uri.lower().endswith(".png"):
                img.mimeType = "image/png"
            elif img.uri and img.uri.lower().endswith((".jpg", ".jpeg")):
                img.mimeType = "image/jpeg"

    # Load images into memory first
    scene.convert_images(ImageFormat.DATAURI)

    # Store images with buffer instead of base64
    scene.convert_images(ImageFormat.BUFFERVIEW)

    # The magic: convert_to_binary() packages everything into the GLB chunk
    # and removes external URI references automatically.
    scene.convert_buffers(BufferFormat.BINARYBLOB)
    # scene.convert_images(ImageFormat.DATAURI)

    Path(output_dir).mkdir(parents=True, exist_ok=True)
    scene.save_binary(output_path)
    print(f"Successfully saved to: {output_path}")


def main():
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("--url", required=True, help="URL to the polycam capture")
    parser.add_argument("--out", default=os.getcwd(), help="Output directory")
    args = parser.parse_args()

    # Extract ID
    id_match = re.search(r"poly.cam/capture/([a-zA-Z0-9-]+)", args.url)
    if not id_match:
        print("Invalid Polycam URL")
        return
    capture_id = id_match.group(1)

    with tempfile.TemporaryDirectory() as tmp_dir:
        print(f"Downloading assets for {capture_id}...")
        download_all_assets(args.url, tmp_dir, capture_id)

        print("Converting to GLB...")
        gltf_to_glb(tmp_dir, args.out, capture_id)


if __name__ == "__main__":
    main()

# usage: main.py [-h] --url URL [--out OUT]

# options:
#   -h, --help  show this help message and exit
#   --url URL   URL to the polycam capture
#   --out OUT   Output directory
