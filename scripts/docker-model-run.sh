#!/usr/bin/env bash
set -euo pipefail

MODEL_REF="${1:-hf.co/meta-llama/Llama-3.2-3B-Instruct}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker CLI not found in PATH" >&2
  exit 1
fi

if ! docker model version >/dev/null 2>&1; then
  echo "Docker Model Runner is not available or not enabled." >&2
  echo "Enable Model Runner first in Docker Desktop, or install docker-model-plugin on Docker Engine." >&2
  exit 1
fi

if ! command -v hf >/dev/null 2>&1; then
  echo "hf CLI not found. Install it first, then run: hf auth login" >&2
  exit 1
fi

echo "Make sure you already ran: hf auth login"
echo "Starting model: ${MODEL_REF}"
docker model run --detach "${MODEL_REF}"

echo "Model is warming up in Docker Model Runner."
echo "Checking running models:"
docker model ps || true

echo "Default .env values for backend are already aligned with DMR:"
echo "LLM_BASE_URL=http://model-runner.docker.internal/engines"
echo "LLM_MODEL_NAME=${MODEL_REF}"
