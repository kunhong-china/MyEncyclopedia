#!/bin/bash
# Run the Jarvis client

cd "$(dirname "$0")"

if [ ! -d "venv" ]; then
    echo "❌ Virtual environment not found. Please run setup.sh first."
    exit 1
fi

source venv/bin/activate

if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "🎙️  Starting Jarvis Client..."
echo "🌐 Server: ${JARVIS_SERVER_URL:-ws://localhost:8080/ws}"
echo "👂 Wake word: ${WAKE_WORD:-jarvis}"
echo ""

python main.py
