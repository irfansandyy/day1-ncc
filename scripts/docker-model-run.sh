#!/usr/bin/env bash
set -euo pipefail

MODEL_REF="${1:-hf.co/bartowski/Llama-3.2-3B-Instruct-GGUF:Q6_K}"
HF_TOKEN_FILE="${HF_TOKEN_FILE:-${HOME}/.cache/huggingface/token}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker CLI not found in PATH" >&2
  exit 1
fi

if ! docker model version >/dev/null 2>&1; then
  echo "Docker Model Runner is not available or not enabled." >&2
  echo "Enable Model Runner first in Docker Desktop, or install docker-model-plugin on Docker Engine." >&2
  exit 1
fi

if [[ -z "${HF_TOKEN:-}" ]] && [[ -f "${HF_TOKEN_FILE}" ]]; then
  export HF_TOKEN
  HF_TOKEN="$(tr -d '\r\n' < "${HF_TOKEN_FILE}")"
fi

if [[ -z "${HF_TOKEN:-}" ]]; then
  echo "HF_TOKEN is not set and token file was not found at ${HF_TOKEN_FILE}." >&2
  echo "Use one of these methods before running this script:" >&2
  echo "  1) export HF_TOKEN=\$(cat ~/.cache/huggingface/token)" >&2
  echo "  2) env PATH=\"${HOME}/.local/bin:\$PATH\" hf auth login" >&2
  exit 1
fi

echo "Starting model: ${MODEL_REF}"
docker model run --detach "${MODEL_REF}"

echo "Model is warming up in Docker Model Runner."
echo "Checking running models:"
docker model ps || true

echo "Default .env values for backend are already aligned with DMR:"
echo "LLM_BASE_URL=http://model-runner.docker.internal/engines"
echo "LLM_MODEL_NAME=${MODEL_REF}"
