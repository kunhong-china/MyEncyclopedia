#!/bin/bash
# Quick Start Script for Project Jarvis

set -e

echo "🚀 Project Jarvis - Quick Start Script"
echo "========================================"

# Check if we're in the right directory
if [ ! -d "client" ] || [ ! -d "server" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Server setup
echo ""
echo "📦 Setting up server..."
cd server

if [ ! -f ".env" ]; then
    echo "Creating server .env from template..."
    cp .env.example .env
    echo "⚠️  Please edit server/.env with your configuration"
fi

echo "Installing Go dependencies..."
go mod tidy

echo "✅ Server setup complete"
cd ..

# Client setup
echo ""
echo "📦 Setting up client..."
cd client

if [ ! -f ".env" ]; then
    echo "Creating client .env from template..."
    cp .env.example .env
    echo "⚠️  Please edit client/.env with your server IP address"
fi

if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

echo "Activating virtual environment and installing dependencies..."
source venv/bin/activate
pip install --upgrade pip -q
pip install -r requirements.txt

echo "✅ Client setup complete"
cd ..

echo ""
echo "✨ Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Edit server/.env with your Ollama configuration"
echo "2. Edit client/.env with your server IP address"
echo "3. Start the server: cd server && go run ."
echo "4. Start the client: cd client && source venv/bin/activate && python main.py"
echo ""
echo "For production deployment, see DEPLOYMENT.md"
