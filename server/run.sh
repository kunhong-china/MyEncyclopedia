#!/bin/bash
# Run the Jarvis server

cd "$(dirname "$0")"

if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "🚀 Starting Jarvis Server..."
echo "📡 Port: ${JARVIS_PORT:-:8080}"
echo "🤖 Ollama: ${OLLAMA_ENDPOINT:-http://localhost:11434}"
echo "🧠 Model: ${OLLAMA_MODEL:-gemma2:27b}"
echo ""

go run .
