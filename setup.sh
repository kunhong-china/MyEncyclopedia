#!/bin/bash
# Quick Start Script for Project Jarvis

set -e

MODE="all"

show_usage() {
    cat <<EOF
Usage: ./setup.sh [--client | --server | --all]

Options:
  --client   Set up only the client machine
  --server   Set up only the server machine
  --all      Set up both client and server (default)
  -h, --help Show this help message
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --client)
            MODE="client"
            ;;
        --server)
            MODE="server"
            ;;
        --all)
            MODE="all"
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "❌ Error: Unknown option '$1'"
            show_usage
            exit 1
            ;;
    esac
    shift
done

echo "🚀 Project Jarvis - Quick Start Script"
echo "========================================"

# Check if we're in the right directory
if [ ! -d "client" ] || [ ! -d "server" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

setup_server() {
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
}

setup_client() {
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
}

case "$MODE" in
    client)
        setup_client
        ;;
    server)
        setup_server
        ;;
    all)
        setup_server
        setup_client
        ;;
esac

echo ""
echo "✨ Setup Complete!"
echo ""
echo "Next steps:"

if [ "$MODE" = "server" ]; then
    echo "1. Edit server/.env with your Ollama configuration"
    echo "2. Start the server: cd server && go run ."
elif [ "$MODE" = "client" ]; then
    echo "1. Edit client/.env with your server IP address"
    echo "2. Start the client: cd client && source venv/bin/activate && python main.py"
else
    echo "1. Edit server/.env with your Ollama configuration"
    echo "2. Edit client/.env with your server IP address"
    echo "3. Start the server: cd server && go run ."
    echo "4. Start the client: cd client && source venv/bin/activate && python main.py"
fi

echo ""
echo "For production deployment, see DEPLOYMENT.md"
