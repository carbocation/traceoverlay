"""
inferd — lightweight PyTorch inference sidecar for traceoverlay.

Usage:
    cp models.yaml.example models.yaml   # then edit to point at your .pt files
    uvicorn server:app --host 0.0.0.0 --port 8081

Then start traceoverlay with:
    traceoverlay -config ... -inference_url http://localhost:8081
"""

from __future__ import annotations

import base64
import io
import os
from typing import Any

import numpy as np
import torch
import yaml
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image

# ── Model registry ──────────────────────────────────────────────────────────

_models: dict[str, torch.jit.ScriptModule] = {}
_configs: dict[str, dict[str, Any]] = {}


def _load_models(config_path: str = "models.yaml") -> None:
    if not os.path.exists(config_path):
        print(f"[inferd] No {config_path} found — starting with no models loaded.")
        return

    with open(config_path) as f:
        config = yaml.safe_load(f) or {}

    for name, cfg in config.get("models", {}).items():
        path = cfg.get("path", "")
        device = cfg.get("device", "cpu")
        try:
            m = torch.jit.load(path, map_location=device)
            m.eval()
            _models[name] = m
            _configs[name] = cfg
            print(f"[inferd] Loaded model '{name}' from {path} on {device}")
        except Exception as exc:
            print(f"[inferd] WARN: Failed to load model '{name}': {exc}")


_load_models()

# ── App ─────────────────────────────────────────────────────────────────────

app = FastAPI(title="inferd", description="PyTorch inference sidecar for traceoverlay")


@app.get("/healthz")
def healthz() -> dict:
    return {"status": "ok", "models": list(_models.keys())}


@app.get("/models")
def list_models() -> list[dict]:
    """Return each model's name and output_map so the UI can populate the dropdown."""
    return [
        {"name": name, "output_map": _configs[name].get("output_map")}
        for name in _models
    ]


@app.post("/infer")
async def infer(
    image: UploadFile = File(..., description="Input image (PNG or DICOM-exported PNG)"),
    model: str = Form(..., description="Model name from models.yaml"),
) -> JSONResponse:
    """
    Run inference on the uploaded image and return a grayscale PNG where each
    pixel value equals the model's predicted category index.

    Response JSON:
        mask_b64   — base64-encoded grayscale PNG (pixel value = category index)
        output_map — per-category mapping from models.yaml (null if not configured)
    """
    if model not in _models:
        raise HTTPException(status_code=404, detail=f"Model '{model}' not found")

    cfg = _configs[model]
    m = _models[model]
    device = cfg.get("device", "cpu")

    # ── Load and preprocess image ──────────────────────────────────────────
    img_bytes = await image.read()
    pil_img = Image.open(io.BytesIO(img_bytes))

    input_channels = cfg.get("input_channels", 1)
    pil_mode = "L" if input_channels == 1 else "RGB"
    pil_img = pil_img.convert(pil_mode)
    orig_size = pil_img.size  # (W, H) — PIL convention

    input_size = cfg.get("input_size", 256)
    pil_resized = pil_img.resize((input_size, input_size), Image.BILINEAR)

    arr = np.array(pil_resized, dtype=np.float32) / 255.0

    norm = cfg.get("normalize", {"mean": [0.5], "std": [0.5]})
    means = norm["mean"]
    stds = norm["std"]

    if input_channels == 1:
        arr = (arr - means[0]) / stds[0]
        # Shape: (1, 1, H, W)
        tensor = torch.tensor(arr).unsqueeze(0).unsqueeze(0).to(device)
    else:
        # arr shape: (H, W, 3)
        for c in range(3):
            arr[:, :, c] = (arr[:, :, c] - means[c]) / stds[c]
        # Shape: (1, 3, H, W)
        tensor = torch.tensor(arr).permute(2, 0, 1).unsqueeze(0).to(device)

    # ── Run model ─────────────────────────────────────────────────────────
    with torch.no_grad():
        output = m(tensor)

    # Support common output shapes:
    #   (B, C, H, W) — multi-class logits   → argmax over C
    #   (B, 1, H, W) — single-channel       → squeeze
    #   (B, H, W)    — already argmaxed     → squeeze batch dim
    #   (H, W)       — bare prediction      → use as-is
    if output.dim() == 4:
        if output.shape[1] > 1:
            pred = output.argmax(dim=1).squeeze(0)
        else:
            pred = output.squeeze(0).squeeze(0)
    elif output.dim() == 3:
        if output.shape[0] == 1:
            pred = output.squeeze(0)
        else:
            pred = output.argmax(dim=0)
    else:
        pred = output

    pred_np = pred.cpu().numpy().astype(np.uint8)

    # ── Resize prediction back to original image dimensions ───────────────
    pred_pil = Image.fromarray(pred_np, mode="L")
    pred_pil = pred_pil.resize(orig_size, Image.NEAREST)  # preserve category values

    # ── Encode as base64 PNG ──────────────────────────────────────────────
    buf = io.BytesIO()
    pred_pil.save(buf, format="PNG")
    mask_b64 = base64.b64encode(buf.getvalue()).decode()

    return JSONResponse({
        "mask_b64": mask_b64,
        "output_map": cfg.get("output_map"),  # None → client uses naive identity mapping
    })


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
